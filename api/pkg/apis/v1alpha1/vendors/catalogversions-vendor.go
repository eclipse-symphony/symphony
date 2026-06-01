/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogversions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
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

var lLog = logger.NewLogger("coa.runtime")

type CatalogVersionsVendor struct {
	vendors.Vendor
	CatalogVersionsManager *catalogversions.CatalogVersionsManager
}

func (e *CatalogVersionsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  e.Vendor.Version,
		Name:     "CatalogVersions",
		Producer: "Microsoft",
	}
}
func (e *CatalogVersionsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*catalogversions.CatalogVersionsManager); ok {
			e.CatalogVersionsManager = c
		}
	}
	if e.CatalogVersionsManager == nil {
		return v1alpha2.NewCOAError(nil, "catalogversions manager is not supplied", v1alpha2.MissingConfig)
	}
	e.Vendor.Context.Subscribe("catalogversion-sync", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			jData, _ := json.Marshal(event.Body)
			var job v1alpha2.JobData
			err := utils2.UnmarshalJson(jData, &job)
			if err == nil {
				var catalogversion model.CatalogVersionState
				jData, _ = json.Marshal(job.Body)
				err = utils2.UnmarshalJson(jData, &catalogversion)
				origin := event.Metadata["origin"]
				if err == nil {
					name := fmt.Sprintf("%s-%s", origin, catalogversion.ObjectMeta.Name)
					catalogversion.ObjectMeta.Name = name
					catalogversion.Spec.RootResource = validation.GetRootResourceFromName(name)
					if catalogversion.Spec.ParentName != "" {
						catalogversion.Spec.ParentName = fmt.Sprintf("%s-%s", origin, catalogversion.Spec.ParentName)
					}
					ctx := context.TODO()
					if event.Context != nil {
						ctx = event.Context
					}
					err := e.CatalogVersionsManager.UpsertState(ctx, name, catalogversion)
					if err != nil {
						return err
					}
				} else {
					iLog.Errorf("Failed to unmarshal job body: %v", err)
					return v1alpha2.NewCOAError(err, "failed to unmarshal job body", v1alpha2.BadConfig)
				}
			} else {
				iLog.Errorf("Failed to unmarshal job data: %v", err)
				return v1alpha2.NewCOAError(err, "failed to unmarshal catalogversion state", v1alpha2.BadConfig)
			}
			return nil
		},
	})
	return nil
}
func (e *CatalogVersionsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "catalogversions"
	if e.Route != "" {
		route = e.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route + "/registry",
			Version:    e.Version,
			Handler:    e.onCatalogVersions,
			Parameters: []string{"name?"},
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/graph",
			Version: e.Version,
			Handler: e.onCatalogVersionsGraph,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/check",
			Version: e.Version,
			Handler: e.onCheck,
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/status",
			Version:    e.Version,
			Handler:    e.onStatus,
			Parameters: []string{"name"},
		},
	}
}
func (e *CatalogVersionsVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("CatalogVersions Vendor", request.Context, &map[string]string{
		"method": "onStatus",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (CatalogVersions Vendor): onStatus, method: %s", string(request.Method))

	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = ""
	}

	switch request.Method {
	case fasthttp.MethodPost:
		var components []model.ComponentSpec
		err := utils2.UnmarshalJson(request.Body, &components)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		id := request.Parameters["__name"]
		if id == "" {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("missing catalogversion name"),
			})
		}
		existingCatalogVersion, err := e.CatalogVersionsManager.GetState(rCtx, id, namespace)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		existingCatalogVersion.Spec.Properties["reported"] = components
		err = e.CatalogVersionsManager.UpsertState(rCtx, id, existingCatalogVersion)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
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
func (e *CatalogVersionsVendor) onCheck(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("CatalogVersions Vendor", request.Context, &map[string]string{
		"method": "onCheck",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (CatalogVersions Vendor): onCheck, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodPost:
		var catalogversion model.CatalogVersionState

		err := utils2.UnmarshalJson(request.Body, &catalogversion)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		errorFields := e.CatalogVersionsManager.CatalogVersionValidator.ValidateCreateOrUpdate(rCtx, catalogversion, nil)
		if len(errorFields) > 0 {
			errorMessage := validation.ConvertErrorFieldsToString(errorFields)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte(errorMessage),
				ContentType: "text/plain",
			})
		}
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
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
func (e *CatalogVersionsVendor) onCatalogVersionsGraph(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("CatalogVersions Vendor", request.Context, &map[string]string{
		"method": "onCatalogVersionsGraph",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (CatalogVersions Vendor): onCatalogVersionsGraph, method: %s", string(request.Method))
	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = ""
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogVersionsGraph-GET", rCtx, nil)
		template := request.Parameters["template"]
		switch template {
		case "config-chains":
			chains, err := e.CatalogVersionsManager.GetChains(ctx, "config", namespace)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  []byte(err.Error()),
				})
			}
			jData, _ := utils.FormatObject(chains, true, "", "")
			resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.OK,
				Body:        jData,
				ContentType: "application/json",
			})
			return resp
		case "asset-trees":
			trees, err := e.CatalogVersionsManager.GetTrees(ctx, "asset", namespace)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  []byte(err.Error()),
				})
			}
			jData, _ := utils.FormatObject(trees, true, "", "")
			resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.OK,
				Body:        jData,
				ContentType: "application/json",
			})
			return resp
		default:
			resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        []byte("{\"result\": \"400 - unknown template\"}"),
				ContentType: "application/json",
			})
			return resp
		}
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
func (e *CatalogVersionsVendor) onCatalogVersions(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("CatalogVersions Vendor", request.Context, &map[string]string{
		"method": "onCatalogVersions",
	})
	defer span.End()

	lLog.InfofCtx(pCtx, "V (CatalogVersions Vendor): onCatalogVersions, method: %s", string(request.Method))

	id := request.Parameters["__name"]
	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = "default"
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogVersions-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			if !namesapceSupplied {
				namespace = ""
			}
			state, err = e.CatalogVersionsManager.ListState(ctx, namespace, request.Parameters["filterType"], request.Parameters["filterValue"])
			isArray = true
		} else {
			state, err = e.CatalogVersionsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			if !utils.IsNotFound(err) {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  []byte(err.Error()),
				})
			} else {
				errorMsg := fmt.Sprintf("catalogversion '%s' is not found in namespace %s", id, namespace)
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(errorMsg),
				})
			}
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
		ctx, span := observability.StartSpan("onCatalogVersions-POST", pCtx, nil)
		if id == "" {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("missing catalogversion name"),
			})
		}
		var catalogversion model.CatalogVersionState

		err := utils2.UnmarshalJson(request.Body, &catalogversion)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = e.CatalogVersionsManager.UpsertState(ctx, id, catalogversion)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCatalogVersions-DELETE", pCtx, nil)
		err := e.CatalogVersionsManager.DeleteState(ctx, id, namespace)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
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
