package cmd

import (
	"fmt"
	"time"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

func NewAptCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apt",
		Short: "apt registry tools",
		Long:  "Commands for checking apt registry status and packages",
	}

	cmd.AddCommand(NewAptStatusCommand(m.Apt))
	cmd.AddCommand(NewAptPackageCmd(m.Apt))
	cmd.AddCommand(NewAptSpeedCmd(m.Apt))

	return cmd
}

func NewUbuntuCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ubuntu",
		Short: "Ubuntu apt registry tools",
		Long:  "Commands for checking Ubuntu apt registry status and packages (release probing prefers Ubuntu suites)",
	}

	cmd.AddCommand(NewAptStatusCommand(m.AptUbuntu))
	cmd.AddCommand(NewAptPackageCmd(m.AptUbuntu))
	cmd.AddCommand(NewAptSpeedCmd(m.AptUbuntu))

	return cmd
}

func NewDebianCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debian",
		Short: "Debian apt registry tools",
		Long:  "Commands for checking Debian apt registry status and packages (release probing prefers Debian suites)",
	}

	cmd.AddCommand(NewAptStatusCommand(m.AptDebian))
	cmd.AddCommand(NewAptPackageCmd(m.AptDebian))
	cmd.AddCommand(NewAptSpeedCmd(m.AptDebian))

	return cmd
}

// NewAptStatusCommand builds the status subcommand for a specific apt
// backend (generic, Ubuntu or Debian flavored).
func NewAptStatusCommand(p mirava.Prober) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check apt registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := normalizeMirrorURL(args[0])

			res, err := p.Status(ctx, url)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			statusString := "❌ FAILED"
			if res.OK {
				statusString = "✅ OK"
			}

			fmt.Printf("Status: %s\n", statusString)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Test Path: %s\n", res.WorkingPath)
			fmt.Printf("Status Code: %d\n", res.StatusCode)
		},
	}

	return cmd
}

// NewAptPackageCmd builds the package subcommand. When the user omits
// release/component the backend discovers suites from the mirror's own
// Release metadata; arch defaults to amd64 (what mirrors serve).
func NewAptPackageCmd(p mirava.Prober) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in apt registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := normalizeMirrorURL(args[0])
			packageName := args[1]
			arch, _ := cmd.Flags().GetString("arch")
			component, _ := cmd.Flags().GetString("component")
			release, _ := cmd.Flags().GetString("release")

			res, err := p.Package(ctx, url, mirava.PackageQuery{
				Name: packageName,
				Options: &mirava.AptQueryOptions{
					Suite:     release,   // empty = discover from Release metadata (latest)
					Component: component, // empty = any, main first
					Arch:      arch,      // empty = amd64
				},
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			statusString := "❌ NOT FOUND"
			if res.Found {
				statusString = "✅ EXISTS"
			}

			fmt.Printf("Status: %s\n", statusString)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Version: %s\n", res.Version)
		},
	}

	// Empty defaults: the backend discovers suite/component from the
	// mirror's InRelease/Release metadata and defaults to amd64.
	cmd.Flags().StringP("arch", "a", "", "Package architecture (default: amd64)")
	cmd.Flags().StringP("component", "c", "", "Repository component, e.g. main, universe (default: discovered)")
	cmd.Flags().StringP("release", "r", "", "Release or codename, e.g. noble, bookworm (default: discovered)")

	return cmd
}

// NewAptSpeedCmd builds the speed subcommand for a specific apt backend.
func NewAptSpeedCmd(p mirava.Prober) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from apt registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := normalizeMirrorURL(args[0])

			res, err := p.Speed(ctx, url, mirava.SpeedOptions{
				Duration: time.Duration(timeout) * time.Second,
				Package:  packageName,
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Printf("Speed: %.2f MB/s\n", res.SpeedMBps)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Timeout: %v seconds\n", res.DurationSec)
			fmt.Printf("Download Size: %.2f MB\n", res.DownloadedMB)
		},
	}

	cmd.Flags().StringP("package", "p", "", "Package name to test download speed")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the download test")

	return cmd
}
