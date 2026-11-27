package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

// NewSetCmd builds the `set` command group: one command per registry that
// switches the local system over to the given mirror. Every subcommand
// health-checks the mirror first (via the core status probe) and aborts on
// failure unless --force is passed, so a typo can't break the machine's
// package tooling.
func NewSetCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a mirror as the active registry on this system",
		Long: `Set a mirror as the active registry on this system with a single command.

  mirava-cli set npm <url>      set npm registry
  mirava-cli set pypi <url>     set pip index-url
  mirava-cli set ubuntu <url>   rewrite apt sources (Ubuntu)
  mirava-cli set debian <url>   rewrite apt sources (Debian)
  mirava-cli set docker <url>   set docker registry mirror (daemon.json)
  mirava-cli set pacman <url>   set pacman Server in mirrorlist
  mirava-cli set go <url>       set GOPROXY
  mirava-cli set cargo <url>    set cargo registry source replacement
  mirava-cli set composer <url> set composer packagist repo

Each target verifies the mirror is reachable before changing anything;
pass --force to skip that check.`,
	}

	cmd.AddCommand(newSetNpmCmd(m))
	cmd.AddCommand(newSetPyPICmd(m))
	cmd.AddCommand(newSetAptCmd("ubuntu", m.AptUbuntu))
	cmd.AddCommand(newSetAptCmd("debian", m.AptDebian))
	cmd.AddCommand(newSetDockerCmd(m))
	cmd.AddCommand(newSetPacmanCmd(m))
	cmd.AddCommand(newSetGoCmd(m))
	cmd.AddCommand(newSetCargoCmd(m))
	cmd.AddCommand(newSetComposerCmd(m))

	return cmd
}

// setTarget describes one registry backend for the set command.
type setTarget struct {
	prober mirava.Prober
	apply  func(url string) error
}

func newSetCmdFor(m *mirava.Mirava, use, short string, t setTarget) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use + " [url]",
		Short: short,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := normalizeMirrorURL(args[0])
			force, _ := cmd.Flags().GetBool("force")

			if !force {
				fmt.Printf("Checking mirror %s ...\n", url)
				res, err := t.prober.Status(cmd.Context(), url)
				if err != nil {
					fail("mirror check failed: %v; mirror was NOT set, use --force to set it anyway", err)
				}
				if !res.OK {
					fail("mirror check failed (status %d, path %s); mirror was NOT set, use --force to set it anyway", res.StatusCode, res.WorkingPath)
				}
				fmt.Printf("Mirror is up (status %d, path %s)\n", res.StatusCode, res.WorkingPath)
			}

			if err := t.apply(url); err != nil {
				fail("%v", err)
			}
			fmt.Printf("✅ %s mirror set to %s\n", use, url)
		},
	}

	cmd.Flags().Bool("force", false, "Skip the reachability check before setting")

	return cmd
}

func newSetNpmCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "npm", "Set the npm registry to a mirror", setTarget{
		prober: m.Npm,
		apply: func(url string) error {
			return runConfig("npm", []string{"config", "set", "registry", url})
		},
	})
}

func newSetPyPICmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "pypi", "Set the pip index-url to a mirror", setTarget{
		prober: m.PyPi,
		apply: func(url string) error {
			// pip config writes to the user's pip.ini/pip.conf; try pip
			// first, fall back to pip3 (Debian/Ubuntu often only ships
			// the versioned name).
			if err := runConfig("pip", []string{"config", "set", "global.index-url", url}); err == nil {
				return nil
			}
			return runConfig("pip3", []string{"config", "set", "global.index-url", url})
		},
	})
}

// runConfig shells out to the native tool so the change lands exactly where
// the tool itself expects it (npmrc, pip.conf), including any tool-specific
// keyring or config-layering behavior we don't want to reimplement.
func runConfig(name string, args []string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found on PATH; install it or set the mirror manually", name)
	}
	out, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %v: %s", name, args[0], err, strings.TrimSpace(string(out)))
	}
	return nil
}

