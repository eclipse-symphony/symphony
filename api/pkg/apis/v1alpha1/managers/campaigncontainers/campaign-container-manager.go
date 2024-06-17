/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package campaigncontainers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type CampaignContainersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *CampaignContainersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *CampaignContainersManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("CampaignContainersManager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigncontainers",
			"kind":      "CampaignContainer",
		},
	})
	return err
}

func (t *CampaignContainersManager) UpsertState(ctx context.Context, name string, state model.CampaignContainerState) error {
	ctx, span := observability.StartSpan("CampaignContainersManager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.WorkflowGroup + "/v1",
		"kind":       "CampaignContainer",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: "",
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigncontainers",
			"kind":      "CampaignContainer",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *CampaignContainersManager) ListState(ctx context.Context, namespace string) ([]model.CampaignContainerState, error) {
	ctx, span := observability.StartSpan("CampaignContainersManager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigncontainers",
			"namespace": namespace,
			"kind":      "CampaignContainer",
		},
	}
	var campaigncontainers []states.StateEntry
	campaigncontainers, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CampaignContainerState, 0)
	for _, t := range campaigncontainers {
		var rt model.CampaignContainerState
		rt, err = getCampaignContainerState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getCampaignContainerState(body interface{}, etag string) (model.CampaignContainerState, error) {
	var CampaignContainerState model.CampaignContainerState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &CampaignContainerState)
	if err != nil {
		return model.CampaignContainerState{}, err
	}
	if CampaignContainerState.Spec == nil {
		CampaignContainerState.Spec = &model.CampaignContainerSpec{}
	}
	return CampaignContainerState, nil
}

func (t *CampaignContainersManager) GetState(ctx context.Context, id string, namespace string) (model.CampaignContainerState, error) {
	ctx, span := observability.StartSpan("CampaignContainersManager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigncontainers",
			"namespace": namespace,
			"kind":      "CampaignContainer",
		},
	}
	var Campaign states.StateEntry
	Campaign, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CampaignContainerState{}, err
	}
	var ret model.CampaignContainerState
	ret, err = getCampaignContainerState(Campaign.Body, Campaign.ETag)
	if err != nil {
		return model.CampaignContainerState{}, err
	}
	return ret, nil
}
