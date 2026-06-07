package cmd

import (
	"fmt"

	"github.com/MiravaOrg/mirava-core/pkg"
	"github.com/spf13/cobra"
)

func NewAptCmd(m *pkg.MiravaService) *cobra.Command {
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

func NewAptStatusCommand(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check apt registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			url := args[0]

			status, data, err := m.Apt.CheckStatus(url, verbose)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			var statusString string
			if status {
				statusString = "✅ OK"
			} else {
				statusString = "❌ FAILED"
			}

			fmt.Printf("Status: %s\n", statusString)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Test Path: %s\n", data.WorkingPath)
			fmt.Printf("Status Code: %d\n", data.StatusCode)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

func NewAptPackageCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in apt registry",
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			url := args[0]
			packageName := args[1]
			arch, _ := cmd.Flags().GetString("arch")
			component, _ := cmd.Flags().GetString("component")
			release, _ := cmd.Flags().GetString("release")

			params := pkg.AptCheckPackageParams{
				Arch:      arch,
				Component: component,
				Release:   release,
			}

			exists, data, err := m.Apt.CheckPackage(url, packageName, verbose, params)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			var statusString string
			if exists {
				statusString = "✅ EXISTS"
			} else {
				statusString = "❌ NOT FOUND"
			}

			fmt.Printf("Status: %s\n", statusString)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Version: %s\n", data.Version)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringP("arch", "a", "arm64", "Get OS architecture")
	cmd.Flags().StringP("component", "c", "main", "Get OS component")
	cmd.Flags().StringP("release", "r", "jammy", "Get OS release")

	return cmd
}

func NewAptSpeedCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from apt registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := args[0]

			if timeout == 0 {
				timeout = 10
			}

			speed, data, err := m.Apt.CheckSpeed(url, timeout, verbose)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Printf("Speed: %.2f MB/s\n", speed)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Timeout: %v seconds\n", data.DurationSec)
			fmt.Printf("Download Size: %.2f MB\n", data.DownloadedMB)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringP("package", "p", "", "Package name to test download speed (required)")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the download test")

	return cmd
}
