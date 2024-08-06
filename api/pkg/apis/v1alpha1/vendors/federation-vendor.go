/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
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
	f.Vendor.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		sites, err := f.SitesManager.ListState(context.TODO())
		if err != nil {
			return err
		}
		for _, site := range sites {
			if site.Spec.Name != f.Vendor.Context.SiteInfo.SiteId {
				event.Metadata["site"] = site.Spec.Name
				f.StagingManager.HandleJobEvent(context.TODO(), event) //TODO: how to handle errors in this case?
			}
		}
		return nil
	})
	f.Vendor.Context.Subscribe("remote", func(topic string, event v1alpha2.Event) error {
		_, ok := event.Metadata["site"]
		if !ok {
			return v1alpha2.NewCOAError(nil, "site is not supplied", v1alpha2.BadRequest)
		}
		f.StagingManager.HandleJobEvent(context.TODO(), event) //TODO: how to handle errors in this case?
		return nil
	})
	f.Vendor.Context.Subscribe("report", func(topic string, event v1alpha2.Event) error {
		fLog.Debugf("V (Federation): received report event: %v", event)
		jData, _ := json.Marshal(event.Body)
		var status model.StageStatus
		err := json.Unmarshal(jData, &status)
		if err == nil {
			err := f.apiClient.SyncStageStatus(context.TODO(), status,
				f.Vendor.Context.SiteInfo.ParentSite.Username,
				f.Vendor.Context.SiteInfo.ParentSite.Password)
			if err != nil {
				fLog.Errorf("V (Federation): error while syncing activation status: %v", err)
				return err
			}
		}
		return v1alpha2.NewCOAError(nil, "report is not an activation status", v1alpha2.BadRequest)
	})
	f.Vendor.Context.Subscribe("trail", func(topic string, event v1alpha2.Event) error {
		if f.TrailsManager != nil {
			jData, _ := json.Marshal(event.Body)
			var trails []v1alpha2.Trail
			err := json.Unmarshal(jData, &trails)
			if err == nil {
				return f.TrailsManager.Append(context.TODO(), trails)
			}
		}
		return nil
	})
	//now register the current site
	return f.SitesManager.UpsertSpec(context.TODO(), f.Context.SiteInfo.SiteId, model.SiteSpec{
		Name:       f.Context.SiteInfo.SiteId,
		Properties: f.Context.SiteInfo.Properties,
		IsSelf:     true,
	})
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

	tLog.Info("V (Federation): OnStatus")
	switch request.Method {
	case fasthttp.MethodPost:
		var state model.SiteState
		json.Unmarshal(request.Body, &state)

		err := c.SitesManager.ReportState(pCtx, state)

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
			if v1alpha2.IsNotFound(err) {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.NotFound,
					Body:  []byte(err.Error()),
				})
			} else {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
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
		// TODO: POST federation/registry need to pass SiteState as request body
		ctx, span := observability.StartSpan("onRegistry-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var site model.SiteSpec
		err := json.Unmarshal(request.Body, &site)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		//TODO: generate site key pair as needed
		err = f.SitesManager.UpsertSpec(ctx, id, site)
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
		ctx, span := observability.StartSpan("onRegistry-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := f.SitesManager.DeleteSpec(ctx, id)
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
func (f *FederationVendor) onSync(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onSync",
	})
	defer span.End()

	tLog.Info("V (Federation): onSync")
	switch request.Method {
	case fasthttp.MethodPost:
		var status model.StageStatus
		err := json.Unmarshal(request.Body, &status)
		if err != nil {
			tLog.Errorf("V (Federation): failed to unmarshal stage status: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}
		err = f.Vendor.Context.Publish("job-report", v1alpha2.Event{
			Body: status,
		})
		if err != nil {
			tLog.Errorf("V (Federation): failed to publish job report: %v", err)
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
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
				State: v1alpha2.InternalError,
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
						State: v1alpha2.InternalError,
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
	_, span := observability.StartSpan("Federation Vendor", request.Context, &map[string]string{
		"method": "onK8sHook",
	})
	defer span.End()

	tLog.Info("V (Federation): onK8sHook")
	switch request.Method {
	case fasthttp.MethodPost:
		objectType := request.Parameters["objectType"]
		if objectType == "catalog" {
			var catalog model.CatalogState
			err := json.Unmarshal(request.Body, &catalog)
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
			})
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
	}

	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
