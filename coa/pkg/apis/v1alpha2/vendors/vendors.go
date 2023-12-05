/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type VendorConfig struct {
	Type         string                   `json:"type"`
	Route        string                   `json:"route"`
	Managers     []managers.ManagerConfig `json:"managers"`
	Properties   map[string]string        `json:"properties,omitempty"`
	LoopInterval int                      `json:"loopInterval,omitempty"`
	SiteInfo     v1alpha2.SiteInfo        `json:"siteInfo"`
}

type IVendor interface {
	RunLoop(interval time.Duration) error
	Init(config VendorConfig, managers []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error
	GetEndpoints() []v1alpha2.Endpoint
	GetInfo() VendorInfo
	SetEvaluationContext(context *utils.EvaluationContext)
}

type IEvaluationContextVendor interface {
	GetEvaluationContext() *utils.EvaluationContext
}

type IVendorFactory interface {
	CreateVendor(config VendorConfig) (IVendor, error)
}

type VendorInfo struct {
	Version  string `json:"version"`
	Name     string `json:"name"`
	Producer string `json:"producer"`
}

type Vendor struct {
	Managers []managers.IManager
	Version  string
	Route    string
	Context  *contexts.VendorContext
	Config   VendorConfig
}

func (v *Vendor) SetEvaluationContext(context *utils.EvaluationContext) {
	v.Context.EvaluationContext = context
}
func (v *Vendor) RunLoop(interval time.Duration) error {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		fmt.Println("CLEANING UP")
		os.Exit(0)
	}()
	for true {
		for _, m := range v.Managers {
			if c, ok := m.(managers.ISchedulable); ok {
				if c.Enabled() {
					c.Poll()     //TODO: report errors
					c.Reconcil() //TODO: report errors
				}
			}
		}
		time.Sleep(interval)
	}
	return nil
}

func (v *Vendor) Init(config VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	v.Context = &contexts.VendorContext{}
	v.Context.SiteInfo = config.SiteInfo
	err := v.Context.Init(pubsubProvider)
	if err != nil {
		return err
	}
	v.Context.Logger.Debugf("V (%s): initialize at route '%s'", config.Type, config.Route)
	v.Managers = []managers.IManager{}
	for _, m := range config.Managers {
		created := false
		for _, factory := range factories {
			manager, err := factory.CreateManager(m)
			if err != nil {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to create manager '%s'", m.Name), v1alpha2.InternalError)
			}
			if manager != nil {
				mp, ok := providers[m.Name]
				if !ok {
					err = manager.Init(v.Context, m, nil)
				} else {
					err = manager.Init(v.Context, m, mp)
				}
				if err != nil {
					return err
				}
				v.Managers = append(v.Managers, manager)
				created = true
				break
			}
		}
		if !created {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("no manager factories can create manager type '%s'", m.Type), v1alpha2.BadConfig)
		}
	}
	v.Version = "v1alpha2"
	v.Route = config.Route
	v.Config = config
	return nil
}
