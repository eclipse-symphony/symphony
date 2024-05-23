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
	"github.com/spf13/cobra"
)

var (
	solutionPath string
	target       string
	instanceName string
)

var RunCmd = &cobra.Command{
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
		//check solution
		solutionData, err := utils.GetArtifactFile(solutionPath)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		var solution Solution
		err = json.Unmarshal(solutionData, &solution)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		solutionName := solution.Metadata.Name
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
		err = utils.Upsert(c.Contexts[ctx].Url, c.Contexts[ctx].User, c.Contexts[ctx].Secret, "solutions", solutionName, solutionData)
		if err != nil {
			fmt.Printf("\n%s  %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
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
	RunCmd.Flags().StringVarP(&solutionPath, "solution", "s", "", "Symphony solution file path")
	RunCmd.Flags().StringVarP(&target, "target", "t", "", "Target to run the solution on")
	RunCmd.Flags().StringVarP(&instanceName, "name", "n", "", "Name of the solution instance")
	RunCmd.Flags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	RunCmd.Flags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")
	RunCmd.MarkFlagRequired("solution")
	RunCmd.MarkFlagRequired("target")
	RunCmd.MarkFlagRequired("name")
	RootCmd.AddCommand(RunCmd)
}
