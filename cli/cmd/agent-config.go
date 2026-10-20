/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
)

func configObject(parent map[string]interface{}, key string) (map[string]interface{}, error) {
	if value, exists := parent[key]; exists {
		object, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s must be an object", key)
		}
		return object, nil
	}
	object := make(map[string]interface{})
	parent[key] = object
	return object, nil
}

func configureAgentRuntime(cfg map[string]interface{}, ctx config.MaestroContext) error {
	vendors, err := runtimeVendors(cfg)
	if err != nil {
		return err
	}
	site, err := configObject(cfg, "siteInfo")
	if err != nil {
		return err
	}
	currentSite, err := configObject(site, "currentSite")
	if err != nil {
		return err
	}
	currentSite["baseUrl"], currentSite["username"], currentSite["password"] = ctx.Url, ctx.User, ctx.Secret
	api := cfg["api"].(map[string]interface{})
	if _, exists := cfg["shutdownGracePeriod"]; !exists {
		cfg["shutdownGracePeriod"] = "30s"
	}
	if _, exists := api["pubsub"]; !exists {
		api["pubsub"] = map[string]interface{}{
			"shared": true, "provider": map[string]interface{}{
				"type": "providers.pubsub.memory", "config": map[string]interface{}{},
			},
		}
	}
	if _, exists := api["keylock"]; !exists {
		api["keylock"] = map[string]interface{}{
			"shared": true, "provider": map[string]interface{}{
				"type": "providers.keylock.memory", "config": map[string]interface{}{
					"mode": "Global", "cleanInterval": 30, "purgeDuration": 43200,
				},
			},
		}
	}
	hasDeployment := false
	for _, entry := range vendors {
		vendor, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("vendor must be an object")
		}
		if vendor["type"] != "vendors.solutionversion" {
			continue
		}
		managers, ok := vendor["managers"].([]interface{})
		if !ok {
			return fmt.Errorf("solutionversion vendor requires managers")
		}
		for _, entry := range managers {
			manager, ok := entry.(map[string]interface{})
			if !ok {
				return fmt.Errorf("manager must be an object")
			}
			if manager["type"] != "managers.symphony.solutionversion" {
				continue
			}
			properties, err := configObject(manager, "properties")
			if err != nil {
				return err
			}
			properties["isTarget"] = "true"
			properties["targetNames"] = targetName
			properties["targetNamespace"] = targetNamespace
			providers, err := configObject(manager, "providers")
			if err != nil {
				return err
			}
			if agentConfigFile == "" {
				if _, exists := providers["instance"]; !exists {
					providers["instance"] = map[string]interface{}{
						"type": "providers.target.mock", "config": map[string]interface{}{},
					}
				}
			}
			hasDeployment = true
		}
	}
	if !hasDeployment && !reportStatus {
		return fmt.Errorf("agent template requires a solutionversion manager or --report-status")
	}
	if hasDeployment {
		bindings, exists := cfg["bindings"]
		if !exists {
			bindings = []interface{}{}
		}
		list, ok := bindings.([]interface{})
		if !ok {
			return fmt.Errorf("bindings must be an array")
		}
		var deploymentBinding map[string]interface{}
		for _, entry := range list {
			binding, ok := entry.(map[string]interface{})
			if !ok {
				return fmt.Errorf("binding must be an object")
			}
			if binding["type"] == "bindings.mqtt" {
				deploymentBinding = binding
				break
			}
		}
		if deploymentBinding == nil {
			deploymentBinding = map[string]interface{}{"type": "bindings.mqtt"}
			list = append(list, deploymentBinding)
		}
		settings, err := configObject(deploymentBinding, "config")
		if err != nil {
			return err
		}
		settings["brokerAddress"], settings["clientID"] = mqttBrokerAddress, mqttClientId
		settings["requestTopic"], settings["responseTopic"] = mqttRequestTopic, mqttResponseTopic
		cfg["bindings"] = list
	}
	if reportStatus {
		return configureStatusVendor(api, ctx)
	}
	return nil
}

