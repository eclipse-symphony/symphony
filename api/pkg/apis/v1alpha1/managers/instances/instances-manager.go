/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package instances

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

type InstancesManager struct {
	managers.Manager
	StateProvider     states.IStateProvider
	needValidate      bool
	InstanceValidator validation.InstanceValidator
}

func (s *InstancesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
		//s.InstanceValidator = validation.NewInstanceValidator(s.instanceUniqueNameLookup, s.solutionLookup, s.targetLookup)
		s.InstanceValidator = validation.NewInstanceValidator(s.instanceUniqueNameLookup, nil, nil)
	}
	return nil
}

func (t *InstancesManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "instances",
			"kind":      "Instance",
		},
	})
	return err
}

func (t *InstancesManager) UpsertState(ctx context.Context, name string, state model.InstanceState) error {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
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
			state.ObjectMeta.Labels[constants.Solution] = state.Spec.Solution
			state.ObjectMeta.Labels[constants.Target] = state.Spec.Target.Name
		}
		if err = t.ValidateCreateOrUpdate(ctx, state); err != nil {
			return err
		}
	}

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Instance",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}
	generation := ""
	if state.Spec != nil {
		generation = state.ObjectMeta.Generation
	}
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: generation,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "instances",
			"kind":      "Instance",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *InstancesManager) ListState(ctx context.Context, namespace string) ([]model.InstanceState, error) {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "instances",
			"namespace": namespace,
			"kind":      "Instance",
		},
	}
	var instances []states.StateEntry
	instances, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.InstanceState, 0)
	for _, t := range instances {
		var rt model.InstanceState
		rt, err = getInstanceState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getInstanceState(body interface{}, etag string) (model.InstanceState, error) {
	var instanceState model.InstanceState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &instanceState)
	if err != nil {
		return model.InstanceState{}, err
	}
	if instanceState.Spec == nil {
		instanceState.Spec = &model.InstanceSpec{}
	}
	instanceState.ObjectMeta.Generation = etag
	return instanceState, nil
}

func (t *InstancesManager) GetState(ctx context.Context, id string, namespace string) (model.InstanceState, error) {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
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
			"resource":  "instances",
			"namespace": namespace,
			"kind":      "Instance",
		},
	}
	var instance states.StateEntry
	instance, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.InstanceState{}, err
	}
	var ret model.InstanceState
	ret, err = getInstanceState(instance.Body, instance.ETag)
	if err != nil {
		return model.InstanceState{}, err
	}
	return ret, nil
}

func (t *InstancesManager) ValidateCreateOrUpdate(ctx context.Context, state model.InstanceState) error {
	old, err := t.GetState(ctx, state.ObjectMeta.Name, state.ObjectMeta.Namespace)
	return validation.ValidateCreateOrUpdateWrapper(ctx, &t.InstanceValidator, state, old, err)
}

func (t *InstancesManager) instanceUniqueNameLookup(ctx context.Context, displayName string, namespace string) (interface{}, error) {
	return states.GetObjectStateWithUniqueName(ctx, t.StateProvider, validation.Instance, displayName, namespace)
}
func (t *InstancesManager) solutionLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.Solution, name, namespace)
}
func (t *InstancesManager) targetLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.Target, name, namespace)
}
