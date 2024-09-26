/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutions

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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type SolutionsManager struct {
	managers.Manager
	StateProvider     states.IStateProvider
	needValidate      bool
	SolutionValidator validation.SolutionValidator
}

func (s *SolutionsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
		// s.SolutionValidator = validation.NewSolutionValidator(s.solutionInstanceLookup, s.solutionContainerLookup, s.uniqueNameSolutionLookup)
		s.SolutionValidator = validation.NewSolutionValidator(nil, nil, s.uniqueNameSolutionLookup)
	}
	return nil
}

func (t *SolutionsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if t.needValidate {
		if err = t.ValidateDelete(ctx, name, namespace); err != nil {
			return err
		}
	}

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutions",
			"kind":      "Solution",
		},
	})
	return err
}

func (t *SolutionsManager) UpsertState(ctx context.Context, name string, state model.SolutionState) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	if t.needValidate {
		if state.ObjectMeta.Labels == nil {
			state.ObjectMeta.Labels = make(map[string]string)
		}
		if state.Spec != nil {
			state.ObjectMeta.Labels[constants.DisplayName] = utils.ConvertStringToValidLabel(state.Spec.DisplayName)
			state.ObjectMeta.Labels[constants.RootResource] = state.Spec.RootResource
		}
		if err = t.ValidateCreateOrUpdate(ctx, state); err != nil {
			return err
		}
	}

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Solution",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutions",
			"kind":      "Solution",
		},
	}

	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (t *SolutionsManager) ListState(ctx context.Context, namespace string) ([]model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutions",
			"namespace": namespace,
			"kind":      "Solution",
		},
	}
	var solutions []states.StateEntry
	solutions, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SolutionState, 0)
	for _, t := range solutions {
		var rt model.SolutionState
		rt, err = getSolutionState(t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSolutionState(body interface{}) (model.SolutionState, error) {
	var solutionState model.SolutionState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &solutionState)
	if err != nil {
		return model.SolutionState{}, err
	}
	if solutionState.Spec == nil {
		solutionState.Spec = &model.SolutionSpec{}
	}
	return solutionState, nil
}

func (t *SolutionsManager) GetState(ctx context.Context, id string, namespace string) (model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutions",
			"namespace": namespace,
			"kind":      "Solution",
		},
	}
	var target states.StateEntry
	target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SolutionState{}, err
	}
	var ret model.SolutionState
	ret, err = getSolutionState(target.Body)
	if err != nil {
		return model.SolutionState{}, err
	}
	return ret, nil
}

func (t *SolutionsManager) ValidateCreateOrUpdate(ctx context.Context, state model.SolutionState) error {
	old, err := t.GetState(ctx, state.ObjectMeta.Name, state.ObjectMeta.Namespace)
	return validation.ValidateCreateOrUpdateWrapper(ctx, &t.SolutionValidator, state, old, err)
}

func (t *SolutionsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.SolutionValidator, state, err)
}

func (t *SolutionsManager) solutionInstanceLookup(ctx context.Context, name string, namespace string) (bool, error) {
	instanceList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.Instance, namespace, map[string]string{constants.Solution: name}, 1)
	if err != nil {
		return false, err
	}
	return len(instanceList) > 0, nil
}

func (t *SolutionsManager) solutionContainerLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.SolutionContainer, name, namespace)
}

func (t *SolutionsManager) uniqueNameSolutionLookup(ctx context.Context, displayName string, namespace string) (interface{}, error) {
	return states.GetObjectStateWithUniqueName(ctx, t.StateProvider, validation.Solution, displayName, namespace)
}
