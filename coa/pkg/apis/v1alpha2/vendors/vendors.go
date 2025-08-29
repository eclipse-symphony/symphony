/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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
	v1alpha2.Terminable
	RunLoop(ctx context.Context, interval time.Duration) error
	Init(config VendorConfig, managers []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error
	GetEndpoints() []v1alpha2.Endpoint
	GetInfo() VendorInfo
	SetEvaluationContext(context *utils.EvaluationContext)
	GetContext() *contexts.VendorContext
	SetContext(context *contexts.VendorContext)
	GetManagers() []managers.IManager
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

func (v *Vendor) RunLoop(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			for _, m := range v.Managers {
				if c, ok := m.(managers.ISchedulable); ok {
					if c.Enabled() {
						c.Poll()     //TODO: report errors
						c.Reconcil() //TODO: report errors
					}
				}
			}
		}
	}
}

func (v *Vendor) Shutdown(ctx context.Context) error {
	for _, m := range v.Managers {
		err := m.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Vendor) Init(config VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	v.Context = &contexts.VendorContext{}
	v.Context.SiteInfo = config.SiteInfo
	// see issue #79 - the following needs to be updated to use Symphony expression
	v.Context.SiteInfo.CurrentSite.BaseUrl = utils.ParseProperty(v.Context.SiteInfo.CurrentSite.BaseUrl)
	v.Context.SiteInfo.CurrentSite.Username = utils.ParseProperty(v.Context.SiteInfo.CurrentSite.Username)
	v.Context.SiteInfo.CurrentSite.Password = utils.ParseProperty(v.Context.SiteInfo.CurrentSite.Password)

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

func (v *Vendor) GetContext() *contexts.VendorContext {
	return v.Context
}

func (v *Vendor) SetContext(context *contexts.VendorContext) {

	v.Context = context
}

func (v *Vendor) GetManagers() []managers.IManager {
	return v.Managers
}
