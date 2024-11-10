/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
)

var (
	mqttBrokerAddress string
	mqttRequestTopic  string
	mqttResponseTopic string
	mqttClientId      string
	asDaemon          bool
	targetName        string
	targetNamespace   string
	upsertTarget      bool
	configOnly        bool
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install Symphony agent on the machine",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.GetMaestroConfig(configFile)
		ctx := c.DefaultContext
		if configContext != "" {
			ctx = configContext
		}

		if ctx == "" {
			ctx = "default"
		}

		u, err := user.Current()
		if err != nil {
			fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		agentFile, err := updateAgentConfig()
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		if configOnly {
			fmt.Printf("\n%s  Agent configuration file updated: %s%s\n\n", utils.ColorGreen(), agentFile, utils.ColorReset())
			return
		}
		if upsertTarget {
			err = createTarget(c.Contexts[ctx])
			if err != nil {
				fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
				return
			}
		}
		if !asDaemon {
			_, err := utils.RunCommandNoCapture("Launching Symphony in standalone mode", "done", filepath.Join(u.HomeDir, ".symphony/symphony-api"), "-c", filepath.Join(u.HomeDir, ".symphony/symphony-agent-"+targetName+".json"), "-l", "Debug")
			if err != nil {
				fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
				return
			}
		} else {
			fmt.Printf("\n%s Running as daemon is not currently supported%s\n\n", utils.ColorRed(), utils.ColorReset())
		}
	},
}

func createTarget(ctx config.MaestroContext) error {
	target := model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      targetName,
			Namespace: targetNamespace,
		},
		Spec: &model.TargetSpec{
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "container",
							Provider: "providers.target.mqtt",
							Config: map[string]string{
								"brokerAddress": mqttBrokerAddress,
								"clientId":      mqttClientId,
								"requestTopic":  mqttRequestTopic,
								"responseTopic": mqttResponseTopic,
							},
						},
					},
				},
			},
		},
	}
	targetData, _ := json.Marshal(target)
	return utils.Upsert(ctx.Url, ctx.User, ctx.Secret, "target", targetName, targetData)
}

func updateAgentConfig() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %s", err.Error())
	}
	agentConfigTemplatePath := filepath.Join(dirname, ".symphony", "symphony-agent.json")
	agentConfigTemplateFile, err := os.Open(agentConfigTemplatePath)
	if err != nil {
		return "", err
	}
	defer agentConfigTemplateFile.Close()
	configData, err := io.ReadAll(agentConfigTemplateFile)
	if err != nil {
		return "", err
	}
	var agentConfig config.SymphonyAgentConfig
	err = json.Unmarshal(configData, &agentConfig)
	if err != nil {
		return "", err
	}

	agentConfig.API.Vendors[1].Managers[0].Properties.ProvidersTarget = targetName
	agentConfig.Bindings[0].Config.BrokerAddress = mqttBrokerAddress
	agentConfig.Bindings[0].Config.ClientID = mqttClientId
	agentConfig.Bindings[0].Config.RequestTopic = mqttRequestTopic
	agentConfig.Bindings[0].Config.ResponseTopic = mqttResponseTopic

	agentConfigPath := filepath.Join(dirname, ".symphony", "symphony-agent-"+targetName+".json")
	agentConfigStr, err := json.MarshalIndent(agentConfig, "", "  ")
	if err != nil {
		return "", err
	}

	// Write the updated JSON to a new file
	err = os.WriteFile(agentConfigPath, agentConfigStr, 0644)
	if err != nil {
		return "", err
	}
	return agentConfigPath, nil
}

func init() {
	AgentCmd.Flags().StringVarP(&targetNamespace, "namespace", "n", "default", "Target namespace")
	AgentCmd.Flags().StringVarP(&targetName, "target", "t", "", "Target name")
	AgentCmd.Flags().StringVarP(&mqttBrokerAddress, "mqtt-broker", "m", "", "MQTT broker address")
	AgentCmd.Flags().StringVarP(&mqttRequestTopic, "request-topic", "i", "coa-request", "MQTT request topic")
	AgentCmd.Flags().StringVarP(&mqttResponseTopic, "response-topic", "o", "coa-response", "MQTT response topic")
	AgentCmd.Flags().StringVarP(&mqttClientId, "mqtt-client-id", "e", "", "MQTT client id")
	AgentCmd.Flags().BoolVar(&asDaemon, "daemon", false, "Run agent as daemon")
	AgentCmd.Flags().BoolVar(&upsertTarget, "upsert-target", false, "Upsert target")
	AgentCmd.Flags().BoolVar(&configOnly, "config-only", false, "Only update the agent configuration file")
	AgentCmd.Flags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	AgentCmd.Flags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")

	AgentCmd.MarkFlagRequired("target")
	AgentCmd.MarkFlagRequired("mqtt-broker")
	AgentCmd.MarkFlagRequired("mqtt-client-id")
	RootCmd.AddCommand(AgentCmd)
}
