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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
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

var lLog = logger.NewLogger("coa.runtime")

type CatalogsVendor struct {
	vendors.Vendor
	CatalogsManager *catalogs.CatalogsManager
}

func (e *CatalogsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  e.Vendor.Version,
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
		return v1alpha2.NewCOAError(nil, "catalogs manager is not supplied", v1alpha2.MissingConfig)
	}
	e.Vendor.Context.Subscribe("catalog-sync", func(topic string, event v1alpha2.Event) error {
		jData, _ := json.Marshal(event.Body)
		var job v1alpha2.JobData
		err := json.Unmarshal(jData, &job)
		if err == nil {
			var catalog model.CatalogState
			jData, _ = json.Marshal(job.Body)
			err = json.Unmarshal(jData, &catalog)
			origin := event.Metadata["origin"]
			if err == nil {
				name := fmt.Sprintf("%s-%s", origin, catalog.ObjectMeta.Name)
				catalog.ObjectMeta.Name = name
				catalog.Spec.RootResource = validation.GetRootResourceFromName(name)
				if catalog.Spec.ParentName != "" {
					catalog.Spec.ParentName = fmt.Sprintf("%s-%s", origin, catalog.Spec.ParentName)
				}
				ctx := context.TODO()
				if event.Context != nil {
					ctx = event.Context
				}
				err := e.CatalogsManager.UpsertState(ctx, name, catalog)
				if err != nil {
					return err
				}
			} else {
				iLog.Errorf("Failed to unmarshal job body: %v", err)
				return v1alpha2.NewCOAError(err, "failed to unmarshal job body", v1alpha2.BadConfig)
			}
		} else {
			iLog.Errorf("Failed to unmarshal job data: %v", err)
			return v1alpha2.NewCOAError(err, "failed to unmarshal catalog state", v1alpha2.BadConfig)
		}
		return nil
	})
	return nil
}
func (e *CatalogsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "catalogs"
	if e.Route != "" {
		route = e.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route + "/registry",
			Version:    e.Version,
			Handler:    e.onCatalogs,
			Parameters: []string{"name?"},
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   route + "/graph",
			Version: e.Version,
			Handler: e.onCatalogsGraph,
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
func (e *CatalogsVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onStatus",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (Catalogs Vendor): onStatus, method: %s", string(request.Method))

	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = ""
	}

	switch request.Method {
	case fasthttp.MethodPost:
		var components []model.ComponentSpec
		err := json.Unmarshal(request.Body, &components)
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
				Body:  []byte("missing catalog name"),
			})
		}
		existingCatalog, err := e.CatalogsManager.GetState(rCtx, id, namespace)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		existingCatalog.Spec.Properties["reported"] = components
		err = e.CatalogsManager.UpsertState(rCtx, id, existingCatalog)
		if err != nil {
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
func (e *CatalogsVendor) onCheck(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCheck",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (Catalogs Vendor): onCheck, method: %s", string(request.Method))
	switch request.Method {
	case fasthttp.MethodPost:
		var catalog model.CatalogState

		err := json.Unmarshal(request.Body, &catalog)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		errorFields := e.CatalogsManager.CatalogValidator.ValidateCreateOrUpdate(rCtx, catalog, nil)
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
func (e *CatalogsVendor) onCatalogsGraph(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCatalogsGraph",
	})
	defer span.End()

	lLog.InfofCtx(rCtx, "V (Catalogs Vendor): onCatalogsGraph, method: %s", string(request.Method))
	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = ""
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogsGraph-GET", rCtx, nil)
		template := request.Parameters["template"]
		switch template {
		case "config-chains":
			chains, err := e.CatalogsManager.GetChains(ctx, "config", namespace)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
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
			trees, err := e.CatalogsManager.GetTrees(ctx, "asset", namespace)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
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
func (e *CatalogsVendor) onCatalogs(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCatalogs",
	})
	defer span.End()

	lLog.InfofCtx(pCtx, "V (Catalogs Vendor): onCatalogs, method: %s", string(request.Method))

	id := request.Parameters["__name"]
	namespace, namesapceSupplied := request.Parameters["namespace"]
	if !namesapceSupplied {
		namespace = "default"
	}

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogs-GET", pCtx, nil)
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			if !namesapceSupplied {
				namespace = ""
			}
			state, err = e.CatalogsManager.ListState(ctx, namespace, request.Parameters["filterType"], request.Parameters["filterValue"])
			isArray = true
		} else {
			state, err = e.CatalogsManager.GetState(ctx, id, namespace)
		}
		if err != nil {
			if !v1alpha2.IsNotFound(err) {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(err.Error()),
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
			resp.ContentType = "application/text"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onCatalogs-POST", pCtx, nil)
		if id == "" {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("missing catalog name"),
			})
		}
		var catalog model.CatalogState

		err := json.Unmarshal(request.Body, &catalog)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = e.CatalogsManager.UpsertState(ctx, id, catalog)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onCatalogs-DELETE", pCtx, nil)
		err := e.CatalogsManager.DeleteState(ctx, id, namespace)
		if err != nil {
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
