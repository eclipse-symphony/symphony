/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/devices"
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

var dLog = logger.NewLogger("coa.runtime")

type DevicesVendor struct {
	vendors.Vendor
	DevicesManager *devices.DevicesManager
}

func (o *DevicesVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Devices",
		Producer: "Microsoft",
	}
}

func (e *DevicesVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*devices.DevicesManager); ok {
			e.DevicesManager = c
		}
	}
	if e.DevicesManager == nil {
		return v1alpha2.NewCOAError(nil, "devices manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *DevicesVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "devices"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onDevices,
			Parameters: []string{"name?"},
		},
	}
}

func (c *DevicesVendor) onDevices(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Devices Vendor", request.Context, &map[string]string{
		"method": "onDevices",
	})
	defer span.End()
	tLog.Infof("V (Devices): onDevices %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onDevices-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = c.DevicesManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = c.DevicesManager.GetSpec(ctx, id)
		}
		if err != nil {
			log.Errorf("V (Devices): failed to get device spec, error %v, traceId: %s", err, span.SpanContext().TraceID().String())
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
		ctx, span := observability.StartSpan("onDevices-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var device model.DeviceSpec

		err := json.Unmarshal(request.Body, &device)
		if err != nil {
			log.Errorf("V (Devices): failed to unmarshall request body, error %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = c.DevicesManager.UpsertSpec(ctx, id, device)
		if err != nil {
			log.Errorf("V (Devices): failed to upsert device spec, error %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onDevices-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := c.DevicesManager.DeleteSpec(ctx, id)
		if err != nil {
			log.Errorf("V (Devices): failed to delete device spec, error %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}

	log.Infof("V (Devices): onDevices returns MethodNotAllowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
