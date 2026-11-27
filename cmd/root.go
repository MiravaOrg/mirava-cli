package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

// normalizeMirrorURL defaults a schemeless mirror argument to https:// so
// `mirava-cli apt package mirror.example.com/ubuntu nginx` behaves the way
// the user typed it. Note url.Parse reads "host.tld/path" as scheme
// "host.tld" with empty host, so the check is on known schemes, not on
// u.Host. URLs that already carry a scheme pass through; other malformed
// input is left untouched for the core to report properly.
func normalizeMirrorURL(raw string) string {
	s := strings.TrimSpace(raw)
	u, err := url.Parse(s)
	if err == nil && u.Scheme != "http" && u.Scheme != "https" {
		return "https://" + s
	}
	return s
}

// fail prints an error to stderr and exits non-zero, so scripts and CI can
// tell failures from successes. All user-facing errors go through this.
func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

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
	cmd.AddCommand(NewSetCmd(m))
	cmd.AddCommand(NewGetCmd(m))
	cmd.AddCommand(NewResetCmd(m))

	addDNSFlag(cmd)

	return cmd
}
