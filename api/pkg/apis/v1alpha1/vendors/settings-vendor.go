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
	"encoding/json"
	"strings"

	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var csLog = logger.NewLogger("coa.runtime")

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
	route := "settings"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet},
			Route:      route + "/config",
			Version:    o.Version,
			Handler:    o.onConfig,
			Parameters: []string{"name?"},
		},
	}
	return nil
}

func (c *SettingsVendor) onConfig(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Settings Vendor", request.Context, &map[string]string{
		"method": "onCampaigns",
	})
	csLog.Info("V (Settings): onConfig")
	switch request.Method {
	case fasthttp.MethodGet:
		id := request.Parameters["__name"]
		overrides := request.Parameters["overrides"]
		field := request.Parameters["field"]
		var parts []string
		if overrides != "" {
			parts = strings.Split(overrides, ",")
		}
		if field != "" {
			val, err := c.EvaluationContext.ConfigProvider.Get(id, field, parts, nil)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			data, _ := json.Marshal(val)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.OK,
				Body:        data,
				ContentType: "text/plain",
			})
		} else {
			val, err := c.EvaluationContext.ConfigProvider.GetObject(id, parts, nil)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			jData, _ := api_utils.FormatObject(val, false, "", "")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.OK,
				Body:        jData,
				ContentType: "application/json",
			})
		}
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	span.End()
	return resp
}
