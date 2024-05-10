/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"

	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/eclipse-symphony/symphony/hydra"
	"github.com/eclipse-symphony/symphony/hydra/margo/v0"
	"github.com/spf13/cobra"
)

var (
	appPackagePath string
)

var MargoCmd = &cobra.Command{
	Use:   "margo",
	Short: "Margo commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n%sPlease use either 'margo run' or 'margo list'%s\n\n", utils.ColorRed(), utils.ColorReset())
	},
}

var MargoRunCmd = &cobra.Command{
	Use:   "run",
	Short: "run a Margo application package",
	Run: func(cmd *cobra.Command, args []string) {
		reader := margo.MargoSolutionReader{}
		reader.Parse(hydra.AppPackageDescription{
			Path:    appPackagePath,
			Type:    "margo",
			Version: "v1",
		})

	},
}

func init() {
	MargoRunCmd.Flags().StringVarP(&appPackagePath, "package", "p", "", "Margo application package definition path")
	MargoCmd.AddCommand(MargoRunCmd)
	RootCmd.AddCommand(MargoCmd)
}
