package cmd

import (
	"fmt"

	"github.com/MiravaOrg/mirava-core/pkg"
	"github.com/spf13/cobra"
)

func NewPyPICmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "PyPI",
		Short: "PyPI registry tools",
		Long:  "Commands for checking PyPI registry status and packages",
	}

	cmd.AddCommand(NewPyPIStatusCmd(m))
	cmd.AddCommand(NewPyPISpeedCmd(m))
	cmd.AddCommand(NewPyPIPackageCmd(m))

	return cmd
}

func NewPyPIStatusCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check PyPI registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			url := args[0]

			status, data, err := m.PyPi.CheckStatus(url, verbose)
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
			fmt.Printf("Test Path: %s\n", data.TestedPath)
			fmt.Printf("Status Code: %d\n", data.StatusCode)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

func NewPyPIPackageCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in PyPI registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			url := args[0]
			packageName := args[1]

			exists, data, err := m.PyPi.CheckPackage(url, packageName, verbose)
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

	return cmd
}

func NewPyPISpeedCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from NPM registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := args[0]

			if timeout == 0 {
				timeout = 10
			}

			speed, data, err := m.PyPi.CheckSpeed(url, timeout, verbose)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Printf("Speed: %.2f MB/s\n", speed)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Timeout: %d seconds\n", data.TimeoutSec)
			fmt.Printf("Download Size: %.2f MB\n", data.DownloadedMB)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringP("package", "p", "", "Package name to test download speed (required)")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the download test")

	return cmd
}
