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

package host

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	bindings "github.com/azure/symphony/coa/pkg/apis/v1alpha2/bindings"
	http "github.com/azure/symphony/coa/pkg/apis/v1alpha2/bindings/http"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	mf "github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	pf "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providerfactory"
	pv "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type HostConfig struct {
	API      APIConfig       `json:"api"`
	Bindings []BindingConfig `json:"bindings"`
}
type PubSubConfig struct {
	Shared   bool              `json:"shared"`
	Provider mf.ProviderConfig `json:"provider"`
}
type APIConfig struct {
	Vendors []vendors.VendorConfig `json:"vendors"`
	PubSub  PubSubConfig           `json:"pubsub,omitempty"`
}

type BindingConfig struct {
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type VendorSpec struct {
	Vendor       vendors.IVendor
	LoopInterval int
}
type APIHost struct {
	Vendors              []VendorSpec
	Bindings             []bindings.IBinding
	SharedPubSubProvider pv.IProvider
}

func (h *APIHost) Launch(config HostConfig,
	vendorFactories []vendors.IVendorFactory,
	managerFactories []mf.IManagerFactroy,
	providerFactories []pf.IProviderFactory, wait bool) error {
	h.Vendors = make([]VendorSpec, 0)
	h.Bindings = make([]bindings.IBinding, 0)
	log.Info("--- launching COA host ---")
	for _, v := range config.API.Vendors {
		created := false
		for _, factory := range vendorFactories {
			vendor, err := factory.CreateVendor(v)
			if err != nil {
				return err
			}
			if vendor != nil {
				var pubsubProvider pv.IProvider
				// make pub/sub provider
				if config.API.PubSub.Provider.Type != "" {
					if config.API.PubSub.Shared && h.SharedPubSubProvider != nil {
						pubsubProvider = h.SharedPubSubProvider
					} else {
						for _, providerFactory := range providerFactories {
							mProvider, err := providerFactory.CreateProvider(
								config.API.PubSub.Provider.Type,
								config.API.PubSub.Provider.Config)
							if err != nil {
								return err
							}
							pubsubProvider = mProvider
							if config.API.PubSub.Shared {
								h.SharedPubSubProvider = pubsubProvider
							}
							break
						}
					}
				}

				// make other providers
				providers := make(map[string]map[string]pv.IProvider, 0)
				for _, providerFactory := range providerFactories {
					mProviders, err := providerFactory.CreateProviders(v)
					if err != nil {
						return err
					}
					for k, _ := range mProviders {
						if _, ok := providers[k]; ok {
							for ik, iv := range mProviders[k] {
								if _, ok := providers[k][ik]; !ok {
									providers[k][ik] = iv
								} else {
									//TODO: what to do if there are conflicts?
								}
							}
						} else {
							providers[k] = mProviders[k]
						}
					}
				}
				if pubsubProvider != nil {
					err = vendor.Init(v, managerFactories, providers, pubsubProvider.(pubsub.IPubSubProvider))
				} else {
					err = vendor.Init(v, managerFactories, providers, nil)
				}
				if err != nil {
					return err
				}
				h.Vendors = append(h.Vendors, VendorSpec{Vendor: vendor, LoopInterval: v.LoopInterval})
				created = true
				break
			}
		}
		if !created {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("no vendor factories can provide vendor type '%s'", v.Type), v1alpha2.BadConfig)
		}
	}
	if len(h.Vendors) > 0 {
		var wg sync.WaitGroup
		for _, v := range h.Vendors {
			if v.LoopInterval > 0 {
				if wait {
					wg.Add(1)
				}
				go func(v VendorSpec) {
					v.Vendor.RunLoop(time.Duration(v.LoopInterval) * time.Second)
				}(v)
			}
		}
		if len(config.Bindings) > 0 {
			endpoints := make([]v1alpha2.Endpoint, 0)
			for _, v := range h.Vendors {
				endpoints = append(endpoints, v.Vendor.GetEndpoints()...)
			}

			for _, b := range config.Bindings {
				switch b.Type {
				case "bindings.http":
					if wait {
						wg.Add(1)
					}
					binding, err := h.launchHTTP(b.Config, endpoints)
					if err != nil {
						return err
					}
					h.Bindings = append(h.Bindings, binding)
				case "bindings.mqtt":
					if wait {
						wg.Add(1)
					}
					binding, err := h.launchMQTT(b.Config, endpoints)
					if err != nil {
						return err
					}
					h.Bindings = append(h.Bindings, binding)
				default:
					return v1alpha2.NewCOAError(nil, fmt.Sprintf("binding type '%s' is not recognized", b.Type), v1alpha2.BadConfig)
				}
			}
		}
		wg.Wait()
		return nil
	} else {
		return v1alpha2.NewCOAError(nil, "no vendors are found", v1alpha2.MissingConfig)
	}
}

func (h *APIHost) launchHTTP(config interface{}, endpoints []v1alpha2.Endpoint) (bindings.IBinding, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	httpConfig := http.HttpBindingConfig{}
	err = json.Unmarshal(data, &httpConfig)
	if err != nil {
		return nil, err
	}
	binding := http.HttpBinding{}
	return binding, binding.Launch(httpConfig, endpoints)
}
func (h *APIHost) launchMQTT(config interface{}, endpoints []v1alpha2.Endpoint) (bindings.IBinding, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	mqttConfig := mqtt.MQTTBindingConfig{}
	err = json.Unmarshal(data, &mqttConfig)
	if err != nil {
		return nil, err
	}
	binding := mqtt.MQTTBinding{}
	return binding, binding.Launch(mqttConfig, endpoints)
}
