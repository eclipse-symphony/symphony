/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

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

var cvLog = logger.NewLogger("coa.runtime")

type VisualizationClientVendor struct {
	vendors.Vendor
	apiClient utils.ApiClient
}

func (s *VisualizationClientVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  s.Vendor.Version,
		Name:     "VisualizationClient",
		Producer: "Microsoft",
	}
}

func (e *VisualizationClientVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	e.apiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}

func (o *VisualizationClientVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "vis-client"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onVisClient,
			Parameters: []string{},
		},
	}
}

func (c *VisualizationClientVendor) onVisClient(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Visualization Client Vendor", request.Context, &map[string]string{
		"method": "onVisClient",
	})
	defer span.End()

	cvLog.InfofCtx(pCtx, "V (VisualizationClient): onVisClient, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onVisClient-POST", pCtx, nil)
		var packet model.Packet
		err := json.Unmarshal(request.Body, &packet)
		if err != nil {
			cvLog.ErrorfCtx(pCtx, "V (VisualizationClient): onVisClient failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		if !packet.IsValid() {
			cvLog.ErrorfCtx(pCtx, "V (VisualizationClient): onVisClient failed - %s", "invalid visualization packet")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("invalid visualization packet"),
			})
		}

		jData, _ := json.Marshal(packet)
		err = c.apiClient.SendVisualizationPacket(ctx, jData,
			c.Vendor.Context.SiteInfo.CurrentSite.Username,
			c.Vendor.Context.SiteInfo.CurrentSite.Password)

		if err != nil {
			cvLog.ErrorfCtx(pCtx, "V (VisualizationClient): onVisClient failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
