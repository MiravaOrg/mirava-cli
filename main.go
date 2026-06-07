package main

import (
	"github.com/MiravaOrg/mirava-core"
	"github.com/miravaOrg/mirava-cli/cmd"
)

var MiravaCore = mirava.CreateMiravaService()

func main() {
	rootCmd := cmd.NewRootCmd(MiravaCore)
	
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
