package cmd

import (
	"fmt"

	"github.com/MiravaOrg/mirava-core/pkg"
	"github.com/spf13/cobra"
)

func NewDockerCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker registry tools",
		Long:  "Commands for checking NPM registry status and packages",
	}

	cmd.AddCommand(NewDockerStatusCommand(m))
	cmd.AddCommand(NewDockerPackageCmd(m))
	cmd.AddCommand(NewDockerSpeedCmd(m))

	return cmd
}

func NewDockerStatusCommand(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check Docker registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]

			status, data, err := m.Docker.CheckStatus(url)
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
			fmt.Printf("Test Path: %s\n", data.Url)
			fmt.Printf("Status: %v\n", data.Status)
		},
	}

	return cmd
}

func NewDockerPackageCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in Docker registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			imageName := args[1]

			exists, data, err := m.Docker.CheckPackage(url, imageName)
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
			fmt.Printf("Image: %s\n", data.ImageName)
		},
	}

	return cmd
}

func NewDockerSpeedCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from Docker registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := args[0]

			if timeout == 0 {
				timeout = 10
			}

			speed, data, err := m.Docker.CheckSpeed(url, timeout, verbose, &packageName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Printf("Speed: %.2f MB/s\n", speed)
			fmt.Printf("URL: %s\n", url)
			fmt.Printf("Package: %s\n", packageName)
			fmt.Printf("Timeout: %d seconds\n", data.TimeoutSec)
			fmt.Printf("Download Size: %.2f MB\n", data.DownloadMb)
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringP("package", "p", "", "Package name to test download speed (required)")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the download test")

	return cmd
}
