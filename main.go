package main

import (
	"fmt"
	"os"

	"github.com/MiravaOrg/mirava-core"
	"github.com/miravaOrg/mirava-cli/cmd"
)

var MiravaCore = mirava.New()

func main() {
	rootCmd := cmd.NewRootCmd(MiravaCore)
	rootCmd.SetArgs(cmd.RewriteDNSArgs(os.Args[1:]))

	err := rootCmd.Execute()
	if err != nil {
		// Print the error and exit non-zero; a CLI must never dump a
		// stack trace on its users.
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
