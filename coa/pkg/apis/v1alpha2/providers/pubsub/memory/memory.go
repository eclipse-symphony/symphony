/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type InMemoryPubSubProvider struct {
	Config      InMemoryPubSubConfig               `json:"config"`
	Subscribers map[string][]v1alpha2.EventHandler `json:"subscribers"`
	Context     *contexts.ManagerContext
}

type InMemoryPubSubConfig struct {
	Name string `json:"name"`
}

func InMemoryPubSubConfigFromMap(properties map[string]string) (InMemoryPubSubConfig, error) {
	ret := InMemoryPubSubConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	return ret, nil
}

func (v *InMemoryPubSubProvider) ID() string {
	return v.Config.Name
}

func (s *InMemoryPubSubProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *InMemoryPubSubProvider) InitWithMap(properties map[string]string) error {
	config, err := InMemoryPubSubConfigFromMap(properties)
	if err != nil {
		log.Errorf("  P (Memory PubSub): failed to parse provider config from map %+v", err)
		return err
	}
	return i.Init(config)
}

func (i *InMemoryPubSubProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toInMemoryPubSubConfig(config)
	if err != nil {
		log.Errorf("  P (Memory PubSub): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid in-memory pub-sub provider config", v1alpha2.BadConfig)
	}
	i.Config = vConfig
	i.Subscribers = make(map[string][]v1alpha2.EventHandler)
	return nil
}
func (i *InMemoryPubSubProvider) Publish(topic string, event v1alpha2.Event) error {
	arr, ok := i.Subscribers[topic]
	if ok && arr != nil {
		for _, s := range arr {
			go func(handler v1alpha2.EventHandler, topic string, event v1alpha2.Event) {
				handler(topic, event)
			}(s, topic, event)
		}
	}
	return nil
}
func (i *InMemoryPubSubProvider) Subscribe(topic string, handler v1alpha2.EventHandler) error {
	arr, ok := i.Subscribers[topic]
	if !ok || arr == nil {
		i.Subscribers[topic] = make([]v1alpha2.EventHandler, 0)
	}
	i.Subscribers[topic] = append(i.Subscribers[topic], handler)

	return nil
}

func toInMemoryPubSubConfig(config providers.IProviderConfig) (InMemoryPubSubConfig, error) {
	ret := InMemoryPubSubConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (a *InMemoryPubSubProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &InMemoryPubSubProvider{}
	if config == nil {
		err := ret.Init(a.Config)
		if err != nil {
			return nil, err
		}
	} else {
		err := ret.Init(config)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
