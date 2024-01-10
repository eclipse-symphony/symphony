/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"

	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
)

var (
	cVerbose bool
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check Symphony prerequisits",
	Run: func(cmd *cobra.Command, args []string) {
		b := utils.CheckDocker(cVerbose)
		b = utils.CheckKubectl(cVerbose) && b
		_, r := utils.CheckK8sConnection(cVerbose)
		b = r && b
		b = utils.CheckHelm(verbose) && b
		if b {
			fmt.Printf("\n%s  All Prerequisites are met!%s\n\n", utils.ColorCyan(), utils.ColorReset())
		}
	},
}

func init() {
	CheckCmd.Flags().BoolVar(&cVerbose, "verbose", false, "Detailed outputs")
	RootCmd.AddCommand(CheckCmd)
}
