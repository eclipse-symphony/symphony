/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/cli/config"
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
		//check context
		c := config.GetMaestroConfig(configFile)
		ctx := c.DefaultContext
		if configContext != "" {
			ctx = configContext
		}

		if ctx == "" {
			ctx = "default"
		}
		//check target
		targets, err := utils.Get(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "targets", "", "", target)
		if err != nil && targets != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(targets) == 0 {
			fmt.Printf("\n%s  Target '%s' is not found%s\n\n", utils.ColorRed(), target, utils.ColorReset())
			return
		}
		reader := margo.MargoSolutionReader{}
		solution, err := reader.Parse(hydra.AppPackageDescription{
			Path:    appPackagePath,
			Type:    "margo",
			Version: "v0",
		})
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		solutionName := solution.ObjectMeta.Name
		solutions, err := utils.Get(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "solutions", "", "", solutionName)
		if err != nil && solutions != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(solutions) != 0 {
			fmt.Printf("\n%s  Solution '%s' already exists%s\n\n", utils.ColorRed(), solutionName, utils.ColorReset())
			return
		}
		//create solution
		solutionData, err := json.Marshal(solution)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		err = utils.Upsert(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "solutions", solutionName, solutionData)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		instanceName := solutionName + "-instance"
		//check instance
		instances, err := utils.Get(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "instances", "", "", instanceName)
		if err != nil && instances != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if len(instances) > 0 {
			fmt.Printf("\n%s  Solution instance '%s' already exists%s\n\n", utils.ColorRed(), instanceName, utils.ColorReset())
			return
		}
		//create instance
		instance := Instance{
			Metadata: model.ObjectMeta{
				Name: instanceName,
			},
			Spec: model.InstanceSpec{
				Solution: solutionName,
				Target: model.TargetSelector{
					Name: target,
				},
			},
		}
		instanceData, err := json.Marshal(instance)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		err = utils.Upsert(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "instances", instanceName, instanceData)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		fmt.Printf("\n%s  Solution instance created successfully%s\n\n", utils.ColorGreen(), utils.ColorReset())
	},
}

func init() {
	MargoRunCmd.Flags().StringVarP(&appPackagePath, "package", "p", "", "Margo application package definition path")
	MargoRunCmd.Flags().StringVarP(&target, "target", "t", "", "Target to run the solution on")
	MargoRunCmd.MarkFlagRequired("package")
	MargoRunCmd.MarkFlagRequired("target")
	MargoCmd.AddCommand(MargoRunCmd)
	RootCmd.AddCommand(MargoCmd)
}
