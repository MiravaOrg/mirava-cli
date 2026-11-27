package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

// NewGetCmd builds the `get` command group: show the mirror currently
// configured on this system for each technology. Every subcommand is
// read-only.
func NewGetCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show the mirror currently set for a technology",
		Long: `Show the mirror currently configured on this system.

  mirava-cli get npm       show npm registry
  mirava-cli get pypi      show pip index-url
  mirava-cli get ubuntu    show apt sources (Ubuntu)
  mirava-cli get debian    show apt sources (Debian)
  mirava-cli get docker    show docker registry mirrors
  mirava-cli get pacman    show pacman mirrorlist
  mirava-cli get go        show GOPROXY
  mirava-cli get cargo     show cargo registry source
  mirava-cli get composer  show composer repositories`,
	}

	cmd.AddCommand(newGetNpmCmd(m))
	cmd.AddCommand(newGetPyPICmd(m))
	cmd.AddCommand(newGetAptCmd("ubuntu", m.AptUbuntu))
	cmd.AddCommand(newGetAptCmd("debian", m.AptDebian))
	cmd.AddCommand(newGetDockerCmd(m))
	cmd.AddCommand(newGetPacmanCmd(m))
	cmd.AddCommand(newGetGoCmd(m))
	cmd.AddCommand(newGetCargoCmd(m))
	cmd.AddCommand(newGetComposerCmd(m))

	return cmd
}

// NewResetCmd builds the `reset` command group: restore each technology's
// default upstream registry, undoing what `set` did.
func NewResetCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset a technology back to its default upstream registry",
		Long: `Reset a technology back to its default upstream registry.

  mirava-cli reset npm       reset npm registry to registry.npmjs.org
  mirava-cli reset pypi      reset pip index-url to pypi.org
  mirava-cli reset ubuntu    restore apt sources from backup (Ubuntu)
  mirava-cli reset debian    restore apt sources from backup (Debian)
  mirava-cli reset docker    remove docker registry mirrors
  mirava-cli reset pacman    restore pacman mirrorlist from backup
  mirava-cli reset go        reset GOPROXY to the default
  mirava-cli reset cargo     remove the cargo mirror source
  mirava-cli reset composer  remove the composer mirror repo`,
	}

	cmd.AddCommand(newResetNpmCmd(m))
	cmd.AddCommand(newResetPyPICmd(m))
	cmd.AddCommand(newResetAptCmd("ubuntu", m.AptUbuntu))
	cmd.AddCommand(newResetAptCmd("debian", m.AptDebian))
	cmd.AddCommand(newResetDockerCmd(m))
	cmd.AddCommand(newResetPacmanCmd(m))
	cmd.AddCommand(newResetGoCmd(m))
	cmd.AddCommand(newResetCargoCmd(m))
	cmd.AddCommand(newResetComposerCmd(m))

	return cmd
}

// runConfigOutput shells out to a native tool and returns its stdout, for
// read-only queries (get). It reports a friendly error when the tool is
// missing so `get` degrades gracefully instead of crashing.
func runConfigOutput(name string, args ...string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found on PATH", name)
	}
	out, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %v: %s", name, args[0], err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// --- npm ---

func newGetNpmCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "npm",
		Short: "Show the npm registry",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := runConfigOutput("npm", "config", "get", "registry")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Registry: %s\n", out)
		},
	}
}

func newResetNpmCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "npm",
		Short: "Reset the npm registry to registry.npmjs.org",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfig("npm", []string{"config", "delete", "registry"}); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("✅ npm registry reset to default (registry.npmjs.org)")
		},
	}
}

// --- pypi ---

func newGetPyPICmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "pypi",
		Short: "Show the pip index-url",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := runConfigOutput("pip", "config", "get", "global.index-url")
			if err != nil {
				if out2, err2 := runConfigOutput("pip3", "config", "get", "global.index-url"); err2 == nil {
					fmt.Printf("Index URL: %s\n", out2)
					return
				}
				fmt.Println("Index URL: (not set, using default https://pypi.org/simple)")
				return
			}
			fmt.Printf("Index URL: %s\n", out)
		},
	}
}

func newResetPyPICmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "pypi",
		Short: "Reset the pip index-url to pypi.org",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfig("pip", []string{"config", "unset", "global.index-url"}); err != nil {
				if err2 := runConfig("pip3", []string{"config", "unset", "global.index-url"}); err2 != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
			}
			fmt.Println("✅ pip index-url reset to default (pypi.org)")
		},
	}
}

// --- docker ---

func newGetDockerCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "docker",
		Short: "Show the docker registry mirrors",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			mirrors, err := readDockerMirrors()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if len(mirrors) == 0 {
				fmt.Println("Registry mirrors: none configured")
				return
			}
			fmt.Println("Registry mirrors:")
			for _, u := range mirrors {
				fmt.Printf("  %s\n", u)
			}
		},
	}
}

func newResetDockerCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "docker",
		Short: "Remove all docker registry mirrors",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := resetDockerMirrors(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("✅ docker registry mirrors removed")
		},
	}
}

// readDockerMirrors reads the registry-mirrors list from daemon.json.
func readDockerMirrors() ([]string, error) {
	const daemonPath = "/etc/docker/daemon.json"
	data, err := os.ReadFile(daemonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg struct {
		RegistryMirrors []string `json:"registry-mirrors"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", daemonPath, err)
	}
	return cfg.RegistryMirrors, nil
}

// resetDockerMirrors removes the registry-mirrors key from daemon.json and
// reloads the daemon. It needs root on Linux.
func resetDockerMirrors() error {
	const daemonPath = "/etc/docker/daemon.json"

	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("docker daemon config can only be reset on Linux or macOS")
	}
	if runtime.GOOS == "linux" && os.Geteuid() != 0 {
		return fmt.Errorf("writing %s requires root; re-run with sudo", daemonPath)
	}

	data, err := os.ReadFile(daemonPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No docker daemon config found; nothing to reset")
			return nil
		}
		return err
	}

	cfg := map[string]any{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing %s: %w (fix it manually before resetting)", daemonPath, err)
	}
	if _, ok := cfg["registry-mirrors"]; !ok {
		fmt.Println("No registry-mirrors configured; nothing to reset")
		return nil
	}
	delete(cfg, "registry-mirrors")

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(daemonPath, append(out, '\n'), 0o644); err != nil {
		return err
	}

	fmt.Println("Reloading docker daemon ...")
	if err := runConfig("systemctl", []string{"reload", "docker"}); err != nil {
		fmt.Printf("Note: could not reload docker automatically (%v); run 'sudo systemctl reload docker' to apply\n", err)
	}
	return nil
}

// --- pacman ---

func newGetPacmanCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "pacman",
		Short: "Show the pacman mirrorlist servers",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile("/etc/pacman.d/mirrorlist")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("Servers (uncommented):")
			found := false
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Server =") {
					fmt.Printf("  %s\n", line)
					found = true
				}
			}
			if !found {
				fmt.Println("  (none)")
			}
		},
	}
}

func newResetPacmanCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "pacman",
		Short: "Restore the pacman mirrorlist from backup",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := restoreFromBackup("/etc/pacman.d/mirrorlist"); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("✅ pacman mirrorlist restored")
		},
	}
}

// --- go ---

func newGetGoCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "go",
		Short: "Show GOPROXY",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := runConfigOutput("go", "env", "GOPROXY")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("GOPROXY: %s\n", out)
		},
	}
}

func newResetGoCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "go",
		Short: "Reset GOPROXY to the Go default",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfig("go", []string{"env", "-w", "GOPROXY=https://proxy.golang.org,direct"}); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("✅ GOPROXY reset to default (https://proxy.golang.org,direct)")
		},
	}
}

// --- cargo ---

func newGetCargoCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "cargo",
		Short: "Show the cargo registry source",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			data, err := os.ReadFile(filepath.Join(home, ".cargo", "config.toml"))
			if err != nil {
				fmt.Println("No cargo config found; using default crates.io")
				return
			}
			fmt.Printf("Config (%s):\n%s", filepath.Join(home, ".cargo", "config.toml"), string(data))
		},
	}
}

func newResetCargoCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "cargo",
		Short: "Remove the cargo mirror source",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := resetCargoMirror(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("✅ cargo reset to default crates.io")
		},
	}
}

// resetCargoMirror removes the [source.mirava] block and the replace-with
// line from ~/.cargo/config.toml, leaving any other user config intact.
func resetCargoMirror() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(home, ".cargo", "config.toml")

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No cargo config found; nothing to reset")
			return nil
		}
		return err
	}

	cfg := string(data)
	// Drop the mirava source block: from its header to the next section
	// header or end of file.
	if idx := strings.Index(cfg, "[source.mirava]"); idx >= 0 {
		end := len(cfg)
		if next := strings.Index(cfg[idx+1:], "\n["); next >= 0 {
			end = idx + 1 + next
		}
		cfg = cfg[:idx] + cfg[end:]
	}
	// Drop the replace-with line pointing at mirava.
	cfg = strings.ReplaceAll(cfg, "replace-with = \"mirava\"\n", "")

	// Drop a [source.crates-io] section left with no keys inside: the
	// header plus any blank lines up to the next section or EOF.
	if idx := strings.Index(cfg, "[source.crates-io]"); idx >= 0 {
		rest := cfg[idx+len("[source.crates-io]"):]
		trimmed := strings.TrimLeft(rest, " \t\n")
		if strings.HasPrefix(trimmed, "[") || trimmed == "" {
			cfg = cfg[:idx] + trimmed
		}
	}

	cfg = strings.TrimSpace(cfg)
	if cfg == "" {
		if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Printf("Removed empty %s\n", cfgPath)
		return nil
	}
	if err := os.WriteFile(cfgPath, []byte(cfg+"\n"), 0o644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s\n", cfgPath)
	return nil
}

// --- composer ---

func newGetComposerCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "composer",
		Short: "Show the composer repositories",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := runConfigOutput("composer", "config", "--global", "repositories")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("Repositories:\n%s\n", out)
		},
	}
}

func newResetComposerCmd(m *mirava.Mirava) *cobra.Command {
	return &cobra.Command{
		Use:   "composer",
		Short: "Remove the composer mirror repo and re-enable packagist",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfig("composer", []string{"config", "--global", "--unset", "repositories.mirava"}); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			// Re-enable the official packagist that `set` disabled.
			if err := runConfig("composer", []string{"config", "--global", "--unset", "repositories.packagist.org"}); err != nil {
				fmt.Printf("Note: could not re-enable packagist (%v); check 'composer config --global repositories'\n", err)
				return
			}
			fmt.Println("✅ composer reset to default packagist")
		},
	}
}

// --- apt (ubuntu/debian) ---

func newGetAptCmd(distro string, prober mirava.Prober) *cobra.Command {
	return &cobra.Command{
		Use:   distro,
		Short: "Show the apt sources for " + distro,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				fmt.Println("Error: apt sources can only be read on Linux")
				return
			}
			fmt.Println("Classic /etc/apt/sources.list:")
			if data, err := os.ReadFile("/etc/apt/sources.list"); err == nil {
				fmt.Println(strings.TrimRight(string(data), "\n"))
			} else {
				fmt.Println("  (not present)")
			}
			entries, err := filepath.Glob("/etc/apt/sources.list.d/*")
			if err == nil && len(entries) > 0 {
				fmt.Println("\n/etc/apt/sources.list.d/:")
				for _, p := range entries {
					fmt.Printf("  %s\n", p)
				}
			}
		},
	}
}

func newResetAptCmd(distro string, prober mirava.Prober) *cobra.Command {
	return &cobra.Command{
		Use:   distro,
		Short: "Restore apt sources from backup for " + distro,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := resetAptSources(distro); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("✅ apt sources restored for %s\n", distro)
		},
	}
}

// resetAptSources undoes what setAptSources did: remove mirava.sources,
// restore the stock .sources/.list files from their .mirava-bak copies.
func resetAptSources(distro string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("apt sources can only be reset on Linux")
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("resetting apt sources requires root; re-run with sudo")
	}

	stockName := "ubuntu.sources"
	if distro == "debian" {
		stockName = "debian.sources"
	}

	// Remove the mirava-written deb822 file if present.
	miravaPath := "/etc/apt/sources.list.d/mirava.sources"
	if _, err := os.Stat(miravaPath); err == nil {
		if err := os.Remove(miravaPath); err != nil {
			return fmt.Errorf("removing %s: %w", miravaPath, err)
		}
		fmt.Printf("Removed %s\n", miravaPath)
	}

	// Restore any backups we made.
	for _, path := range []string{
		"/etc/apt/sources.list.d/" + stockName,
		"/etc/apt/sources.list",
	} {
		if err := restoreFromBackup(path); err != nil {
			return err
		}
	}

	return runConfig("apt-get", []string{"update"})
}

// restoreFromBackup restores path from path+".mirava-bak" if the backup
// exists, and is a no-op (with a note) when it does not.
func restoreFromBackup(path string) error {
	backup := path + ".mirava-bak"
	if _, err := os.Stat(backup); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("%s has no mirava backup; leaving it untouched\n", path)
			}
			return nil
		}
		return err
	}
	if err := copyFile(backup, path); err != nil {
		return fmt.Errorf("restoring %s: %w", path, err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("cleaning up %s: %w", backup, err)
	}
	fmt.Printf("Restored %s from backup\n", path)
	return nil
}
