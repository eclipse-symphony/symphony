/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package targets

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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/registry"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type TargetsManager struct {
	managers.Manager
	StateProvider    states.IStateProvider
	RegistryProvider registry.IRegistryProvider
	needValidate     bool
	TargetValidator  validation.TargetValidator
	SecretProvider   secret.ISecretProvider
}

func (s *TargetsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
		// s.TargetValidator = validation.NewTargetValidator(s.targetInstanceLookup, s.targetUniqueNameLookup)
		s.TargetValidator = validation.NewTargetValidator(nil, s.targetUniqueNameLookup)
	}

	for _, p := range providers {
		if c, ok := p.(secret.ISecretProvider); ok {
			s.SecretProvider = c
		}
	}

	return nil
}

func (t *TargetsManager) DeleteSpec(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
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
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	return err
}

func (t *TargetsManager) UpsertState(ctx context.Context, name string, state model.TargetState) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	oldState, getStateErr := t.GetState(ctx, state.ObjectMeta.Name, state.ObjectMeta.Namespace)
	if getStateErr == nil {
		state.ObjectMeta.PreserveSystemMetadata(oldState.ObjectMeta)
	}

	if t.needValidate {
		if state.ObjectMeta.Labels == nil {
			state.ObjectMeta.Labels = make(map[string]string)
		}
		if state.Spec != nil {
			state.ObjectMeta.Labels[constants.DisplayName] = utils.ConvertStringToValidLabel(state.Spec.DisplayName)
		}
		if err = validation.ValidateCreateOrUpdateWrapper(ctx, &t.TargetValidator, state, oldState, getStateErr); err != nil {
			return err
		}
	}

	body := map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "Target",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: state.ObjectMeta.ETag,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	}

	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// Caller need to explicitly set namespace in current.Metadata!
func (t *TargetsManager) ReportState(ctx context.Context, current model.TargetState) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "ReportState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: current.ObjectMeta.Name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": current.ObjectMeta.Namespace,
		},
	}

	var target states.StateEntry
	target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	var targetState model.TargetState
	bytes, _ := json.Marshal(target.Body)
	err = json.Unmarshal(bytes, &targetState)
	if err != nil {
		return model.TargetState{}, err
	}

	for k, v := range current.Status.Properties {
		if targetState.Status.Properties == nil {
			targetState.Status.Properties = make(map[string]string)
		}
		targetState.Status.Properties[k] = v
	}
	targetState.Status.LastModified = current.Status.LastModified

	target.Body = targetState

	updateRequest := states.UpsertRequest{
		Value: target,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": current.ObjectMeta.Namespace,
			"kind":      "Target",
		},
		Options: states.UpsertOption{
			UpdateStatusOnly: true,
		},
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return model.TargetState{}, err
	}
	return targetState, nil
}
func (t *TargetsManager) ListState(ctx context.Context, namespace string) ([]model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	var targets []states.StateEntry
	targets, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.TargetState, 0)
	for _, t := range targets {
		var rt model.TargetState
		rt, err = getTargetState(t.Body)
		if err != nil {
			return nil, err
		}
		rt.ObjectMeta.UpdateEtag(t.ETag)
		ret = append(ret, rt)
	}
	return ret, nil
}

func getTargetState(body interface{}) (model.TargetState, error) {
	var targetState model.TargetState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &targetState)
	if err != nil {
		return model.TargetState{}, err
	}
	if targetState.Spec == nil {
		targetState.Spec = &model.TargetSpec{}
	}
	return targetState, nil
}

func (t *TargetsManager) GetState(ctx context.Context, id string, namespace string) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	var target states.StateEntry
	target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	var ret model.TargetState
	ret, err = getTargetState(target.Body)
	if err != nil {
		return model.TargetState{}, err
	}
	ret.ObjectMeta.UpdateEtag(target.ETag)
	return ret, nil
}

func (t *TargetsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.TargetValidator, state, err)
}

func (t *TargetsManager) targetUniqueNameLookup(ctx context.Context, displayName string, namespace string) (interface{}, error) {
	return states.GetObjectStateWithUniqueName(ctx, t.StateProvider, validation.Target, displayName, namespace)
}

func (t *TargetsManager) targetInstanceLookup(ctx context.Context, name string, namespace string) (bool, error) {
	instanceList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.Instance, namespace, map[string]string{constants.Target: name}, 1)
	if err != nil {
		return false, err
	}
	return len(instanceList) > 0, nil
}
