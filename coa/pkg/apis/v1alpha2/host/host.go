/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package host

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	bindings "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings"
	http "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/http"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	mf "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	pf "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providerfactory"
	pv "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var log = logger.NewLogger("coa.runtime")
var defaultShutdownGracePeriod = "30s"

var hostIsReadyFlag bool = false
var rwLock sync.RWMutex

func IsHostReady() bool {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return hostIsReadyFlag
}

func SetHostReadyFlag(ready bool) {
	rwLock.Lock()
	defer rwLock.Unlock()
	hostIsReadyFlag = ready
}

type HostConfig struct {
	SiteInfo            v1alpha2.SiteInfo `json:"siteInfo"`
	API                 APIConfig         `json:"api"`
	Bindings            []BindingConfig   `json:"bindings"`
	ShutdownGracePeriod string            `json:"shutdownGracePeriod"`
}
type PubSubConfig struct {
	Shared   bool              `json:"shared"`
	Provider mf.ProviderConfig `json:"provider"`
}

type PublicProviderConfig struct {
	Type   string                `json:"type"`
	Config KeyLockProviderConfig `json:"config"`
}

type KeyLockProviderConfig struct {
	Mode string `json:"mode"`
}

type KeyLockConfig struct {
	Shared   bool                 `json:"shared"`
	Provider PublicProviderConfig `json:"provider"`
}

type APIConfig struct {
	Vendors []vendors.VendorConfig `json:"vendors"`
	PubSub  PubSubConfig           `json:"pubsub,omitempty"`
	KeyLock KeyLockConfig          `json:"keylock,omitempty"`
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
	Vendors               []VendorSpec
	Bindings              []bindings.IBinding
	SharedPubSubProvider  pv.IProvider
	SharedKeyLockProvider pv.IProvider
	ShutdownGracePeriod   time.Duration
}

func overrideWithEnvVariable(value string, env string) string {
	if os.Getenv(env) != "" {
		return os.Getenv(env)
	}
	return value
}

