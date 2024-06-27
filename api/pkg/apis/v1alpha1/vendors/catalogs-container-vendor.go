/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogcontainers"
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

var ctLog = logger.NewLogger("coa.runtime")

type CatalogContainersVendor struct {
	vendors.Vendor
	CatalogContainersManager *catalogcontainers.CatalogContainersManager
}

func (o *CatalogContainersVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "CatalogContainers",
		Producer: "Microsoft",
	}
}

func (e *CatalogContainersVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*catalogcontainers.CatalogContainersManager); ok {
			e.CatalogContainersManager = c
		}
	}
	if e.CatalogContainersManager == nil {
		return v1alpha2.NewCOAError(nil, "Catalog container manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *CatalogContainersVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "catalogcontainers"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onCatalogContainers,
			Parameters: []string{"name?"},
		},
	}
}

func (c *CatalogContainersVendor) onCatalogContainers(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("onCatalogContainers", request.Context, &map[string]string{
		"method": "onCatalogContainers",
	})
	defer span.End()
	ctLog.Infof("V (CatalogContainers): onCatalogContainers, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	ctLog.Infof("V (CatalogContainers): onCatalogContainers, method: %s, traceId: %s", string(request.Method), span.SpanContext().TraceID().String())
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogContainers-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.CatalogContainersManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.CatalogContainersManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			ctLog.Errorf("V (CatalogContainers): onCatalogContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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
		ctx, span := observability.StartSpan("onCatalogContainers-POST", pCtx, nil)
		catalog := model.CatalogContainerState{
			ObjectMeta: model.ObjectMeta{
				Name:      id,
				Namespace: namespace,
			},
			Spec: &model.CatalogContainerSpec{},
		}

		err := c.CatalogContainersManager.UpsertState(ctx, id, catalog)
		if err != nil {
			ctLog.Infof("V (CatalogContainers): onCatalogContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCatalogContainers-DELETE", pCtx, nil)
		err := c.CatalogContainersManager.DeleteState(ctx, id, namespace)
		if err != nil {
			ctLog.Infof("V (CatalogContainers): onCatalogContainers failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	ctLog.Infof("V (CatalogContainers): onCatalogContainers failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
