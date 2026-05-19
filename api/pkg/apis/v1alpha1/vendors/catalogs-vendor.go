/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
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

var ctLog = logger.NewLogger("coa.runtime")

type CatalogsVendor struct {
	vendors.Vendor
	CatalogsManager *catalogs.CatalogsManager
}

func (o *CatalogsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Catalogs",
		Producer: "Microsoft",
	}
}

func (e *CatalogsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*catalogs.CatalogsManager); ok {
			e.CatalogsManager = c
		}
	}
	if e.CatalogsManager == nil {
		return v1alpha2.NewCOAError(nil, "Catalog container manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *CatalogsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "catalogs"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onCatalogs,
			Parameters: []string{"name?"},
		},
	}
}

func (c *CatalogsVendor) onCatalogs(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("onCatalogs", request.Context, &map[string]string{
		"method": "onCatalogs",
	})
	defer span.End()
	ctLog.InfofCtx(pCtx, "V (Catalogs): onCatalogs, method: %s", request.Method)

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	ctLog.InfofCtx(pCtx, "V (Catalogs): onCatalogs, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogs-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change partition back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.CatalogsManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.CatalogsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			ctLog.ErrorfCtx(ctx, "V (Catalogs): onCatalogs failed - %s", err.Error())
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
		ctx, span := observability.StartSpan("onCatalogs-POST", pCtx, nil)
		var catalog model.CatalogState
		err := utils2.UnmarshalJson(request.Body, &catalog)
		if err != nil {
			ctLog.ErrorfCtx(ctx, "V (Catalogs): onCatalogs failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		catalog.ObjectMeta.Name = id
		catalog.ObjectMeta.Namespace = namespace
		catalog.Spec = &model.CatalogSpec{}

		err = c.CatalogsManager.UpsertState(ctx, id, catalog)
		if err != nil {
			ctLog.ErrorfCtx(ctx, "V (Catalogs): onCatalogs failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCatalogs-DELETE", pCtx, nil)
		err := c.CatalogsManager.DeleteState(ctx, id, namespace)
		if err != nil {
			ctLog.ErrorfCtx(ctx, "V (Catalogs): onCatalogs failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	ctLog.InfoCtx(pCtx, "V (Catalogs): onCatalogs failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
