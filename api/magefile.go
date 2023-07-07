//go:build mage

package main

import (
	//mage:import
	_ "dev.azure.com/msazure/One/_git/symphony.git/packages/mage"
	"github.com/princjef/mageutil/shellcmd"
)

func Build() error {
	return shellcmd.Command("go build -o bin/symphony-api -tags=azure").Run()
}
