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

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
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
	StateProvider     states.IStateProvider
	needValidate      bool
	CampaignValidator validation.CampaignValidator
}

func (s *CampaignsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
	s.needValidate = managers.NeedObjectValidate(config, providers)
	if s.needValidate {
		// Turn off validation of differnt types: https://github.com/eclipse-symphony/symphony/issues/445
		//s.CampaignValidator = validation.NewCampaignValidator(s.CampaignContainerLookup, s.CampaignActivationsLookup)
		s.CampaignValidator = validation.NewCampaignValidator(nil, nil)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "Get campaign state %s in namespace", name, namespace)

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
		log.ErrorfCtx(ctx, "Failed to convert to campaign state for %s in namespace %s: %v", name, namespace, err)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "Upsert campaign state %s in namespace %s", name, state.ObjectMeta.Namespace)
	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	if m.needValidate {
		if state.ObjectMeta.Labels == nil {
			state.ObjectMeta.Labels = make(map[string]string)
		}
		if state.Spec != nil {
			state.ObjectMeta.Labels[constants.RootResource] = state.Spec.RootResource
		}
		if err = m.ValidateCreateOrUpdate(ctx, state); err != nil {
			return err
		}
	}

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
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigns",
			"kind":      "Campaign",
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "Delete campaign state %s in namespace %s", name, namespace)
	if m.needValidate {
		if err = m.ValidateDelete(ctx, name, namespace); err != nil {
			return err
		}
	}

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigns",
			"kind":      "Campaign",
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "List campaign state for namespace %s", namespace)
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "campaigns",
			"namespace": namespace,
			"kind":      "Campaign",
		},
	}
	var campaigns []states.StateEntry
	campaigns, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CampaignState, 0)
	for _, t := range campaigns {
		var rt model.CampaignState
		rt, err = getCampaignState(t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	log.InfofCtx(ctx, "List campaign state for namespace %s get total count %d", namespace, len(ret))
	return ret, nil
}

func (t *CampaignsManager) ValidateCreateOrUpdate(ctx context.Context, state model.CampaignState) error {
	old, err := t.GetState(ctx, state.ObjectMeta.Name, state.ObjectMeta.Namespace)
	return validation.ValidateCreateOrUpdateWrapper(ctx, &t.CampaignValidator, state, old, err)
}

func (t *CampaignsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.CampaignValidator, state, err)
}

func (t *CampaignsManager) CampaignContainerLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.CampaignContainer, name, namespace)
}

func (t *CampaignsManager) CampaignActivationsLookup(ctx context.Context, name string, namespace string) (bool, error) {
	activationList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.Activation, namespace, map[string]string{constants.Campaign: name, constants.StatusMessage: v1alpha2.Running.String()}, 1)
	if err != nil {
		return false, err
	}
	return len(activationList) > 0, nil
}
