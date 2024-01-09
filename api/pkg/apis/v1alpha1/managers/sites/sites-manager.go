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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

type SitesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SitesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

// GetCampaign retrieves a CampaignSpec object by name
func (m *SitesManager) GetSpec(ctx context.Context, name string) (model.SiteState, error) {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	entry, err := m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SiteState{}, err
	}

	ret, err := getSiteState(name, entry.Body)
	if err != nil {
		return model.SiteState{}, err
	}
	return ret, nil
}

func getSiteState(id string, body interface{}) (model.SiteState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]
	status := dict["status"]

	j, _ := json.Marshal(spec)
	var rSpec model.SiteSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.SiteState{}, err
	}

	var rStatus model.SiteStatus

	if status != nil {
		j, _ = json.Marshal(status)
		err = json.Unmarshal(j, &rStatus)
		if err != nil {
			return model.SiteState{}, err
		}
	}
	state := model.SiteState{
		Id:     id,
		Spec:   &rSpec,
		Status: &rStatus,
	}
	return state, nil
}

func (t *SitesManager) ReportState(ctx context.Context, current model.SiteState) error {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "ReportState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	current.Metadata = map[string]string{
		"version":  "v1",
		"group":    model.FederationGroup,
		"resource": "sites",
	}
	getRequest := states.GetRequest{
		ID: current.Id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}

	entry, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			return err
		}
		err = t.UpsertSpec(ctx, current.Id, *current.Spec)
		if err != nil {
			return err
		}
		entry, err = t.StateProvider.Get(ctx, getRequest)
		if err != nil {
			return err
		}
	}

	// This copy is necessary becasue otherwise you could be modifying data in memory stage provider
	jTransfer, _ := json.Marshal(entry.Body)
	var dict map[string]interface{}
	json.Unmarshal(jTransfer, &dict)

	delete(dict, "spec")
	status := dict["status"]

	j, _ := json.Marshal(status)
	var rStatus model.SiteStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return err
	}
	// if current.Status is not nil, update the status using new IsOnline, InstanceStatuses and TargetStatuses
	// otherwise, only update LastReported as time.Now()
	if current.Status != nil {
		rStatus.IsOnline = current.Status.IsOnline
		rStatus.InstanceStatuses = current.Status.InstanceStatuses
		rStatus.TargetStatuses = current.Status.TargetStatuses
	}
	rStatus.LastReported = time.Now().UTC().Format(time.RFC3339)
	dict["status"] = rStatus

	entry.Body = dict

	updateRequest := states.UpsertRequest{
		Value:    entry,
		Metadata: current.Metadata,
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) UpsertSpec(ctx context.Context, name string, spec model.SiteSpec) error {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Site",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Site", "metadata": {"name": "${{$site()}}"}}`, model.FederationGroup),
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "sites",
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

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "sites",
		},
	})

	return err
}

func (t *SitesManager) ListSpec(ctx context.Context) ([]model.SiteState, error) {
	ctx, span := observability.StartSpan("Sites Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	sites, _, err := t.StateProvider.List(ctx, listRequest)
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

	thisSite, err := s.GetSpec(ctx, s.VendorContext.SiteInfo.SiteId)
	if err != nil {
		//TOOD: only ignore not found, and log the error
		return nil
	}
	thisSite.Spec.IsSelf = false
	jData, _ := json.Marshal(thisSite)
	utils.UpdateSite(
		ctx,
		s.VendorContext.SiteInfo.ParentSite.BaseUrl,
		s.VendorContext.SiteInfo.SiteId,
		s.VendorContext.SiteInfo.ParentSite.Username,
		s.VendorContext.SiteInfo.ParentSite.Password,
		jData,
	)
	return nil
}
func (s *SitesManager) Reconcil() []error {
	return nil
}
