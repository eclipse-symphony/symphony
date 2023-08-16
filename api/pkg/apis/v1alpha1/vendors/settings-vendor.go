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
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type SettingsVendor struct {
	vendors.Vendor
	EvaluationContext *utils.EvaluationContext
}

func (e *SettingsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  e.Vendor.Version,
		Name:     "Settings",
		Producer: "Microsoft",
	}
}
func (e *SettingsVendor) Init(cfg vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(cfg, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	var configProvider config.IExtConfigProvider
	for _, m := range e.Managers {
		if c, ok := m.(config.IExtConfigProvider); ok {
			configProvider = c
		}
	}
	e.EvaluationContext = &utils.EvaluationContext{
		ConfigProvider: configProvider,
	}
	return nil
}
func (e *SettingsVendor) GetEvaluationContext() *utils.EvaluationContext {
	return e.EvaluationContext
}
func (o *SettingsVendor) GetEndpoints() []v1alpha2.Endpoint {
	return nil
}