// newSetDockerCmd configures the Docker daemon's registry-mirrors list.
// Docker mirrors only accelerate pulls from Docker Hub; the daemon must be
// reloaded for the change to take effect.
func newSetDockerCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "docker", "Set a Docker Hub registry mirror in daemon.json", setTarget{
		prober: m.Docker,
		apply:  setDockerMirror,
	})
}

// setDockerMirror adds the mirror to /etc/docker/daemon.json's
// "registry-mirrors" array, preserving other keys, and restarts the daemon.
// It needs root.
func setDockerMirror(url string) error {
	const daemonPath = "/etc/docker/daemon.json"

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("docker daemon config can only be set on Linux or macOS")
	}
	if runtime.GOOS == "linux" && os.Geteuid() != 0 {
		return fmt.Errorf("writing %s requires root; re-run with sudo", daemonPath)
	}

	cfg := map[string]any{}
	if data, err := os.ReadFile(daemonPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parsing %s: %w (fix it manually before setting a mirror)", daemonPath, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	mirrors := []string{}
	if raw, ok := cfg["registry-mirrors"].([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				mirrors = append(mirrors, s)
			}
		}
	}
	for _, existing := range mirrors {
		if strings.TrimSuffix(existing, "/") == strings.TrimSuffix(url, "/") {
			fmt.Printf("Mirror %s is already configured\n", url)
			return nil
		}
	}
	mirrors = append(mirrors, url)
	cfg["registry-mirrors"] = mirrors

	if err := os.MkdirAll(filepath.Dir(daemonPath), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(daemonPath, append(out, '\n'), 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s (registry-mirrors: %v)\n", daemonPath, mirrors)
	fmt.Println("Reloading docker daemon ...")
	if err := runConfig("systemctl", []string{"reload", "docker"}); err != nil {
		fmt.Printf("Note: could not reload docker automatically (%v); run 'sudo systemctl reload docker' to apply\n", err)
	}
	return nil
}

// newSetPacmanCmd sets the first Server line in /etc/pacman.d/mirrorlist.
// Arch only: the command refuses to run when /etc/pacman.d does not exist.
func newSetPacmanCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "pacman", "Set the pacman mirrorlist Server", setTarget{
		prober: m.Pacman,
		apply:  setPacmanMirror,
	})
}

func setPacmanMirror(url string) error {
	const mirrorlist = "/etc/pacman.d/mirrorlist"

	if runtime.GOOS != "linux" {
		return fmt.Errorf("pacman can only be set on Linux")
	}
	if _, err := os.Stat("/etc/pacman.d"); err != nil {
		return fmt.Errorf("this system does not look like Arch (no /etc/pacman.d): refusing to write %s", mirrorlist)
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("writing %s requires root; re-run with sudo", mirrorlist)
	}

	if _, err := os.Stat(mirrorlist); err == nil {
		if err := copyFile(mirrorlist, mirrorlist+".mirava-bak"); err != nil {
			return fmt.Errorf("backing up %s: %w", mirrorlist, err)
		}
		fmt.Printf("Backed up original to %s.mirava-bak\n", mirrorlist)
	}

	entry := fmt.Sprintf("Server = %s/$repo/os/$arch\n", strings.TrimSuffix(url, "/"))
	if err := os.WriteFile(mirrorlist, []byte(entry), 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s\n", strings.TrimSpace(entry))
	return runConfig("pacman", []string{"-Sy"})
}

// newSetGoCmd sets GOPROXY for the Go toolchain.
func newSetGoCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "go", "Set GOPROXY to a mirror", setTarget{
		prober: m.Go,
		apply: func(url string) error {
			return runConfig("go", []string{"env", "-w", "GOPROXY=" + strings.TrimSuffix(url, "/")})
		},
	})
}

// newSetCargoCmd replaces the crates.io source with the mirror in
// ~/.cargo/config.toml, the standard sparse-registry mirror setup.
func newSetCargoCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "cargo", "Set the cargo registry source to a mirror", setTarget{
		prober: m.Cargo,
		apply:  setCargoMirror,
	})
}

