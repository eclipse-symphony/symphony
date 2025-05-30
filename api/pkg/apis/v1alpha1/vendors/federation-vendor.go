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
	"strconv"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sites"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/staging"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/sync"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/trails"
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

var fLog = logger.NewLogger("coa.runtime")

type FederationVendor struct {
	vendors.Vendor
	SitesManager    *sites.SitesManager
	CatalogsManager *catalogs.CatalogsManager
	StagingManager  *staging.StagingManager
	SyncManager     *sync.SyncManager
	TrailsManager   *trails.TrailsManager
	apiClient       utils.ApiClient
}

func (f *FederationVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  f.Vendor.Version,
		Name:     "Federation",
		Producer: "Microsoft",
	}
}
func (f *FederationVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := f.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range f.Managers {
		if c, ok := m.(*sites.SitesManager); ok {
			f.SitesManager = c
		}
		if c, ok := m.(*staging.StagingManager); ok {
			f.StagingManager = c
		}
		if c, ok := m.(*catalogs.CatalogsManager); ok {
			f.CatalogsManager = c
		}
		if c, ok := m.(*sync.SyncManager); ok {
			f.SyncManager = c
		}
		if c, ok := m.(*trails.TrailsManager); ok {
			f.TrailsManager = c
		}
	}
	if f.StagingManager == nil {
		return v1alpha2.NewCOAError(nil, "staging manager is not supplied", v1alpha2.MissingConfig)
	}
	if f.SitesManager == nil {
		return v1alpha2.NewCOAError(nil, "sites manager is not supplied", v1alpha2.MissingConfig)
	}
	if f.CatalogsManager == nil {
		return v1alpha2.NewCOAError(nil, "catalogs manager is not supplied", v1alpha2.MissingConfig)
	}
	f.apiClient, err = utils.GetParentApiClient(f.Vendor.Context.SiteInfo.ParentSite.BaseUrl)
	if err != nil {
		return err
	}
	f.Vendor.Context.Subscribe("catalog", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			sites, err := f.SitesManager.ListState(context.TODO())
			if err != nil {
				return err
			}
			for _, site := range sites {
				if site.Spec.Name != f.Vendor.Context.SiteInfo.SiteId {
					event.Metadata["site"] = site.Spec.Name
					ctx := context.TODO()
					if event.Context != nil {
						ctx = event.Context
					}
					f.StagingManager.HandleJobEvent(ctx, event) //TODO: how to handle errors in this case?
				}
			}
			return nil
		},
	})
	f.Vendor.Context.Subscribe("remote", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			_, ok := event.Metadata["site"]
			if !ok {
				return v1alpha2.NewCOAError(nil, "site is not supplied", v1alpha2.BadRequest)
			}
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			f.StagingManager.HandleJobEvent(ctx, event) //TODO: how to handle errors in this case?
			return nil
		},
	})
	f.Vendor.Context.Subscribe("report", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			fLog.DebugfCtx(ctx, "V (Federation): received report event: %v", event)
			jData, _ := json.Marshal(event.Body)
			var status model.StageStatus
			err := utils2.UnmarshalJson(jData, &status)
			if err == nil {
				ctx := context.TODO()
				if event.Context != nil {
					ctx = event.Context
				}
				err := f.apiClient.SyncStageStatus(ctx, status,
					f.Vendor.Context.SiteInfo.ParentSite.Username,
					f.Vendor.Context.SiteInfo.ParentSite.Password)
				if err != nil {
					fLog.ErrorfCtx(ctx, "V (Federation): error while syncing activation status: %v", err)
					return err
				}
			}
			return v1alpha2.NewCOAError(nil, "report is not an activation status", v1alpha2.BadRequest)
		},
	})
	f.Vendor.Context.Subscribe("trail", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			ctx := context.TODO()
			if event.Context != nil {
				ctx = event.Context
			}
			if f.TrailsManager != nil {
				jData, _ := json.Marshal(event.Body)
				var trails []v1alpha2.Trail
				err := utils2.UnmarshalJson(jData, &trails)
				if err == nil {
					return f.TrailsManager.Append(ctx, trails)
				}
			}
			return nil
		},
	})
	// now register the current site
	site := model.SiteState{
		Id: f.Context.SiteInfo.SiteId,
		ObjectMeta: model.ObjectMeta{
			Name: f.Context.SiteInfo.SiteId,
		},
		Spec: &model.SiteSpec{
			Name:       f.Context.SiteInfo.SiteId,
			Properties: f.Context.SiteInfo.Properties,
			IsSelf:     true,
		},
	}
	oldSite, err := f.SitesManager.GetState(context.Background(), f.Context.SiteInfo.SiteId)
	if err != nil && !utils.IsNotFound(err) {
		return v1alpha2.NewCOAError(err, "Get previous site state failed", v1alpha2.InternalError)
	} else if err == nil {
		site.ObjectMeta.UpdateEtag(oldSite.ObjectMeta.ETag)
	}

	return f.SitesManager.UpsertState(context.Background(), f.Context.SiteInfo.SiteId, site)
}
func (f *FederationVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "federation"
	if f.Route != "" {
		route = f.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodPost, fasthttp.MethodGet},
			Route:      route + "/sync",
			Version:    f.Version,
			Handler:    f.onSync,
			Parameters: []string{"site?"},
		},
		{
			Methods:    []string{fasthttp.MethodPost, fasthttp.MethodGet},
			Route:      route + "/registry",
			Version:    f.Version,
			Handler:    f.onRegistry,
			Parameters: []string{"name?"},
		},
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route + "/status",
			Version:    f.Version,
			Handler:    f.onStatus,
			Parameters: []string{"name"},
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/trail",
			Version: f.Version,
			Handler: f.onTrail,
		},
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   route + "/k8shook",
			Version: f.Version,
			Handler: f.onK8sHook,
		},
	}
}
func (c *FederationVendor) onStatus(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onStatus",
	})
	defer span.End()

	tLog.InfoCtx(pCtx, "V (Federation): OnStatus")
	switch request.Method {
	case fasthttp.MethodPost:
		var state model.SiteState
		utils2.UnmarshalJson(request.Body, &state)

		err := c.SitesManager.ReportState(pCtx, state)

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

func (f *FederationVendor) onRegistry(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onRegistry",
	})
	defer span.End()

	tLog.Info("V (Federation): onRegistry")
	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onRegistry-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = f.SitesManager.ListState(ctx)
			isArray = true
		} else {
			state, err = f.SitesManager.GetState(ctx, id)
		}
		if err != nil {
			if utils.IsNotFound(err) {
				errorMsg := fmt.Sprintf("site '%s' is not found", id)
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(errorMsg),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
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
			resp.ContentType = "text/plain"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onRegistry-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var site model.SiteState
		err := utils2.UnmarshalJson(request.Body, &site)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		//TODO: generate site key pair as needed
		err = f.SitesManager.UpsertState(ctx, id, site)
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
		ctx, span := observability.StartSpan("onRegistry-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := f.SitesManager.DeleteSpec(ctx, id)
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
func (f *FederationVendor) onSync(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onSync",
	})
	defer span.End()

	tLog.Info("V (Federation): onSync")
	switch request.Method {
	case fasthttp.MethodPost:
		var status model.StageStatus
		err := utils2.UnmarshalJson(request.Body, &status)
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Federation): failed to unmarshal stage status: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		err = f.Vendor.Context.Publish("job-report", v1alpha2.Event{
			Body:    status,
			Context: pCtx,
		})
		if err != nil {
			tLog.ErrorfCtx(pCtx, "V (Federation): failed to publish job report: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		tLog.Debugf("V (Federation): published job report: %v", status)
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSync-GET", pCtx, nil)
		id := request.Parameters["__site"]
		count := request.Parameters["count"]
		namespace, exist := request.Parameters["namespace"]
		if !exist {
			namespace = "default"
		}
		if count == "" {
			count = "1"
		}
		intCount, err := strconv.Atoi(count)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		batch, err := f.StagingManager.GetABatchForSite(id, intCount)

		pack := model.SyncPackage{
			Origin: f.Context.SiteInfo.SiteId,
		}

		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.GetErrorState(err),
				Body:  []byte(err.Error()),
			})
		}
		catalogs := make([]model.CatalogState, 0)
		jobs := make([]v1alpha2.JobData, 0)
		for _, c := range batch {
			if c.Action == v1alpha2.JobRun { //TODO: I don't really like this
				jobs = append(jobs, c)
			} else {
				catalog, err := f.CatalogsManager.GetState(ctx, c.Id, namespace)
				if err != nil {
					return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
						State: v1alpha2.GetErrorState(err),
						Body:  []byte(err.Error()),
					})
				}
				catalogs = append(catalogs, catalog)
			}
		}
		pack.Catalogs = catalogs
		pack.Jobs = jobs
		jData, _ := utils.FormatObject(pack, true, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "text/plain"
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
func (f *FederationVendor) onTrail(request v1alpha2.COARequest) v1alpha2.COAResponse {
	_, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onTrail",
	})
	defer span.End()

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	return resp
}
func (f *FederationVendor) onK8sHook(request v1alpha2.COARequest) v1alpha2.COAResponse {
	ctx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onK8sHook",
	})
	defer span.End()

	tLog.Info("V (Federation): onK8sHook")
	switch request.Method {
	case fasthttp.MethodPost:
		objectType := request.Parameters["objectType"]
		if objectType == "catalog" {
			var catalog model.CatalogState
			err := utils2.UnmarshalJson(request.Body, &catalog)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.BadRequest,
					Body:  []byte(err.Error()),
				})
			}
			err = f.Vendor.Context.Publish("catalog", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": catalog.Spec.CatalogType,
				},
				Body: v1alpha2.JobData{
					Id:     catalog.ObjectMeta.Name,
					Action: v1alpha2.JobUpdate, //TODO: handle deletion, this probably requires BetBachForSites return flags
					Body:   catalog,
				},
				Context: ctx,
			})
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
	}

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
