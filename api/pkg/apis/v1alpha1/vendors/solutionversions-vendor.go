/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutionversions"
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

var uLog = logger.NewLogger("coa.runtime")

type SolutionVersionsVendor struct {
	vendors.Vendor
	SolutionVersionsManager *solutionversions.SolutionVersionsManager
}

func (o *SolutionVersionsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "SolutionVersions",
		Producer: "Microsoft",
	}
}

func (e *SolutionVersionsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*solutionversions.SolutionVersionsManager); ok {
			e.SolutionVersionsManager = c
		}
	}
	if e.SolutionVersionsManager == nil {
		return v1alpha2.NewCOAError(nil, "solutionversions manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *SolutionVersionsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "solutionversions"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onSolutionVersions,
			Parameters: []string{"name?"},
		},
	}
}

func (c *SolutionVersionsVendor) onSolutionVersions(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("SolutionVersions Vendor", request.Context, &map[string]string{
		"method": "onSolutionVersions",
	})
	defer span.End()
	uLog.InfofCtx(pCtx, "V (SolutionVersions): onSolutionVersions, method: %s", request.Method)

	id := request.Parameters["__name"]
	namespace, exist := request.Parameters["namespace"]
	if !exist {
		namespace = constants.DefaultScope
	}
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSolutionVersions-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			// Change namespace back to empty to indicate ListSpec need to query all namespaces
			if !exist {
				namespace = ""
			}
			state, err = c.SolutionVersionsManager.ListState(ctx, namespace)
			isArray = true
		} else {
			state, err = c.SolutionVersionsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			uLog.ErrorfCtx(ctx, "V (SolutionVersions): onSolutionVersions failed - %s", err.Error())
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
		ctx, span := observability.StartSpan("onSolutionVersions-POST", pCtx, nil)
		embed_type := request.Parameters["embed-type"]
		embed_component := request.Parameters["embed-component"]
		embed_property := request.Parameters["embed-property"]

		var solutionversion model.SolutionVersionState

		if embed_type != "" && embed_component != "" && embed_property != "" {
			solutionversion = model.SolutionVersionState{
				ObjectMeta: model.ObjectMeta{
					Name:      id,
					Namespace: namespace,
				},
				Spec: &model.SolutionVersionSpec{
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
				},
			}
		} else {
			err := utils2.UnmarshalJson(request.Body, &solutionversion)
			if err != nil {
				uLog.ErrorfCtx(ctx, "V (SolutionVersions): onSolutionVersions failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
			if solutionversion.ObjectMeta.Name == "" {
				solutionversion.ObjectMeta.Name = id
			}
		}
		err := c.SolutionVersionsManager.UpsertState(ctx, id, solutionversion)
		if err != nil {
			uLog.ErrorfCtx(ctx, "V (SolutionVersions): onSolutionVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		// TODO: this is a PoC of publishing trails when an object is updated
		strCat := ""
		if solutionversion.Spec.Metadata != nil {
			if v, ok := solutionversion.Spec.Metadata["catalogversion"]; ok {
				strCat = v
			}
		}
		c.Vendor.Context.Publish("trail", v1alpha2.Event{
			Body: []v1alpha2.Trail{
				{
					Origin:  c.Vendor.Context.SiteInfo.SiteId,
					CatalogVersion: strCat,
					Type:    "solutionversions.solutionversion.symphony/v1",
					Properties: map[string]interface{}{
						"spec": solutionversion,
					},
				},
			},
			Metadata: map[string]string{
				"namespace": namespace,
			},
			Context: ctx,
		})
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onSolutionVersions-DELETE", pCtx, nil)
		err := c.SolutionVersionsManager.DeleteState(ctx, id, namespace)
		if err != nil {
			uLog.ErrorfCtx(ctx, "V (SolutionVersions): onSolutionVersions failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	uLog.ErrorCtx(pCtx, "V (SolutionVersions): onSolutionVersions failed - 405 method not allowed")
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