func setCargoMirror(url string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(home, ".cargo", "config.toml")

	cfg := ""
	if data, err := os.ReadFile(cfgPath); err == nil {
		cfg = string(data)
		if strings.Contains(cfg, "[source.crates-io]") {
			// Replace the existing replace-with line rather than appending
			// a second one, which cargo would reject.
			re := regexp.MustCompile(`(?m)^\s*replace-with\s*=.*$`)
			cfg = re.ReplaceAllString(cfg, "replace-with = \"mirava\"")
		} else {
			cfg += "\n[source.crates-io]\nreplace-with = \"mirava\"\n"
		}
	} else {
		cfg = "[source.crates-io]\nreplace-with = \"mirava\"\n"
	}

	mirrorBlock := fmt.Sprintf("[source.mirava]\nregistry = \"sparse+%s\"\n", strings.TrimSuffix(url, "/"))
	if !strings.Contains(cfg, "[source.mirava]") {
		cfg += "\n" + mirrorBlock
	} else {
		// Update the existing mirava source block's registry URL, keeping
		// the sparse+ prefix cargo needs for the sparse protocol.
		registry := "sparse+" + strings.TrimSuffix(url, "/")
		re := regexp.MustCompile(`(?m)(\[source\.mirava\]\s*\nregistry\s*=\s*)"[^"]*"`)
		cfg = re.ReplaceAllString(cfg, "${1}\""+registry+"\"")
		// The replace-with line may still be missing if a user removed it.
		if !strings.Contains(cfg, "replace-with") {
			cfg += "\n[source.crates-io]\nreplace-with = \"mirava\"\n"
		}
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s\n", cfgPath)
	return nil
}

// newSetComposerCmd adds the mirror as a packagist repository via composer
// config, so the native tool manages its own config file.
func newSetComposerCmd(m *mirava.Mirava) *cobra.Command {
	return newSetCmdFor(m, "composer", "Set the composer packagist repository to a mirror", setTarget{
		prober: m.Composer,
		apply: func(url string) error {
			if err := runConfig("composer", []string{"config", "--global", "repositories.packagist.org", "false"}); err != nil {
				return err
			}
			return runConfig("composer", []string{"config", "--global", "repositories.mirava", "{\"type\":\"composer\",\"url\":\"" + strings.TrimSuffix(url, "/") + "\"}"})
		},
	})
}

// newSetAptCmd builds the ubuntu/debian variants. Both rewrite
// /etc/apt/sources.list to point at the mirror; the distro argument only
// selects which core backend verifies the mirror and which suites are
// written as a fallback when /etc/os-release has no codename.
func newSetAptCmd(distro string, prober mirava.Prober) *cobra.Command {
	return newSetCmdFor(nil, distro, "Rewrite /etc/apt/sources.list to a "+distro+" mirror", setTarget{
		prober: prober,
		apply: func(url string) error {
			return setAptSources(url, distro)
		},
	})
}

// setAptSources replaces the system apt sources with a single mirror entry,
// keeping backups of the originals. It needs root; without root it prints
// the exact sudo command instead of failing halfway.
//
// Two formats exist and both are supported:
//   - Debian <= 12 and Ubuntu <= 23.10: classic one-line deb entries in
//     /etc/apt/sources.list.
//   - Debian 13 (trixie) and Ubuntu 24.04+: deb822 stanzas in
//     /etc/apt/sources.list.d/*.sources. Writing the classic format there
//     still works but fights the distro's own .sources file, so we follow the
//     new pattern and disable the stock one instead.
func setAptSources(url, distro string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("apt sources can only be set on Linux")
	}

	// Deny cross-distro use: `set ubuntu` on a Debian machine (or any
	// other distro) must not rewrite sources. ID_LIKE covers derivatives
	// (e.g. linuxmint ID_LIKE "debian ubuntu"), so those are allowed for
	// the matching parent.
	osID := osReleaseField("ID")
	osLike := osReleaseField("ID_LIKE")
	if osID != distro && !strings.Contains(osLike, distro) {
		return fmt.Errorf("this system is %q, not %s: refusing to rewrite apt sources for %s", osID, distro, distro)
	}

	suite := osReleaseField("VERSION_CODENAME")
	if suite == "" {
		// Fallback suites per distro when os-release has no codename
		// (e.g. some containers). Good enough to bootstrap; apt itself
		// will error loudly if the suite does not exist.
		if distro == "debian" {
			suite = "stable"
		} else {
			suite = "noble"
		}
	}

	mirrorURL := strings.TrimSuffix(url, "/") + "/" + distro

	if os.Geteuid() != 0 {
		return fmt.Errorf("writing apt sources requires root; re-run with sudo")
	}

	if useDeb822(distro) {
		return setAptDeb822(mirrorURL, suite, distro)
	}
	return setAptClassic(mirrorURL, suite)
}

