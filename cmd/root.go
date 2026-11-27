package cmd

import (
	"fmt"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

func NewRootCmd(m *mirava.Mirava) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Mirava CLI",
		Short: "Testing Mirrors CLI",
		Long:  "A tool for testing mirrors",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("CLI version: 0.0.1 -- Core version v%s\n", mirava.Version)
		},
	}

	cmd.AddCommand(NewNpmCmd(m))
	cmd.AddCommand(NewPyPICmd(m))
	cmd.AddCommand(NewAptCmd(m))
	cmd.AddCommand(NewUbuntuCmd(m))
	cmd.AddCommand(NewDebianCmd(m))
	cmd.AddCommand(NewDockerCmd(m))
	cmd.AddCommand(NewPacmanCmd(m))
	cmd.AddCommand(NewGoCmd(m))
	cmd.AddCommand(NewCargoCmd(m))
	cmd.AddCommand(NewComposerCmd(m))

	return cmd
}
