package main

import (
	"github.com/azure/symphony/cli/cmd"
)

// Version value is injected by the build.
var (
	version = "0.40.60"
)

func main() {
	cmd.Execute(version)
}