// useDeb822 reports whether this release ships the deb822 .sources format
// by default: Debian 13+ (trixie and later) and Ubuntu 24.04+.
func useDeb822(distro string) bool {
	ver := osReleaseField("VERSION_ID")
	majorStr, _, _ := strings.Cut(ver, ".")
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return false // unknown version: stay on the classic, widely-safe path
	}
	if distro == "debian" {
		return major >= 13
	}
	if distro == "ubuntu" {
		return major >= 24
	}
	return false
}

// setAptDeb822 writes a deb822 stanza to /etc/apt/sources.list.d/ and
// disables the stock sources: the classic sources.list and the distro's own
// .sources file, so apt does not see both the mirror and the originals.
func setAptDeb822(mirrorURL, suite, distro string) error {
	keyring := "/usr/share/keyrings/ubuntu-archive-keyring.gpg"
	stockName := "ubuntu.sources"
	if distro == "debian" {
		keyring = "/usr/share/keyrings/debian-archive-keyring.gpg"
		stockName = "debian.sources"
	}

	entry := fmt.Sprintf("Types: deb\nURIs: %s\nSuites: %s\nComponents: main\nSigned-By: %s\n",
		mirrorURL, suite, keyring)

	// Disable the stock .sources file if it exists (Ubuntu 24.04+/Debian 13
	// ship one); renaming keeps it recoverable.
	stockPath := "/etc/apt/sources.list.d/" + stockName
	if _, err := os.Stat(stockPath); err == nil {
		if err := disableFile(stockPath); err != nil {
			return fmt.Errorf("disabling %s: %w", stockPath, err)
		}
		fmt.Printf("Disabled stock %s (renamed to %s.mirava-bak)\n", stockPath, stockPath)
	}
	// Same for a classic sources.list left over from an upgrade.
	if _, err := os.Stat("/etc/apt/sources.list"); err == nil {
		if err := disableFile("/etc/apt/sources.list"); err != nil {
			return fmt.Errorf("disabling /etc/apt/sources.list: %w", err)
		}
		fmt.Println("Disabled stock /etc/apt/sources.list (renamed to .mirava-bak)")
	}

	out := "/etc/apt/sources.list.d/mirava.sources"
	if err := os.WriteFile(out, []byte(entry), 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s:\n%s", out, entry)
	return runConfig("apt-get", []string{"update"})
}

// setAptClassic is the pre-deb822 path: rewrite /etc/apt/sources.list.
func setAptClassic(mirrorURL, suite string) error {
	listPath := "/etc/apt/sources.list"
	entry := fmt.Sprintf("deb %s %s main\n", mirrorURL, suite)

	if _, err := os.Stat(listPath); err == nil {
		if err := copyFile(listPath, listPath+".mirava-bak"); err != nil {
			return fmt.Errorf("backing up %s: %w", listPath, err)
		}
		fmt.Printf("Backed up original to %s.mirava-bak\n", listPath)
	}

	if err := os.WriteFile(listPath, []byte(entry), 0o644); err != nil {
		return err
	}

	fmt.Printf("Wrote %s\n", strings.TrimSpace(entry))
	return runConfig("apt-get", []string{"update"})
}

// disableFile renames a sources file aside so apt stops reading it, keeping
// the original recoverable under a .mirava-bak suffix.
func disableFile(path string) error {
	_, err := os.Stat(path + ".mirava-bak")
	if err == nil {
		return nil // already disabled by a previous run
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(path, path+".mirava-bak")
}

// osReleaseField reads a field from /etc/os-release (e.g. VERSION_CODENAME).
func osReleaseField(field string) string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok && k == field {
			return strings.Trim(v, `"`)
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
