/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/campaigncontainers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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

var ccLog = logger.NewLogger("coa.runtime")

type CampaignContainersVendor struct {
	vendors.Vendor
	CampaignContainersManager *campaigncontainers.CampaignContainersManager
}

func (o *CampaignContainersVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "CampaignContainers",
		Producer: "Microsoft",
	}
}

func (e *CampaignContainersVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*campaigncontainers.CampaignContainersManager); ok {
			e.CampaignContainersManager = c
		}
	}
	if e.CampaignContainersManager == nil {
		return v1alpha2.NewCOAError(nil, "Campaign container manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *CampaignContainersVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "campaigncontainers"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onCampaignContainers,
			Parameters: []string{"name?"},
		},
	}
}

func (c *CampaignContainersVendor) onCampaignContainers(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("onCampaignContainers", request.Context, &map[string]string{
		"method": "onCampaignContainers",
	})
	defer span.End()
	ccLog.Infof("V (CampaignContainers): onCampaignContainers, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	ccLog.Infof("V (CampaignContainers): onCampaignContainers, method: %s, traceId: %s", string(request.Method), span.SpanContext().TraceID().String())
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCampaignContainers-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.CampaignContainersManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.CampaignContainersManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			ccLog.Errorf("V (CampaignContainers): onCampaignContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
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
			resp.ContentType = "application/text"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onCampaignContainers-POST", pCtx, nil)
		campaign := model.CampaignContainerState{
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Spec: &model.CampaignContainerSpec{},
		}

		err := c.CampaignContainersManager.UpsertState(ctx, id, campaign)
		if err != nil {
			ccLog.Infof("V (CampaignContainers): onCampaignContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCampaignContainers-DELETE", pCtx, nil)
		err := c.CampaignContainersManager.DeleteState(ctx, id, namespace)
		if err != nil {
			ccLog.Infof("V (CampaignContainers): onCampaignContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	ccLog.Infof("V (CampaignContainers): onCampaignContainers failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
