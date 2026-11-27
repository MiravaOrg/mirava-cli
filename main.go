package main

import (
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
		panic(err)
	}
}
