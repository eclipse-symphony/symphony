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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/registry"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type TargetsManager struct {
	managers.Manager
	StateProvider    states.IStateProvider
	RegistryProvider registry.IRegistryProvider
}

func (s *TargetsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *TargetsManager) DeleteSpec(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

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
			ETag: state.Spec.Generation,
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

	getRequest := states.GetRequest{
		ID: current.ObjectMeta.Name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": current.ObjectMeta.Namespace,
		},
	}

	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		observ_utils.CloseSpanWithError(span, &err)
		return model.TargetState{}, err
	}

	var targetState model.TargetState
	bytes, _ := json.Marshal(target.Body)
	err = json.Unmarshal(bytes, &targetState)
	if err != nil {
		observ_utils.CloseSpanWithError(span, &err)
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
			UpdateStateOnly: true,
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

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	targets, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.TargetState, 0)
	for _, t := range targets {
		var rt model.TargetState
		rt, err = getTargetState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getTargetState(id string, body interface{}, etag string) (model.TargetState, error) {
	if v, ok := body.(model.TargetState); ok {
		return v, nil
	}
	dict := body.(map[string]interface{})

	//read spec
	spec := dict["spec"]
	j, _ := json.Marshal(spec)
	var rSpec model.TargetSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.TargetState{}, err
	}

	//read status
	status := dict["status"]
	j, _ = json.Marshal(status)
	var rStatus model.TargetStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.TargetState{}, err
	}

	rSpec.Generation = etag

	//read metadata
	metadata := dict["metadata"]
	j, _ = json.Marshal(metadata)
	var rMetadata model.ObjectMeta
	err = json.Unmarshal(j, &rMetadata)
	if err != nil {
		return model.TargetState{}, err
	}

	state := model.TargetState{
		ObjectMeta: rMetadata,
		Spec:       &rSpec,
		Status:     rStatus,
	}

	return state, nil
}

func (t *TargetsManager) GetState(ctx context.Context, id string, namespace string) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	ret, err := getTargetState(id, target.Body, target.ETag)
	if err != nil {
		return model.TargetState{}, err
	}
	return ret, nil
}
