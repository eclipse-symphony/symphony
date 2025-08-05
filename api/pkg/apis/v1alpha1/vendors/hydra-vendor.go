/*
* Copyright (c) Microsoft Corporation.
* Licensed under the MIT license.
* SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/hydra"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

type HydraVendor struct {
	vendors.Vendor
	HydraManager *hydra.HydraManager
}

func (o *HydraVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Hydra",
		Producer: "Microsoft",
	}
}

func (e *HydraVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*hydra.HydraManager); ok {
			e.HydraManager = c
		}
	}
	if e.HydraManager == nil {
		return v1alpha2.NewCOAError(nil, "Hydra manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *HydraVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "hydra"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onHydra,
			Parameters: []string{"system", "objType", "key"},
		},
	}
}

func (c *HydraVendor) onHydra(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Hydra Vendor", request.Context, &map[string]string{
		"method": "onHydra",
	})
	defer span.End()
	tLog.InfofCtx(pCtx, "V (Hydra) : onHydra, method: %s", request.Method)

	system := request.Parameters["__system"]
	objType := request.Parameters["__objType"]
	key := request.Parameters["__key"]
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onHydra-GET", pCtx, nil)
		var artifacts []interface{}
		err := utils2.UnmarshalJson(request.Body, &artifacts)
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Hydra) : onHydra failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		payloads, err := c.HydraManager.GetArtifacts(pCtx, system, objType, key)
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Hydra) : onHydra failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        payloads,
			ContentType: "application/json",
		}
		tLog.InfofCtx(pCtx, "V (Hydra) : onHydra succeeded")
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onHydra-POST", pCtx, nil)
		err := c.HydraManager.SetArtifacts(system, request.Body)
		if err != nil {
			tLog.ErrorfCtx(ctx, "V (Hydra) : onHydra failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		resp := v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        []byte("{\"result\":\"200 - OK\"}"),
			ContentType: "application/json",
		}
		tLog.InfofCtx(pCtx, "V (Hydra) : onHydra succeeded")
		return resp
	}
	tLog.ErrorCtx(pCtx, "V (Hydra) : onHydra failed - method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
