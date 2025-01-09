/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sites

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type SitesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
	apiClient     utils.ApiClient
}

func (s *SitesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	stateprovider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	s.apiClient, err = utils.GetParentApiClient(s.VendorContext.SiteInfo.ParentSite.BaseUrl)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) GetState(ctx context.Context, name string) (model.SiteState, error) {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	var entry states.StateEntry
	entry, err = m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SiteState{}, err
	}
	var ret model.SiteState
	ret, err = getSiteState(name, entry.Body)
	if err != nil {
		return model.SiteState{}, err
	}
	ret.ObjectMeta.UpdateEtag(entry.ETag)
	return ret, nil
}

func getSiteState(id string, body interface{}) (model.SiteState, error) {
	var siteState model.SiteState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &siteState)
	if err != nil {
		return model.SiteState{}, err
	}
	siteState.Id = id
	if siteState.Spec == nil {
		siteState.Spec = &model.SiteSpec{}
	}
	if siteState.Status == nil {
		siteState.Status = &model.SiteStatus{}
	}
	return siteState, nil
}

func (t *SitesManager) ReportState(ctx context.Context, current model.SiteState) error {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "ReportState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: current.Id,
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}

	// Need to persist site spec if not exists
	var entry states.StateEntry
	entry, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		if !utils.IsNotFound(err) {
			return err
		}
		err = t.UpsertState(ctx, current.Id, current)
		if err != nil {
			return err
		}
		entry, err = t.StateProvider.Get(ctx, getRequest)
		if err != nil {
			return err
		}
	}

	var siteState model.SiteState
	siteState, err = getSiteState(entry.ID, entry.Body)
	if err != nil {
		return err
	}
	siteState.ObjectMeta.UpdateEtag(entry.ETag)
	if siteState.Status == nil {
		siteState.Status = &model.SiteStatus{}
	}

	// if current.Status is not nil, update the status using new IsOnline, InstanceStatuses and TargetStatuses
	// otherwise, only update LastReported as time.Now()
	if current.Status != nil {
		siteState.Status.IsOnline = current.Status.IsOnline
		siteState.Status.InstanceStatuses = current.Status.InstanceStatuses
		siteState.Status.TargetStatuses = current.Status.TargetStatuses
	}
	siteState.Status.LastReported = time.Now().UTC().Format(time.RFC3339)

	updateRequest := states.UpsertRequest{
		Value: states.StateEntry{ID: current.Id, Body: siteState, ETag: entry.ETag},
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) UpsertState(ctx context.Context, name string, state model.SiteState) error {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "UpsertState: Upsert site state %s", name)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		log.ErrorfCtx(ctx, "Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name)
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}

	oldState, getStateErr := m.GetState(ctx, state.ObjectMeta.Name)
	if getStateErr == nil {
		state.ObjectMeta.PreserveSystemMetadata(oldState.ObjectMeta)
	}

	log.InfofCtx(ctx, "Site state %+v", state)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Site",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
			ETag: state.ObjectMeta.ETag,
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Site", "metadata": {"name": "${{$site()}}"}}`, model.FederationGroup),
			"namespace": "",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "sites",
			"kind":      "Site",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) DeleteSpec(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": "",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "sites",
			"kind":      "Site",
		},
	})

	return err
}

func (t *SitesManager) ListState(ctx context.Context) ([]model.SiteState, error) {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	var sites []states.StateEntry
	sites, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SiteState, 0)
	for _, t := range sites {
		var rt model.SiteState
		rt, err = getSiteState(t.ID, t.Body)
		if err != nil {
			return nil, err
		}
		rt.ObjectMeta.UpdateEtag(t.ETag)
		ret = append(ret, rt)
	}
	return ret, nil
}
func (s *SitesManager) Enabled() bool {
	return s.VendorContext.SiteInfo.ParentSite.BaseUrl != ""
}
func (s *SitesManager) Poll() []error {
	ctx, span := observability.StartSpan("Sites Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var thisSite model.SiteState
	thisSite, err = s.GetState(ctx, s.VendorContext.SiteInfo.SiteId)
	if err != nil {
		//TOOD: only ignore not found, and log the error
		return nil
	}
	thisSite.Spec.IsSelf = false
	jData, _ := json.Marshal(thisSite)
	s.apiClient.UpdateSite(
		ctx,
		s.VendorContext.SiteInfo.SiteId,
		jData,
		s.VendorContext.SiteInfo.ParentSite.Username,
		s.VendorContext.SiteInfo.ParentSite.Password)
	return nil
}
func (s *SitesManager) Reconcil() []error {
	return nil
}
