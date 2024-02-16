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
		Metadata: map[string]string{
			"namespace": namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
		},
	})
	return err
}

func (t *TargetsManager) UpsertState(ctx context.Context, name string, namespace string, state model.TargetState) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	metadata := map[string]interface{}{
		"name": name,
	}
	for k, v := range state.Metadata {
		metadata[k] = v
	}
	body := map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "Target",
		"metadata":   metadata,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: state.Spec.Generation,
		},
		Metadata: map[string]string{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
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
		ID:       current.Id,
		Metadata: current.Metadata,
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

	target.Body = targetState

	updateRequest := states.UpsertRequest{
		Value:    target,
		Metadata: current.Metadata,
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
		Metadata: map[string]string{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
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
	spec := dict["spec"]
	status := dict["status"]

	j, _ := json.Marshal(spec)
	var rSpec model.TargetSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.TargetState{}, err
	}

	j, _ = json.Marshal(status)
	var rStatus model.TargetStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.TargetState{}, err
	}

	rSpec.Generation = etag

	namespace, exist := dict["namespace"]
	var s string
	if !exist {
		s = "default"
	} else {
		s = namespace.(string)
	}

	state := model.TargetState{
		Id:        id,
		Namespace: s,
		Spec:      &rSpec,
		Status:    rStatus,
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
		Metadata: map[string]string{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
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
