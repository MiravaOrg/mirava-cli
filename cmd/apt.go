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

	cmd.AddCommand(NewAptStatusCommand(m))
	cmd.AddCommand(NewAptPackageCmd(m))
	cmd.AddCommand(NewAptSpeedCmd(m))

	return cmd
}

func NewAptStatusCommand(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check apt registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := args[0]

			res, err := m.Apt.Status(ctx, url)
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

func NewAptPackageCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in apt registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := args[0]
			packageName := args[1]
			arch, _ := cmd.Flags().GetString("arch")
			component, _ := cmd.Flags().GetString("component")
			release, _ := cmd.Flags().GetString("release")

			res, err := m.Apt.Package(ctx, url, mirava.PackageQuery{
				Name: packageName,
				Options: &mirava.AptQueryOptions{
					Suite:     release,
					Component: component,
					Arch:      arch,
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

	// Empty defaults let mirava-core pick what the mirror serves (amd64,
	// latest codename from Release metadata) instead of hardcoding one.
	cmd.Flags().StringP("arch", "a", "", "Package architecture (e.g. amd64, arm64; default: any served by the mirror)")
	cmd.Flags().StringP("component", "c", "", "Repository component (e.g. main, universe; default: any)")
	cmd.Flags().StringP("release", "r", "", "Release or codename (e.g. noble, bookworm; default: latest)")

	return cmd
}

func NewAptSpeedCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from apt registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := args[0]

			res, err := m.Apt.Speed(ctx, url, mirava.SpeedOptions{
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
