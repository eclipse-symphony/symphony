/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"github.com/azure/symphony/api/cmd"
)

// Version value is injected by the build.
var (
	version = ""
)

func main() {
	cmd.Execute(version)
}
