/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/eclipse-symphony/symphony/cli/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"
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
	agentBinary       string
	agentConfigFile   string
	reportStatus      bool
	reportInterval    time.Duration
	statusProperties  []string
	statusBagFile     string
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Symphony agent process and service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentWorkflow(cmd, "run")
	},
}

var AgentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Symphony agent in foreground",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentWorkflow(cmd, "run")
	},
}

var AgentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Symphony agent as a service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgentWorkflow(cmd, "install")
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

func runAgentWorkflow(cmd *cobra.Command, mode string) (result error) {
	ctx, err := resolveCurrentContext()
	if err != nil {
		return err
	}

	err = resolveMqttSettingsFromContext(ctx)
	if err != nil {
		return err
	}
	if !reportStatus && (len(statusProperties) > 0 || statusBagFile != "" || cmd.Flags().Changed("report-interval")) {
		return fmt.Errorf("--status-property, --status-properties and --report-interval require --report-status")
	}
	agentFile, err := updateAgentConfig(ctx)
	if err != nil {
		return err
	}
	if configOnly {
		fmt.Printf("\n%s  Agent configuration file updated: %s%s\n\n", utils.ColorGreen(), agentFile, utils.ColorReset())
		return nil
	}
	binary, err := resolveLocalBinary(agentBinary)
	if err != nil {
		return err
	}
	if upsertTarget {
		err = createTarget(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Registered Target %s/%s\n", targetNamespace, targetName)
	}

	switch mode {
	case "run":
		runCtx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		command := exec.Command(binary, "-c", agentFile, "-l", "Debug")
		command.Env = append(os.Environ(), "SYMPHONY_TARGET_NAME="+targetName, "SYMPHONY_API_URL="+strings.TrimRight(ctx.Url, "/")+"/")
		if ctx.Port > 0 {
			command.Env = append(command.Env, "USE_SERVICE_ACCOUNT_TOKENS=false")
		}
		command.Stdin, command.Stdout, command.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
		process, err := startLocalProcess(command)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, process.stop()) }()
		fmt.Fprintf(cmd.OutOrStdout(), "Agent configuration: %s\nPress Ctrl+C to stop.\n", agentFile)
		select {
		case <-runCtx.Done():
			return process.stop()
		case <-process.done:
			return process.err
		}
	case "install":
		err = installAgentService(targetName, agentFile, binary)
		if err != nil {
			return fmt.Errorf("install agent service: %w", err)
		}
		fmt.Printf("\n%s  Agent service installed for target: %s%s\n\n", utils.ColorGreen(), targetName, utils.ColorReset())
	default:
		return fmt.Errorf("unsupported agent mode: %s", mode)
	}
	return nil
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
	if err := validateAgentTarget(); err != nil {
		return err
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

func validateAgentTarget() error {
	if strings.TrimSpace(targetName) == "" {
		return fmt.Errorf("--target is required")
	}
	if errors := validation.IsDNS1123Subdomain(targetName); len(errors) != 0 {
		return fmt.Errorf("invalid Target name %q: %s", targetName, strings.Join(errors, "; "))
	}
	if errors := validation.IsDNS1123Label(targetNamespace); len(errors) != 0 {
		return fmt.Errorf("invalid Target namespace %q: %s", targetNamespace, strings.Join(errors, "; "))
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
								"clientID":      "symphony-control-plane-" + targetName,
								"requestTopic":  mqttRequestTopic,
								"responseTopic": mqttResponseTopic,
							},
						},
					},
				},
			},
		},
	}
	targetData, err := json.Marshal(target)
	if err != nil {
		return err
	}
	requestCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return utils.UpsertInNamespace(requestCtx, ctx.Url, ctx.User, ctx.Secret, "target", targetName, targetData, targetNamespace)
}

func updateAgentConfig(ctx config.MaestroContext) (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %s", err.Error())
	}
	agentConfigTemplatePath := filepath.Join(dirname, ".symphony", "symphony-agent.json")
	if agentConfigFile != "" {
		agentConfigTemplatePath = agentConfigFile
	}
	configData, err := os.ReadFile(agentConfigTemplatePath)
	if err != nil {
		return "", fmt.Errorf("read agent template (use --agent-config): %w", err)
	}
	agentConfig, err := parseRuntimeConfig(configData)
	if err != nil {
		return "", err
	}
	if err := configureAgentRuntime(agentConfig, ctx); err != nil {
		return "", err
	}
	agentConfigPath := filepath.Join(dirname, ".symphony", "symphony-agent-"+targetNamespace+"-"+targetName+".json")
	source, err := os.Stat(agentConfigTemplatePath)
	if err != nil {
		return "", err
	}
	destination, err := os.Stat(agentConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if destination != nil && os.SameFile(source, destination) {
		file, err := os.CreateTemp(filepath.Dir(agentConfigPath), "symphony-agent-effective-*.json")
		if err != nil {
			return "", err
		}
		agentConfigPath = file.Name()
		if err := file.Close(); err != nil {
			return "", err
		}
	}
	if err := writeRuntimeConfig(agentConfigPath, agentConfig); err != nil {
		return "", err
	}
	return agentConfigPath, nil
}

func installAgentService(target string, agentConfigPath string, binaryPath string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

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
`, target, strconv.Quote(binaryPath), strconv.Quote(agentConfigPath))
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
	AgentCmd.PersistentFlags().StringVar(&agentBinary, "agent-binary", "", "Symphony agent executable (default: ~/.symphony/symphony-api)")
	AgentCmd.PersistentFlags().StringVar(&agentConfigFile, "agent-config", "", "Agent template override (default: ~/.symphony/symphony-agent.json)")
	AgentCmd.PersistentFlags().BoolVar(&reportStatus, "report-status", false, "Configure the periodic status vendor and MQTT reporter")
	AgentCmd.PersistentFlags().DurationVar(&reportInterval, "report-interval", 5*time.Second, "Status reporting interval (whole seconds)")
	AgentCmd.PersistentFlags().StringArrayVar(&statusProperties, "status-property", nil, "Reported key=value, literal, $env:NAME, or Symphony expression (repeatable)")
	AgentCmd.PersistentFlags().StringVar(&statusBagFile, "status-properties", "", "JSON file containing the local status property bag")
	AgentCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Maestro CLI config file")
	AgentCmd.PersistentFlags().StringVarP(&configContext, "context", "", "", "Maestro CLI configuration context")

	AgentCmd.AddCommand(AgentRunCmd)
	AgentCmd.AddCommand(AgentInstallCmd)
	AgentCmd.AddCommand(AgentUninstallCmd)
	RootCmd.AddCommand(AgentCmd)
}
