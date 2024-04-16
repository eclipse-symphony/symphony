/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

var lock sync.Mutex

var log = logger.NewLogger("coa.runtime")

type ActivationsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *ActivationsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (m *ActivationsManager) GetState(ctx context.Context, name string, namespace string) (model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": namespace,
			"kind":      "Activation",
		},
	}
	var entry states.StateEntry
	entry, err = m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.ActivationState{}, err
	}
	var ret model.ActivationState
	ret, err = getActivationState(entry.Body, entry.ETag)
	if err != nil {
		return model.ActivationState{}, err
	}
	return ret, nil
}

func getActivationState(body interface{}, etag string) (model.ActivationState, error) {
	var activationState model.ActivationState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &activationState)
	if err != nil {
		return model.ActivationState{}, err
	}
	if activationState.Spec == nil {
		activationState.Spec = &model.ActivationSpec{}
	}
	activationState.Spec.Generation = etag
	if activationState.Status == nil {
		activationState.Status = &model.ActivationStatus{}
	}
	return activationState, nil
}

func (m *ActivationsManager) UpsertState(ctx context.Context, name string, state model.ActivationState) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
			ETag: state.Spec.Generation,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *ActivationsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	return err
}

func (t *ActivationsManager) ListState(ctx context.Context, namespace string) ([]model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": namespace,
			"kind":      "Activation",
		},
	}
	var activations []states.StateEntry
	activations, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.ActivationState, 0)
	for _, t := range activations {
		var rt model.ActivationState
		rt, err = getActivationState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
func (t *ActivationsManager) ReportStatus(ctx context.Context, name string, namespace string, current model.ActivationStatus) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ReportStatus",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	lock.Lock()
	defer lock.Unlock()
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": namespace,
		},
	}
	var entry states.StateEntry
	entry, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return err
	}

	var activationState model.ActivationState
	bytes, _ := json.Marshal(entry.Body)
	err = json.Unmarshal(bytes, &activationState)
	if err != nil {
		observ_utils.CloseSpanWithError(span, &err)
		return err
	}

	current.UpdateTime = time.Now().Format(time.RFC3339) // TODO: is this correct? Shouldn't it be reported?
	activationState.Status = &current

	entry.Body = activationState

	upsertRequest := states.UpsertRequest{
		Value: entry,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": activationState.ObjectMeta.Namespace,
			"kind":      "Activation",
		},
		Options: states.UpsertOption{
			UpdateStateOnly: true,
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}
