/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targetcontainers"
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

var tcLog = logger.NewLogger("coa.runtime")

type TargetContainersVendor struct {
	vendors.Vendor
	TargetContainersManager *targetcontainers.TargetContainersManager
}

func (o *TargetContainersVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "TargetContainers",
		Producer: "Microsoft",
	}
}

func (e *TargetContainersVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*targetcontainers.TargetContainersManager); ok {
			e.TargetContainersManager = c
		}
	}
	if e.TargetContainersManager == nil {
		return v1alpha2.NewCOAError(nil, "target container manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *TargetContainersVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "targetcontainers"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onTargetContainers,
			Parameters: []string{"name?"},
		},
	}
}

func (c *TargetContainersVendor) onTargetContainers(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("onTargetContainers", request.Context, &map[string]string{
		"method": "onTargetContainers",
	})
	defer span.End()
	tcLog.Infof("V (TargetContainers): onTargetContainers, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	tcLog.Infof("V (TargetContainers): onTargetContainers, method: %s, traceId: %s", string(request.Method), span.SpanContext().TraceID().String())
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onTargetContainers-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.TargetContainersManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.TargetContainersManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			tcLog.Errorf("V (TargetContainers): onTargetContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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
		ctx, span := observability.StartSpan("onTargetContainers-POST", pCtx, nil)
		target := model.TargetContainerState{
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Spec: &model.TargetContainerSpec{},
		}

		err := c.TargetContainersManager.UpsertState(ctx, id, target)
		if err != nil {
			tcLog.Infof("V (TargetContainers): onTargetContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onTargetContainers-DELETE", pCtx, nil)
		err := c.TargetContainersManager.DeleteState(ctx, id, namespace)
		if err != nil {
			tcLog.Infof("V (TargetContainers): onTargetContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	tcLog.Infof("V (TargetContainers): onTargetContainers failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
