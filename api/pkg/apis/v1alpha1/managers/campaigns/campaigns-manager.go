/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package campaigns

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

type CampaignsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *CampaignsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
func (m *CampaignsManager) GetState(ctx context.Context, name string, namespace string) (model.CampaignState, error) {
	ctx, span := observability.StartSpan("Campaigns Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigns",
			"namespace": namespace,
			"kind":      "Campaign",
		},
	}
	var entry states.StateEntry
	entry, err = m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CampaignState{}, err
	}
	var ret model.CampaignState
	ret, err = getCampaignState(entry.Body)
	if err != nil {
		return model.CampaignState{}, err
	}
	return ret, nil
}

func getCampaignState(body interface{}) (model.CampaignState, error) {
	var campaignState model.CampaignState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &campaignState)
	if err != nil {
		return model.CampaignState{}, err
	}
	if campaignState.Spec == nil {
		campaignState.Spec = &model.CampaignSpec{}
	}
	return campaignState, nil
}

func (m *CampaignsManager) UpsertState(ctx context.Context, name string, state model.CampaignState) error {
	ctx, span := observability.StartSpan("Campaigns Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	var rootResource string
	var version string
	var refreshLabels bool
	log.Info("  M (Campaign manager): debug upsert state >>>>>>>>>>>>>>>>>>>>  %v, %v, %v", state.Spec.Version, state.Spec.RootResource, name)

	if state.Spec.Version != "" {
		version = state.Spec.Version
	}
	if state.Spec.RootResource == "" && version != "" {
		suffix := "-" + version
		rootResource = strings.TrimSuffix(name, suffix)
	} else {
		rootResource = state.Spec.RootResource
	}

	if state.ObjectMeta.Labels == nil {
		state.ObjectMeta.Labels = make(map[string]string)
	}

	_, versionLabelExists := state.ObjectMeta.Labels["version"]
	_, rootLabelExists := state.ObjectMeta.Labels["rootResource"]
	if (!versionLabelExists || !rootLabelExists) && version != "" && rootResource != "" {
		log.Info("  M (Campaign manager): update labels to true >>>>>>>>>>>>>>>>>>>>  %v, %v", rootResource, version)

		state.ObjectMeta.Labels["rootResource"] = rootResource
		state.ObjectMeta.Labels["version"] = version
		refreshLabels = true
	}
	log.Info("  M (Campaign manager): update labels to versionLabelExists, rootLabelExists >>>>>>>>>>>>>>>>>>>>  %v, %v", versionLabelExists, rootLabelExists)
	log.Info("  M (Campaign manager): debug refresh >>>>>>>>>>>>>>>>>>>>  %v", refreshLabels)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Campaign",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
		},
		Metadata: map[string]interface{}{
			"namespace":     state.ObjectMeta.Namespace,
			"group":         model.WorkflowGroup,
			"version":       "v1",
			"resource":      "campaigns",
			"kind":          "Campaign",
			"rootResource":  rootResource,
			"refreshLabels": strconv.FormatBool(refreshLabels),
		},
	}

	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (m *CampaignsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Campaigns Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var rootResource string
	var version string
	var id string
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		rootResource = parts[0]
		version = parts[1]
		id = rootResource + "-" + version
	} else {
		id = name
	}
	log.Info("  M (Catalog manager): delete state >>>>>>>>>>>>>>>>>>>>parts  %v, %v", rootResource, version)

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"namespace":    namespace,
			"group":        model.WorkflowGroup,
			"version":      "v1",
			"resource":     "campaigns",
			"kind":         "Campaign",
			"rootResource": rootResource,
		},
	})
	return err
}

func (t *CampaignsManager) ListState(ctx context.Context, namespace string) ([]model.CampaignState, error) {
	ctx, span := observability.StartSpan("Campaigns Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigns",
			"namespace": namespace,
			"kind":      "Campaign",
		},
	}
	var solutions []states.StateEntry
	solutions, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CampaignState, 0)
	for _, t := range solutions {
		var rt model.CampaignState
		rt, err = getCampaignState(t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func (t *CampaignsManager) GetLatestState(ctx context.Context, id string, namespace string) (model.CampaignState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "GetLatest",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info("  M (Campaign manager): debug get latest state >>>>>>>>>>>>>>>>>>>>  %v, %v", id, namespace)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigns",
			"namespace": namespace,
			"kind":      "Campaign",
		},
	}
	entry, err := t.StateProvider.GetLatest(ctx, getRequest)
	if err != nil {
		return model.CampaignState{}, err
	}

	ret, err := getCampaignState(entry.Body)
	if err != nil {
		return model.CampaignState{}, err
	}
	return ret, nil
}
