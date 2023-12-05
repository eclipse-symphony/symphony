/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"

	"github.com/azure/symphony/cli/utils"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n%s  Maestro Version: %s%s\n\n", utils.ColorPurple(), SymphonyAPIVersion, utils.ColorReset())

	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
