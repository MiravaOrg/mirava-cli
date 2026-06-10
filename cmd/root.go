package cmd

import (
	"fmt"

	"github.com/MiravaOrg/mirava-core/pkg"
	"github.com/spf13/cobra"
)

func NewRootCmd(m *pkg.MiravaService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Mirava CLI",
		Short: "Testing Mirrors CLI",
		Long:  "A tool for testing mirrors",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("CLI version: 0.0.1 -- Core version v0.3.3")
		},
	}

	cmd.AddCommand(NewNpmCmd(m))
	cmd.AddCommand(NewPyPICmd(m))
	cmd.AddCommand(NewAptCmd(m))
	cmd.AddCommand(NewDockerCmd(m))
	cmd.AddCommand(NewPacmanCmd(m))
	cmd.AddCommand(NewGoCmd(m))
	cmd.AddCommand(NewCargoCmd(m))
	cmd.AddCommand(NewComposerCmd(m))

	return cmd
}
