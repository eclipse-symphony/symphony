/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage maestro CLI configuration",
}

var ConfigUseContextCmd = &cobra.Command{
	Use:   "use-context <context-name>",
	Short: "Set the default context in ~/.symphony/.config.json",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctxName := args[0]
		c := config.GetMaestroConfig("")
		if _, ok := c.Contexts[ctxName]; !ok {
			available := make([]string, 0, len(c.Contexts))
			for k := range c.Contexts {
				available = append(available, k)
			}
			sort.Strings(available)
			fmt.Printf("\n%s  context '%s' not found. Available contexts: %v%s\n\n",
				utils.ColorRed(), ctxName, available, utils.ColorReset())
			os.Exit(1)
		}
		c.DefaultContext = ctxName
		if err := config.SaveMaestroConfig(c); err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			os.Exit(1)
		}
		fmt.Printf("\n%s  Switched to context '%s'.%s\n\n",
			utils.ColorCyan(), ctxName, utils.ColorReset())
	},
}

func init() {
	ConfigCmd.AddCommand(ConfigUseContextCmd)
	ConfigCmd.AddCommand(ConfigGetContextsCmd)
	ConfigCmd.AddCommand(ConfigDeleteContextCmd)
	RootCmd.AddCommand(ConfigCmd)
}

var ConfigDeleteContextCmd = &cobra.Command{
	Use:   "delete-context <context-name>",
	Short: "Delete a context from ~/.symphony/.config.json",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctxName := args[0]
		c := config.GetMaestroConfig("")
		if _, ok := c.Contexts[ctxName]; !ok {
			available := make([]string, 0, len(c.Contexts))
			for k := range c.Contexts {
				available = append(available, k)
			}
			sort.Strings(available)
			fmt.Printf("\n%s  context '%s' not found. Available contexts: %v%s\n\n",
				utils.ColorRed(), ctxName, available, utils.ColorReset())
			os.Exit(1)
		}
		delete(c.Contexts, ctxName)
		if c.DefaultContext == ctxName {
			c.DefaultContext = ""
		}
		if err := config.SaveMaestroConfig(c); err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			os.Exit(1)
		}
		fmt.Printf("\n%s  Deleted context '%s'.%s\n\n",
			utils.ColorCyan(), ctxName, utils.ColorReset())
	},
}

var ConfigGetContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List maestro CLI configuration contexts",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.GetMaestroConfig("")
		names := make([]string, 0, len(c.Contexts))
		for k := range c.Contexts {
			names = append(names, k)
		}
		sort.Strings(names)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Current", "Name", "URL"})
		for _, name := range names {
			marker := ""
			if name == c.DefaultContext {
				marker = "*"
			}
			t.AppendRow(table.Row{marker, name, c.Contexts[name].Url})
		}
		t.SetStyle(table.StyleColoredBright)
		t.Render()
	},
}
