/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type TargetProviderConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}
type SymphonyAgentConfig struct {
	SiteInfo struct {
		SiteId      string `json:"siteId"`
		CurrentSite struct {
			BaseURL  string `json:"baseUrl"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"currentSite"`
	} `json:"siteInfo"`
	API struct {
		Vendors []struct {
			Type         string `json:"type"`
			Route        string `json:"route"`
			LoopInterval int    `json:"loopInterval,omitempty"`
			Managers     []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				Properties struct {
					ProvidersPersistentState string `json:"providers.persistentstate"`
					IsTarget                 string `json:"isTarget"`
					TargetNames              string `json:"targetNames"`
					ProvidersConfig          string `json:"providers.config"`
					ProvidersSecret          string `json:"providers.secret"`
					PollEnabled              string `json:"poll.enabled"`
				} `json:"properties"`
				Providers map[string]TargetProviderConfig `json:"providers"`
			} `json:"managers"`
		} `json:"vendors"`
	} `json:"api"`
	Bindings []struct {
		Type   string `json:"type"`
		Config struct {
			BrokerAddress string `json:"brokerAddress"`
			ClientID      string `json:"clientID"`
			RequestTopic  string `json:"requestTopic"`
			ResponseTopic string `json:"responseTopic"`
		} `json:"config"`
	} `json:"bindings"`
}

type MaestroContext struct {
	Url    string `json:"url"`
	User   string `json:"user"`
	Secret string `json:"secret,omitempty"`
}
type MaestroConfig struct {
	DefaultContext string                    `json:"default,omitempty"`
	Contexts       map[string]MaestroContext `json:"contexts,omitempty"`
}

func UpdateMaestroConfig(context string, address string) error {
	config := GetMaestroConfig("")
	if config.Contexts == nil {
		config.Contexts = make(map[string]MaestroContext)
	}
	config.Contexts[context] = MaestroContext{
		Url:    "http://" + address + ":8080/v1alpha2",
		User:   "admin",
		Secret: "",
	}
	config.DefaultContext = context
	return SaveMaestroConfig(config)
}
func SaveMaestroConfig(config MaestroConfig) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(dirname, ".symphony", ".config.json")
	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 064)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if _, err := file.Write(b); err != nil {
		return err
	}
	return nil
}
func GetMaestroConfig(path string) MaestroConfig {
	var files []string
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	folderName := filepath.Join(dirname, ".symphony")
	fileName := filepath.Join(folderName, ".config.json")
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		err = os.MkdirAll(folderName, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
	if path == "" {
		files = []string{fileName}
	} else {
		files = strings.Split(path, ":")
	}
	ret := MaestroConfig{
		Contexts: make(map[string]MaestroContext),
	}
	for _, f := range files {
		if f != "" {
			content, err := os.ReadFile(f)
			if err == nil {
				var config MaestroConfig
				err = json.Unmarshal(content, &config)
				if err == nil {
					for k, v := range config.Contexts {
						ret.Contexts[k] = v
					}
					if config.DefaultContext != "" {
						ret.DefaultContext = config.DefaultContext
					}
				}
			}
		}
	}
	if len(ret.Contexts) == 0 {
		ret.Contexts["default"] = MaestroContext{
			Url:    "http://localhost:8080/v1alpha2",
			User:   "admin",
			Secret: "",
		}
		ret.DefaultContext = "default"
	}
	return ret
}
