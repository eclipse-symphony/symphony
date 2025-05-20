/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"github.com/eclipse-symphony/symphony/cli/cmd"
)

// Version value is injected by the build.
var (
	version = "0.40.60"
)

func main() {
	cmd.Execute(version)
}
