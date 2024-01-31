/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/models"
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

var mLog = logger.NewLogger("coa.runtime")

type ModelsVendor struct {
	vendors.Vendor
	ModelsManager *models.ModelsManager
}

func (o *ModelsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Models",
		Producer: "Microsoft",
	}
}

func (e *ModelsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*models.ModelsManager); ok {
			e.ModelsManager = c
		}
	}
	if e.ModelsManager == nil {
		return v1alpha2.NewCOAError(nil, "models manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *ModelsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "models"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onModels,
			Parameters: []string{"name?"},
		},
	}
}

func (c *ModelsVendor) onModels(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Models Vendor", request.Context, &map[string]string{
		"method": "onModels",
	})
	defer span.End()
	tLog.Debugf("V (Models): onModels, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onModels-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = c.ModelsManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = c.ModelsManager.GetSpec(ctx, id)
		}
		if err != nil {
			if isArray {
				tLog.Errorf(" V (Models): onModels failed to ListSpec, err: %v, traceId: %s", err, span.SpanContext().TraceID().String())
			} else {
				tLog.Errorf(" V (Models): onModels failed to GetSpec, id: %s, err: %v, traceId: %s", id, err, span.SpanContext().TraceID().String())
			}
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
		ctx, span := observability.StartSpan("onModels-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var model model.ModelSpec

		err := json.Unmarshal(request.Body, &model)
		if err != nil {
			tLog.Errorf("V (Models): onModels failed to pause model from request body, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = c.ModelsManager.UpsertSpec(ctx, id, model)
		if err != nil {
			tLog.Errorf("V (Models): onModels failed to UpsertSpec, id: %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onModels-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := c.ModelsManager.DeleteSpec(ctx, id)
		if err != nil {
			tLog.Errorf("V (Models): onModels failed to DeleteSpec, id: %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	tLog.Errorf("V (Models): onModels returned MethodNotAllowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
