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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/registry"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"

	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
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

func (t *TargetsManager) DeleteSpec(ctx context.Context, name string, scope string) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    scope,
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	})
	return err
}

func (t *TargetsManager) UpsertSpec(ctx context.Context, name string, scope string, spec model.TargetSpec) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
			ETag: spec.Generation,
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"scope":    scope,
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "targets",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// Caller need to explicitly set scope in current.Metadata!
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

	dict := target.Body.(map[string]interface{})

	specCol := dict["spec"].(model.TargetSpec)

	delete(dict, "spec")
	if dict["status"] == nil {
		dict["status"] = make(map[string]interface{})
	}
	status := dict["status"]

	j, _ := json.Marshal(status)
	var rStatus map[string]interface{}
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.TargetState{}, err
	}
	j, _ = json.Marshal(rStatus["properties"])
	var rProperties map[string]string
	err = json.Unmarshal(j, &rProperties)
	if err != nil {
		return model.TargetState{}, err
	}
	if rProperties == nil {
		rProperties = make(map[string]string)
	}
	for k, v := range current.Status {
		rProperties[k] = v
	}

	dict["status"].(map[string]interface{})["properties"] = rProperties

	target.Body = dict

	updateRequest := states.UpsertRequest{
		Value:    target,
		Metadata: current.Metadata,
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	return model.TargetState{
		Id:       current.Id,
		Metadata: specCol.Metadata,
		Status:   rProperties,
	}, nil
}
func (t *TargetsManager) ListSpec(ctx context.Context, scope string) ([]model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "targets",
			"scope":    scope,
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
	var rStatus map[string]interface{}
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.TargetState{}, err
	}
	j, _ = json.Marshal(rStatus["properties"])
	var rProperties map[string]string
	err = json.Unmarshal(j, &rProperties)
	if err != nil {
		return model.TargetState{}, err
	}
	rSpec.Generation = etag

	scope, exist := dict["scope"]
	var s string
	if !exist {
		s = "default"
	} else {
		s = scope.(string)
	}

	state := model.TargetState{
		Id:     id,
		Scope:  s,
		Spec:   &rSpec,
		Status: rProperties,
	}
	return state, nil
}

func (t *TargetsManager) GetSpec(ctx context.Context, id string, scope string) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "targets",
			"scope":    scope,
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