func (h *APIHost) Launch(config HostConfig,
	vendorFactories []vendors.IVendorFactory,
	managerFactories []mf.IManagerFactroy,
	providerFactories []pf.IProviderFactory, wait bool) error {
	h.Vendors = make([]VendorSpec, 0)
	h.Bindings = make([]bindings.IBinding, 0)
	log.Info("--- launching COA host ---")
	var err error
	h.ShutdownGracePeriod, err = time.ParseDuration(config.ShutdownGracePeriod)
	if err != nil {
		log.Warnf("failed to parse shutdownGracePeriod '%s' from config, using default '%s'", config.ShutdownGracePeriod, defaultShutdownGracePeriod)
		h.ShutdownGracePeriod, _ = time.ParseDuration(defaultShutdownGracePeriod)
	}
	if config.SiteInfo.SiteId == "" {
		return v1alpha2.NewCOAError(nil, "siteId is not specified", v1alpha2.BadConfig)
	}

	config.SiteInfo.SiteId = overrideWithEnvVariable(config.SiteInfo.SiteId, "SYMPHONY_SITE_ID")
	config.SiteInfo.CurrentSite.BaseUrl = overrideWithEnvVariable(config.SiteInfo.CurrentSite.BaseUrl, "SYMPHONY_API_BASE_URL")
	config.SiteInfo.CurrentSite.Username = overrideWithEnvVariable(config.SiteInfo.CurrentSite.Username, "SYMPHONY_API_USER")
	config.SiteInfo.CurrentSite.Password = overrideWithEnvVariable(config.SiteInfo.CurrentSite.Password, "SYMPHONY_API_PASSWORD")
	config.SiteInfo.ParentSite.BaseUrl = overrideWithEnvVariable(config.SiteInfo.ParentSite.BaseUrl, "PARENT_SYMPHONY_API_BASE_URL")
	config.SiteInfo.ParentSite.Username = overrideWithEnvVariable(config.SiteInfo.ParentSite.Username, "PARENT_SYMPHONY_API_USER")
	config.SiteInfo.ParentSite.Password = overrideWithEnvVariable(config.SiteInfo.ParentSite.Password, "PARENT_SYMPHONY_API_PASSWORD")

	var pubsubProvider pv.IProvider
	for _, v := range config.API.Vendors {
		v.SiteInfo = config.SiteInfo
		created := false
		for _, factory := range vendorFactories {
			vendor, err := factory.CreateVendor(v)
			if err != nil {
				return err
			}
			if vendor != nil {
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

				if config.API.KeyLock.Provider.Type != "" {
					if h.SharedKeyLockProvider == nil {
						if config.API.KeyLock.Provider.Config.Mode != "Global" {
							return errors.New("Expected Global KeyLockProviderConfig")
						}
						for _, providerFactory := range providerFactories {
							mProvider, err := providerFactory.CreateProvider(
								config.API.KeyLock.Provider.Type,
								config.API.KeyLock.Provider.Config)
							if err != nil {
								return err
							}
							h.SharedKeyLockProvider = mProvider
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
	if len(h.Vendors) == 0 {
		return v1alpha2.NewCOAError(nil, "no vendors are found", v1alpha2.MissingConfig)
	}
	var evaluationContext *utils.EvaluationContext
	for _, v := range h.Vendors {
		if _, ok := v.Vendor.(vendors.IEvaluationContextVendor); ok {
			log.Info("--- evaluation context established ---")
			evaluationContext = v.Vendor.(vendors.IEvaluationContextVendor).GetEvaluationContext()
		}
	}

	if evaluationContext != nil {
		for _, v := range h.Vendors {
			log.Infof("--- evaluation context is sent to vendor: %s ---", v.Vendor.GetInfo().Name)
			v.Vendor.SetEvaluationContext(evaluationContext)
		}
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, v := range h.Vendors {
		if v.LoopInterval > 0 {
			if wait {
				wg.Add(1)
			}
			go func(v VendorSpec) {
				defer func() {
					if wait {
						wg.Done()
					}
				}()
				v.Vendor.RunLoop(ctx, time.Duration(v.LoopInterval)*time.Second)
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
				var binding bindings.IBinding
				var err error
				if h.SharedPubSubProvider != nil {
					binding, err = h.launchHTTP(b.Config, endpoints, h.SharedPubSubProvider.(pubsub.IPubSubProvider))
				} else {
					var bindingPubsub pv.IProvider
					for _, providerFactory := range providerFactories {
						mProvider, err := providerFactory.CreateProvider(
							config.API.PubSub.Provider.Type,
							config.API.PubSub.Provider.Config)
						if err != nil {
							return err
						}
						bindingPubsub = mProvider
						break
					}
					binding, err = h.launchHTTP(b.Config, endpoints, bindingPubsub.(pubsub.IPubSubProvider))
				}
				if err != nil {
					return err
				}
				h.Bindings = append(h.Bindings, binding)
			case "bindings.mqtt":
				binding, err := h.launchMQTT(b.Config, endpoints)
				if err != nil {
					return err
				}
				h.Bindings = append(h.Bindings, binding)
			default:
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("binding type '%s' is not recognized", b.Type), v1alpha2.BadConfig)
			}
		}

		if len(config.Bindings) > 0 {
			for _, binding := range h.Bindings {
				if mqttBinding, ok := binding.(*mqtt.MQTTBinding); ok {
					for _, v := range h.Vendors {
						// set MQTT binding to VendorContext
						v.Vendor.GetContext().SetMQTTBinding(mqttBinding)
					}
				}
			}
		}
	}
	SetHostReadyFlag(true)
	return h.WaitForShutdown(&wg, cancel)
}

func (h *APIHost) WaitForShutdown(wg *sync.WaitGroup, cancel context.CancelFunc) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Debug("received interrupt signal, shutting down...")
		cancel()

		<-sigCh
		log.Debug("received second interrupt signal, shutting down immediately...")
		os.Exit(1)
	}()

	wg.Wait() // Wait for all original goroutines to finish

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), h.ShutdownGracePeriod)
	defer shutdownCancel()

	eg := errgroup.Group{}
	terminables := make([]v1alpha2.Terminable, 0, len(h.Vendors)+len(h.Bindings))
	for _, v := range h.Vendors {
		terminables = append(terminables, v.Vendor)
	}
	for _, b := range h.Bindings {
		terminables = append(terminables, b)
	}

	log.Debug("waiting for services to shutdown...")

	for _, t := range terminables {
		terminable := t
		eg.Go(func() error {
			return terminable.Shutdown(shutdownCtx)
		})
	}
	return eg.Wait()
}

func (h *APIHost) launchHTTP(config interface{}, endpoints []v1alpha2.Endpoint, pubsubProvider pubsub.IPubSubProvider) (bindings.IBinding, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	httpConfig := http.HttpBindingConfig{}
	err = json.Unmarshal(data, &httpConfig)
	if err != nil {
		return nil, err
	}
	binding := &http.HttpBinding{}
	return binding, binding.Launch(httpConfig, endpoints, pubsubProvider)
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
	binding := &mqtt.MQTTBinding{}
	return binding, binding.Launch(mqttConfig, endpoints)
}
