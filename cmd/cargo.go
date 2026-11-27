package cmd

import (
	"fmt"
	"time"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

func NewCargoCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cargo",
		Short: "Cargo registry tools",
		Long:  "Commands for checking Cargo registry status and packages",
	}

	cmd.AddCommand(NewCargoStatusCmd(m))
	cmd.AddCommand(NewCargoPackageCmd(m))
	cmd.AddCommand(NewCargoSpeedCmd(m))

	return cmd
}

func NewCargoStatusCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [url]",
		Short: "Check Cargo registry status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := normalizeMirrorURL(args[0])

			res, err := m.Cargo.Status(ctx, url)
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
		},
	}

	return cmd
}

func NewCargoPackageCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package [url] [package-name]",
		Short: "Check if a package exists in Cargo registry",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			url := normalizeMirrorURL(args[0])
			packageName := args[1]

			res, err := m.Cargo.Package(ctx, url, mirava.PackageQuery{Name: packageName})
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

	return cmd
}

func NewCargoSpeedCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speed [url]",
		Short: "Test download speed from Cargo registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			timeout, _ := cmd.Flags().GetInt("timeout")
			packageName, _ := cmd.Flags().GetString("package")
			url := normalizeMirrorURL(args[0])

			res, err := m.Cargo.Speed(ctx, url, mirava.SpeedOptions{
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
			fmt.Printf("Timeout: %.0f seconds\n", res.DurationSec)
			fmt.Printf("Download Size: %.2f MB\n", res.DownloadedMB)
		},
	}

	cmd.Flags().StringP("package", "p", "", "Package name to test download speed")
	cmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for the download test")

	return cmd
}
