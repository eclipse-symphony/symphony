/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutions"
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

var uLog = logger.NewLogger("coa.runtime")

type SolutionsVendor struct {
	vendors.Vendor
	SolutionsManager *solutions.SolutionsManager
}

func (o *SolutionsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Solutions",
		Producer: "Microsoft",
	}
}

func (e *SolutionsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*solutions.SolutionsManager); ok {
			e.SolutionsManager = c
		}
	}
	if e.SolutionsManager == nil {
		return v1alpha2.NewCOAError(nil, "solutions manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *SolutionsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solutions"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onSolutions,
			Parameters: []string{"name", "version?"},
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route,
			Version: o.Version,
			Handler: o.onSolutionsList,
		},
	}
}

func (c *SolutionsVendor) onSolutions(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Solutions Vendor", request.Context, &map[string]string{
		"method": "onSolutions",
	})
	defer span.End()
	uLog.Infof("V (Solutions): onSolutions, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}

	version := request.Parameters["__version"]
	rootResource := request.Parameters["__name"]
	var id string
	var resourceId string
	if version != "" {
		id = rootResource + "-" + version
		resourceId = rootResource + ":" + version
	} else {
		id = rootResource
		resourceId = rootResource
	}
	uLog.Infof("V (Solutions): onSolutions, id: %s, version: %s", id, version)

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSolutions-GET", pCtx, nil)
		var err error
		var state interface{}

		if version == "latest" {
			state, err = c.SolutionsManager.GetLatestState(ctx, rootResource, namespace)
		} else {
			state, err = c.SolutionsManager.GetState(ctx, id, namespace)
		}

		if err != nil {
			uLog.Infof("V (Solutions): onSolutions Get failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, false, request.Parameters["path"], request.Parameters["doc-type"])
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
		ctx, span := observability.StartSpan("onSolutions-POST", pCtx, nil)

		if version == "" || version == "latest" {
			uLog.Infof("V (Solutions): onSolutions Post failed - version is required for POST, traceId: %s", span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte("version is required for POST"),
			})
		}

		embed_type := request.Parameters["embed-type"]
		embed_component := request.Parameters["embed-component"]
		embed_property := request.Parameters["embed-property"]

		var solution model.SolutionState

		if embed_type != "" && embed_component != "" && embed_property != "" {
			solution = model.SolutionState{
				ObjectMeta: model.ObjectMeta{
					Name:      id,
					Namespace: namespace,
				},
				Spec: &model.SolutionSpec{
					DisplayName: id,
					Components: []model.ComponentSpec{
						{
							Name: embed_component,
							Type: embed_type,
							Properties: map[string]interface{}{
								embed_property: string(request.Body),
							},
						},
					},
					Version:      version,
					RootResource: rootResource,
				},
			}
		} else {
			err := json.Unmarshal(request.Body, &solution)
			if err != nil {
				uLog.Infof("V (Solutions): onSolutions Post failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			if solution.ObjectMeta.Name == "" {
				solution.ObjectMeta.Name = id
			}
			if solution.Spec.Version == "" && version != "" {
				solution.Spec.Version = version
			}
			if solution.Spec.RootResource == "" && rootResource != "" {
				solution.Spec.RootResource = rootResource
			}
		}
		err := c.SolutionsManager.UpsertState(ctx, id, solution)
		if err != nil {
			uLog.Infof("V (Solutions): onSolutions Post failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		// TODO: this is a PoC of publishing trails when an object is updated
		strCat := ""
		if solution.Spec.Metadata != nil {
			if v, ok := solution.Spec.Metadata["catalog"]; ok {
				strCat = v
			}
		}
		c.Vendor.Context.Publish("trail", v1alpha2.Event{
			Body: []v1alpha2.Trail{
				{
					Origin:  c.Vendor.Context.SiteInfo.SiteId,
					Catalog: strCat,
					Type:    "solutions.solution.symphony/v1",
					Properties: map[string]interface{}{
						"spec": solution,
					},
				},
			},
			Metadata: map[string]string{
				"namespace": namespace,
			},
		})
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onSolutions-DELETE", pCtx, nil)
		err := c.SolutionsManager.DeleteState(ctx, resourceId, namespace)
		if err != nil {
			uLog.Infof("V (Solutions): onSolutions Delete failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	uLog.Infof("V (Solutions): onSolutions failed - 405 method not allowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *SolutionsVendor) onSolutionsList(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Solutions Vendor", request.Context, &map[string]string{
		"method": "onSolutionsList",
	})
	defer span.End()
	uLog.Infof("V (Solutions): onSolutionsList, method: %s, traceId: %s", request.Method, span.SpanContext().TraceID().String())
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = "default"
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSolutionsList-GET", pCtx, nil)

		var err error
		var state interface{}
		if !exist {
			namespace = ""
		}
		state, err = c.SolutionsManager.ListState(ctx, namespace)

		if err != nil {
			uLog.Infof("V (Solutions): onSolutionsList failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, true, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "application/text"
		}
		return resp
	}

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
