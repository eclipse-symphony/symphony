/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaignversions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var cLog = logger.NewLogger("coa.runtime")

type CampaignVersionsVendor struct {
	vendors.Vendor
	CampaignVersionsManager *campaignversions.CampaignVersionsManager
}

func (o *CampaignVersionsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "CampaignVersions",
		Producer: "Microsoft",
	}
}

func (e *CampaignVersionsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*campaignversions.CampaignVersionsManager); ok {
			e.CampaignVersionsManager = c
		}
	}
	if e.CampaignVersionsManager == nil {
		return v1alpha2.NewCOAError(nil, "campaignversions manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *CampaignVersionsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "campaignversions"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onCampaignVersions,
			Parameters: []string{"name?"},
		},
	}
}

func (c *CampaignVersionsVendor) onCampaignVersions(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("CampaignVersions Vendor", request.Context, &map[string]string{
		"method": "onCampaignVersions",
	})
	defer span.End()
	cLog.InfofCtx(pCtx, "V (CampaignVersions): onCampaignVersions, method: %s", string(request.Method))

	id := request.Parameters["__name"]
	namespace, namespaceSupplied := request.Parameters["namespace"]
	if !namespaceSupplied {
		namespace = "default"
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCampaignVersions-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			if !namespaceSupplied {
				namespace = ""
			}
			state, err = c.CampaignVersionsManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.CampaignVersionsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			cLog.InfofCtx(ctx, "V (CampaignVersions): onCampaignVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, isArray, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "text/plain"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onCampaignVersions-POST", pCtx, nil)
		var campaignversion model.CampaignVersionState

		err := utils2.UnmarshalJson(request.Body, &campaignversion)
		if err != nil {
			cLog.ErrorfCtx(ctx, "V (CampaignVersions): onCampaignVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = c.CampaignVersionsManager.UpsertState(ctx, id, campaignversion)
		if err != nil {
			cLog.ErrorfCtx(ctx, "V (CampaignVersions): onCampaignVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCampaignVersions-DELETE", pCtx, nil)
		err := c.CampaignVersionsManager.DeleteState(ctx, id, namespace)
		if err != nil {
			cLog.ErrorfCtx(ctx, "V (CampaignVersions): onCampaignVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	cLog.InfoCtx(pCtx, "V (CampaignVersions): onCampaignVersions failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