func parseStatusBag(data []byte) (map[string]string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("status property bag must be an object of strings")
	}
	bag := make(map[string]string, len(raw))
	for key, value := range raw {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(key) == "" || key == "agent.reportedAt" {
			return nil, fmt.Errorf("invalid status property %q: values must be strings and agent.reportedAt is reserved", key)
		}
		bag[key] = text
	}
	return bag, nil
}

func configureStatusVendor(api map[string]interface{}, ctx config.MaestroContext) error {
	if reportInterval < time.Second || reportInterval%time.Second != 0 {
		return fmt.Errorf("--report-interval must be a positive whole number of seconds")
	}
	vendors := api["vendors"].([]interface{})
	var vendor map[string]interface{}
	for _, entry := range vendors {
		candidate := entry.(map[string]interface{})
		if candidate["type"] == "vendors.status" {
			vendor = candidate
			break
		}
	}
	if vendor == nil {
		vendor = map[string]interface{}{"type": "vendors.status", "managers": []interface{}{}}
		vendors = append(vendors, vendor)
	}
	vendor["loopInterval"] = int(reportInterval / time.Second)
	managers, ok := vendor["managers"].([]interface{})
	if !ok {
		return fmt.Errorf("status vendor requires managers")
	}
	var manager map[string]interface{}
	for _, entry := range managers {
		candidate, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("status manager must be an object")
		}
		if candidate["type"] == "managers.symphony.status" {
			manager = candidate
			break
		}
	}
	if manager == nil {
		manager = map[string]interface{}{"name": "status-manager", "type": "managers.symphony.status"}
		managers = append(managers, manager)
	}
	properties, err := configObject(manager, "properties")
	if err != nil {
		return err
	}
	bag := map[string]string{"agent.connected": "true"}
	if raw, exists := properties["report.properties"]; exists {
		text, ok := raw.(string)
		if !ok {
			return fmt.Errorf("report.properties must be a JSON-encoded string")
		}
		bag, err = parseStatusBag([]byte(text))
		if err != nil {
			return err
		}
	}
	if statusBagFile != "" {
		data, err := os.ReadFile(statusBagFile)
		if err != nil {
			return err
		}
		bag, err = parseStatusBag(data)
		if err != nil {
			return err
		}
	}
	for _, property := range statusProperties {
		key, value, found := strings.Cut(property, "=")
		if !found || strings.TrimSpace(key) == "" || key == "agent.reportedAt" {
			return fmt.Errorf("--status-property requires key=value; agent.reportedAt is reserved")
		}
		bag[key] = value
	}
	data, err := json.Marshal(bag)
	if err != nil {
		return err
	}
	properties["poll.enabled"] = "true"
	properties["target.name"], properties["target.namespace"] = targetName, targetNamespace
	properties["report.properties"] = string(data)
	reporterName, _ := properties["providers.reporter"].(string)
	if reporterName == "" {
		reporterName = "mqtt-status-reporter"
	}
	properties["providers.reporter"] = reporterName
	providers, err := configObject(manager, "providers")
	if err != nil {
		return err
	}
	reporter, err := configObject(providers, reporterName)
	if err != nil {
		return err
	}
	reporter["type"] = "providers.reporter.mqtt"
	settings, err := configObject(reporter, "config")
	if err != nil {
		return err
	}
	requestTopic, responseTopic, route := ctx.Mqtt.StatusRequestTopic, ctx.Mqtt.StatusResponseTopic, ctx.Mqtt.StatusRoute
	if requestTopic == "" {
		requestTopic = localStatusRequestTopic
	}
	if responseTopic == "" {
		responseTopic = localStatusResponseTopic
	}
	if route == "" {
		route = "references"
	}
	settings["brokerAddress"], settings["clientID"] = mqttBrokerAddress, mqttClientId+"-status"
	settings["requestTopic"], settings["responseTopic"], settings["route"] = requestTopic, responseTopic, route
	vendor["managers"] = managers
	api["vendors"] = vendors
	return nil
}
