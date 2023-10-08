/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package vendors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/azure/symphony/coa/pkg/logger"
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
			if catalog, ok := job.Body.(model.CatalogSpec); ok {
				name := fmt.Sprintf("%s-%s", catalog.SiteId, catalog.Name)
				catalog.Name = name
				if catalog.ParentName != "" {
					catalog.ParentName = fmt.Sprintf("%s-%s", catalog.SiteId, catalog.ParentName)
				}
				err := e.CatalogsManager.UpsertSpec(context.Background(), name, catalog)
				if err != nil {
					return v1alpha2.NewCOAError(err, "failed to upsert catalog", v1alpha2.InternalError)
				}
			}
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
	}
}
func (e *CatalogsVendor) onCheck(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCheck",
	})
	lLog.Info("V (Catalogs Vendor): onCheck")
	switch request.Method {
	case fasthttp.MethodPost:
		var campaign model.CatalogSpec

		err := json.Unmarshal(request.Body, &campaign)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		res, err := e.CatalogsManager.ValidateSpec(rCtx, campaign)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		if !res.Valid {
			jData, _ := utils.FormatObject(res.Errors, true, "", "")
			resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				Body:        jData,
				ContentType: "application/json",
			})
			return resp
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
	span.End()
	return resp
}
func (e *CatalogsVendor) onCatalogsGraph(request v1alpha2.COARequest) v1alpha2.COAResponse {
	rCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCatalogsGraph",
	})
	lLog.Info("V (Catalogs Vendor): onCatalogsGraph")
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogsGraph-GET", rCtx, nil)
		template := request.Parameters["template"]
		switch template {
		case "config-chains":
			chains, err := e.CatalogsManager.GetChains(ctx, "config")
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
			trees, err := e.CatalogsManager.GetTrees(ctx, "asset")
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
	span.End()
	return resp
}
func (e *CatalogsVendor) onCatalogs(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Catalogs Vendor", request.Context, &map[string]string{
		"method": "onCatalogs",
	})
	lLog.Info("V (Catalogs Vendor): onCatalogs")
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onCatalogs-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = e.CatalogsManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = e.CatalogsManager.GetSpec(ctx, id)
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
		id := request.Parameters["__name"]
		if id == "" {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("missing catalog name"),
			})
		}
		var campaign model.CatalogSpec

		err := json.Unmarshal(request.Body, &campaign)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = e.CatalogsManager.UpsertSpec(ctx, id, campaign)
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
		id := request.Parameters["__name"]
		err := e.CatalogsManager.DeleteSpec(ctx, id)
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
	span.End()
	return resp
}
