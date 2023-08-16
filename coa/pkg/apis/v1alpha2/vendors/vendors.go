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
	Site         string                   `json:"site,omitempty"`
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
	v.Context.Site = config.Site
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
