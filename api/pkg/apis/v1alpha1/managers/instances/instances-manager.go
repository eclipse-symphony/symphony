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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type InstancesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *InstancesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *InstancesManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Instance",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}
	generation := ""
	if state.Spec != nil {
		generation = state.Spec.Generation
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
	instanceState.Spec.Generation = etag
	return instanceState, nil
}

func (t *InstancesManager) GetState(ctx context.Context, id string, namespace string) (model.InstanceState, error) {
	ctx, span := observability.StartSpan("Instances Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

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
