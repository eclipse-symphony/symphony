/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package memory

import (
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

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

func (s *InMemoryPubSubProvider) SetContext(ctx *contexts.ManagerContext) error {
	s.Context = ctx
	return nil
}

func (i *InMemoryPubSubProvider) InitWithMap(properties map[string]string) error {
	config, err := InMemoryPubSubConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (i *InMemoryPubSubProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toInMemoryPubSubConfig(config)
	if err != nil {
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
