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
	"runtime"
	"strings"

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
	targetName        string
	targetNamespace   string
	upsertTarget      bool
	configOnly        bool
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Symphony agent process and service",
	Run: func(cmd *cobra.Command, args []string) {
		runAgentWorkflow("run")
	},
}

var AgentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Symphony agent in foreground",
	Run: func(cmd *cobra.Command, args []string) {
		runAgentWorkflow("run")
	},
}

var AgentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Symphony agent as a service",
	Run: func(cmd *cobra.Command, args []string) {
		runAgentWorkflow("install")
	},
}

var AgentUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Symphony agent service",
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(targetName) == "" {
			fmt.Printf("\n%s  --target is required%s\n\n", utils.ColorRed(), utils.ColorReset())
			return
		}
		err := uninstallAgentService(targetName)
		if err != nil {
			fmt.Printf("\n%s  Failed to uninstall agent service: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		fmt.Printf("\n%s  Agent service removed for target: %s%s\n\n", utils.ColorGreen(), targetName, utils.ColorReset())
	},
}

func runAgentWorkflow(mode string) {
	ctx, err := resolveCurrentContext()
	if err != nil {
		fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
		return
	}

	err = resolveMqttSettingsFromContext(ctx)
	if err != nil {
		fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
		return
	}

	agentFile, err := updateAgentConfig(ctx)
	if err != nil {
		fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
		return
	}
	if configOnly {
		fmt.Printf("\n%s  Agent configuration file updated: %s%s\n\n", utils.ColorGreen(), agentFile, utils.ColorReset())
		return
	}

	if upsertTarget {
		err = createTarget(ctx)
		if err != nil {
			fmt.Printf("\n%s%s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
	}

	switch mode {
	case "run":
		u, err := user.Current()
		if err != nil {
			fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		_, err = utils.RunCommandNoCapture("Launching Symphony agent", "done", filepath.Join(u.HomeDir, ".symphony/symphony-api"), "-c", agentFile, "-l", "Debug")
		if err != nil {
			fmt.Printf("\n%s  Failed: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
		}
	case "install":
		err = installAgentService(targetName, agentFile)
		if err != nil {
			fmt.Printf("\n%s  Failed to install agent service: %s%s\n\n", utils.ColorRed(), err.Error(), utils.ColorReset())
			return
		}
		fmt.Printf("\n%s  Agent service installed for target: %s%s\n\n", utils.ColorGreen(), targetName, utils.ColorReset())
	default:
		fmt.Printf("\n%s  Unsupported agent mode: %s%s\n\n", utils.ColorRed(), mode, utils.ColorReset())
	}
}

func resolveCurrentContext() (config.MaestroContext, error) {
	c := config.GetMaestroConfig(configFile)
	ctxName := c.DefaultContext
	if configContext != "" {
		ctxName = configContext
	}
	if ctxName == "" {
		ctxName = "default"
	}

	ctx, ok := c.Contexts[ctxName]
	if !ok {
		return config.MaestroContext{}, fmt.Errorf("context %q not found in Maestro config", ctxName)
	}
	return ctx, nil
}

func resolveMqttSettingsFromContext(ctx config.MaestroContext) error {
	if strings.TrimSpace(targetName) == "" {
		return fmt.Errorf("--target is required")
	}

	if mqttBrokerAddress == "" {
		mqttBrokerAddress = ctx.Mqtt.BrokerAddress
	}
	if mqttRequestTopic == "" {
		if ctx.Mqtt.RequestTopic != "" {
			mqttRequestTopic = ctx.Mqtt.RequestTopic
		} else {
			mqttRequestTopic = "coa-request"
		}
	}
	if mqttResponseTopic == "" {
		if ctx.Mqtt.ResponseTopic != "" {
			mqttResponseTopic = ctx.Mqtt.ResponseTopic
		} else {
			mqttResponseTopic = "coa-response"
		}
	}
	if mqttClientId == "" {
		mqttClientId = ctx.Mqtt.ClientID
	}
	if mqttClientId == "" {
		mqttClientId = "symphony-agent-" + targetName
	}

	if mqttBrokerAddress == "" {
		return fmt.Errorf("MQTT broker is not configured. Use --mqtt-broker or run maestro up --with-mqtt-broker to persist broker info in context")
	}

	return nil
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

func updateAgentConfig(ctx config.MaestroContext) (string, error) {
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
	ensureAgentConfigDefaults(&agentConfig)

	agentConfig.SiteInfo.CurrentSite.BaseURL = ctx.Url
	agentConfig.SiteInfo.CurrentSite.Username = ctx.User
	agentConfig.SiteInfo.CurrentSite.Password = ctx.Secret

	agentConfig.API.Vendors[1].Managers[0].Properties.TargetNames = targetName
	agentConfig.API.Vendors[1].Managers[0].Properties.TargetNamespace = targetNamespace
	agentConfig.API.Vendors[1].Managers[0].Providers[targetName] = config.TargetProviderConfig{
		Type:   "providers.target.mock",
		Config: map[string]interface{}{},
	}
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

func ensureAgentConfigDefaults(agentConfig *config.SymphonyAgentConfig) {
	if strings.TrimSpace(agentConfig.ShutdownGracePeriod) == "" {
		agentConfig.ShutdownGracePeriod = "30s"
	}
	if len(agentConfig.API.Pubsub) == 0 {
		agentConfig.API.Pubsub = map[string]interface{}{
			"shared": true,
			"provider": map[string]interface{}{
				"type":   "providers.pubsub.memory",
				"config": map[string]interface{}{},
			},
		}
	}
	if len(agentConfig.API.Keylock) == 0 {
		agentConfig.API.Keylock = map[string]interface{}{
			"shared": true,
			"provider": map[string]interface{}{
				"type": "providers.keylock.memory",
				"config": map[string]interface{}{
					"mode":          "Global",
					"cleanInterval": 30,
					"purgeDuration": 43200,
				},
			},
		}
	}
}

func installAgentService(target string, agentConfigPath string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	binaryPath := filepath.Join(u.HomeDir, ".symphony", "symphony-api")
	serviceName := "symphony-agent-" + target

	switch runtime.GOOS {
	case "linux":
		serviceDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
		if err = os.MkdirAll(serviceDir, 0755); err != nil {
			return err
		}
		serviceFile := filepath.Join(serviceDir, serviceName+".service")
		serviceContent := fmt.Sprintf(`[Unit]
Description=Symphony Agent (%s)
After=network.target

[Service]
Type=simple
ExecStart=%s -c %s -l Debug
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, target, binaryPath, agentConfigPath)
		if err = os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
			return err
		}

		if _, _, err = utils.RunCommand("Reloading user systemd", "done", verbose, "systemctl", "--user", "daemon-reload"); err != nil {
			return err
		}
		if _, _, err = utils.RunCommand("Enabling agent service", "done", verbose, "systemctl", "--user", "enable", "--now", serviceName); err != nil {
			return err
		}
		return nil
	case "windows":
		binPath := fmt.Sprintf("\"%s\" -c \"%s\" -l Debug", binaryPath, agentConfigPath)
		if _, _, err = utils.RunCommand("Creating agent Windows service", "done", verbose, "sc.exe", "create", serviceName, "binPath=", binPath, "start=", "auto"); err != nil {
			return err
		}
		if _, _, err = utils.RunCommand("Starting agent Windows service", "done", verbose, "sc.exe", "start", serviceName); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported OS for install: %s", runtime.GOOS)
	}
}

func uninstallAgentService(target string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	serviceName := "symphony-agent-" + target

	switch runtime.GOOS {
	case "linux":
		_, _, _ = utils.RunCommand("Stopping agent service", "done", verbose, "systemctl", "--user", "disable", "--now", serviceName)
		serviceFile := filepath.Join(u.HomeDir, ".config", "systemd", "user", serviceName+".service")
		_ = os.Remove(serviceFile)
		_, _, _ = utils.RunCommand("Reloading user systemd", "done", verbose, "systemctl", "--user", "daemon-reload")
		return nil
	case "windows":
		_, _, _ = utils.RunCommand("Stopping agent Windows service", "done", verbose, "sc.exe", "stop", serviceName)
		_, _, err = utils.RunCommand("Deleting agent Windows service", "done", verbose, "sc.exe", "delete", serviceName)
		return err
	default:
		return fmt.Errorf("unsupported OS for uninstall: %s", runtime.GOOS)
	}
}

func init() {
	AgentCmd.PersistentFlags().StringVarP(&targetNamespace, "namespace", "n", "default", "Target namespace")
	AgentCmd.PersistentFlags().StringVarP(&targetName, "target", "t", "", "Target name")
	AgentCmd.PersistentFlags().StringVarP(&mqttBrokerAddress, "mqtt-broker", "m", "", "MQTT broker address (overrides context)")
	AgentCmd.PersistentFlags().StringVarP(&mqttRequestTopic, "request-topic", "i", "", "MQTT request topic (default: context or coa-request)")
	AgentCmd.PersistentFlags().StringVarP(&mqttResponseTopic, "response-topic", "o", "", "MQTT response topic (default: context or coa-response)")
	AgentCmd.PersistentFlags().StringVarP(&mqttClientId, "mqtt-client-id", "e", "", "MQTT client id (default: context or symphony-agent-<target>)")
	AgentCmd.PersistentFlags().BoolVar(&upsertTarget, "upsert-target", false, "Upsert target")
	AgentCmd.PersistentFlags().BoolVar(&configOnly, "config-only", false, "Only update the agent configuration file")
	AgentCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	AgentCmd.PersistentFlags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")

	AgentCmd.AddCommand(AgentRunCmd)
	AgentCmd.AddCommand(AgentInstallCmd)
	AgentCmd.AddCommand(AgentUninstallCmd)
	RootCmd.AddCommand(AgentCmd)
}
