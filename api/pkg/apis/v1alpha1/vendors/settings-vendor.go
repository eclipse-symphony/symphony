/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"
	"strings"

	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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
}

func (c *SettingsVendor) onConfig(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Settings Vendor", request.Context, &map[string]string{
		"method": "onConfig",
	})
	defer span.End()
	csLog.Infof("V (Settings): onConfig %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

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
				log.Errorf("V (Settings): onConfig failed to get config %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
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
				log.Errorf("V (Settings): onConfig failed to get object %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
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

	log.Infof("V (Settings): onConfig returned MethodNotAllowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
