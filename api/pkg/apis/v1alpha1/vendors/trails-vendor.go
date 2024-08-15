/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/trails"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var trLog = logger.NewLogger("coa.runtime")

type TrailsVendor struct {
	vendors.Vendor
	TrailsManager *trails.TrailsManager
}

func (o *TrailsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Trails",
		Producer: "Microsoft",
	}
}

func (e *TrailsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*trails.TrailsManager); ok {
			e.TrailsManager = c
		}
	}
	if e.TrailsManager == nil {
		return v1alpha2.NewCOAError(nil, "trails manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *TrailsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "trails"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route,
			Version: o.Version,
			Handler: o.onTrails,
		},
	}
}

func (c *TrailsVendor) onTrails(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Trails Vendor", request.Context, &map[string]string{
		"method": "onTrails",
	})
	defer span.End()
	tLog.InfofCtx(pCtx, "V (Trails) : onTrails %s", request.Method)

	switch request.Method {
	case fasthttp.MethodPost:
		var trails []v1alpha2.Trail
		err := json.Unmarshal(request.Body, &trails)
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Trails): onTrails failed to parse trails from request body, error: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		err = c.TrailsManager.Append(pCtx, trails)
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Trails): onTrails failed to Append, error: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
			Body:  []byte("{\"result\":\"ok\"}"),
		})
	}
	tLog.ErrorCtx(pCtx, "V (Trails): onTrails returned MethodNotAllowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
