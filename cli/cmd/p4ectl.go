/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)
var RootCmd = &cobra.Command{
	Use:   "maestro",
	Short: "maestro",
	Long: `
                           _              
     /\/\   __ _  ___  ___| |_ _ __ ___  
    /    \ / _' |/ _ \/ __| __| '__/ _ \ 
   / /\/\ \ (_| |  __/\__ \ |_| | | (_) |
   \/    \/\__,_|\___||___/\__|_|  \___/ 

   Jumpstart your Microsoft Edge Journey
	`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute(versiong string) {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Display verbose tracing info.")
}
