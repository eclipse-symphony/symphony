/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

const (
	DefaultRetryCount      = 5
	DefaultRetryWaitSecond = 10
)

type InMemoryPubSubProvider struct {
	Config      InMemoryPubSubConfig               `json:"config"`
	Subscribers map[string][]v1alpha2.EventHandler `json:"subscribers"`
	Context     *contexts.ManagerContext
}

type InMemoryPubSubConfig struct {
	Name                      string `json:"name"`
	SubscriberRetryCount      int    `json:"subscriberRetryCount"`
	SubscriberRetryWaitSecond int    `json:"subscriberRetryWaitSecond"`
}

func InMemoryPubSubConfigFromMap(properties map[string]string) (InMemoryPubSubConfig, error) {
	ret := InMemoryPubSubConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	ret.SubscriberRetryCount = 0
	if v, ok := properties["subscriberRetryCount"]; ok {
		val := v
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'SubscriberRetryCount' setting of Memory pub-sub provider", v1alpha2.BadConfig)
			}
			ret.SubscriberRetryCount = n
		}
	}
	if ret.SubscriberRetryCount < 0 {
		return ret, v1alpha2.NewCOAError(nil, "negative int value is not allowed in the 'SubscriberRetryCount' setting of Memory pub-sub provider", v1alpha2.BadConfig)
	}
	ret.SubscriberRetryWaitSecond = 0
	if v, ok := properties["subscriberRetryWaitSecond"]; ok {
		val := v
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'SubscriberRetryWaitSecond' setting of Memory pub-sub provider", v1alpha2.BadConfig)
			}
			ret.SubscriberRetryWaitSecond = n
		}
	}
	if ret.SubscriberRetryWaitSecond == 0 {
		ret.SubscriberRetryWaitSecond = DefaultRetryWaitSecond
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
				shouldRetry := true
				count := 0
				for shouldRetry && count <= i.Config.SubscriberRetryCount {
					shouldRetry = v1alpha2.EventShouldRetryWrapper(handler, topic, event)
					if shouldRetry {
						count++
						time.Sleep(time.Duration(i.Config.SubscriberRetryWaitSecond) * time.Second)
					}
				}
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

	var configs map[string]interface{}
	err = json.Unmarshal(data, &configs)
	if err != nil {
		return ret, err
	}
	configStrings := map[string]string{}
	for k, v := range configs {
		configStrings[k] = utils.FormatAsString(v)
	}

	ret, err = InMemoryPubSubConfigFromMap(configStrings)
	if err != nil {
		return ret, err
	}
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

func (a *InMemoryPubSubProvider) Cancel() context.CancelFunc {
	return func() {
		log.Info("  P (Memory PubSub): canceling")
	}
}
