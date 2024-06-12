/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutioncontainers"
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

var scLog = logger.NewLogger("coa.runtime")

type SolutionContainersVendor struct {
	vendors.Vendor
	SolutionContainersManager *solutioncontainers.SolutionContainersManager
}

func (o *SolutionContainersVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "SolutionContainers",
		Producer: "Microsoft",
	}
}

func (e *SolutionContainersVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*solutioncontainers.SolutionContainersManager); ok {
			e.SolutionContainersManager = c
		}
	}
	if e.SolutionContainersManager == nil {
		return v1alpha2.NewCOAError(nil, "solution container manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *SolutionContainersVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solutioncontainers"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onSolutionContainers,
			Parameters: []string{"name?"},
		},
	}
}

func (c *SolutionContainersVendor) onSolutionContainers(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("onSolutionContainers", request.Context, &map[string]string{
		"method": "onSolutionContainers",
	})
	defer span.End()
	scLog.Infof("V (SolutionContainers): onSolutionContainers, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	scLog.Infof("V (SolutionContainers): onSolutionContainers, method: %s, traceId: %s", string(request.Method), span.SpanContext().TraceID().String())
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSolutionContainers-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.SolutionContainersManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.SolutionContainersManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			scLog.Errorf("V (SolutionContainers): onSolutionContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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
		ctx, span := observability.StartSpan("onSolutionContainers-POST", pCtx, nil)
		solution := model.SolutionContainerState{
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Spec: &model.SolutionContainerSpec{},
		}

		err := c.SolutionContainersManager.UpsertState(ctx, id, solution)
		if err != nil {
			scLog.Infof("V (SolutionContainers): onSolutionContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onSolutionContainers-DELETE", pCtx, nil)
		err := c.SolutionContainersManager.DeleteState(ctx, id, namespace)
		if err != nil {
			scLog.Infof("V (SolutionContainers): onSolutionContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	scLog.Infof("V (SolutionContainers): onSolutionContainers failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
