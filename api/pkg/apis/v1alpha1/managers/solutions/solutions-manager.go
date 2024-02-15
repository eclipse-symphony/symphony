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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type SolutionsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SolutionsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *SolutionsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutions",
		},
	})
	return err
}

func (t *SolutionsManager) UpsertState(ctx context.Context, name string, state model.SolutionState, namespace string) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "UpsertState",
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
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Solution",
		"metadata":   metadata,
		"spec":       state.Spec,
	}
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
		},
		Metadata: map[string]string{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Solution", "metadata": {"name": "${{$solution()}}"}}`, model.SolutionGroup),
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutions",
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

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutions",
			"namespace": namespace,
		},
	}
	solutions, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SolutionState, 0)
	for _, t := range solutions {
		var rt model.SolutionState
		rt, err = getSolutionState(t.ID, t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSolutionState(id string, body interface{}) (model.SolutionState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]

	j, _ := json.Marshal(spec)
	var rSpec model.SolutionSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.SolutionState{}, err
	}
	namespace, exist := dict["namespace"]
	var s string
	if !exist {
		s = "default"
	} else {
		s = namespace.(string)
	}

	state := model.SolutionState{
		Id:        id,
		Namespace: s,
		Spec:      &rSpec,
	}
	return state, nil
}

func (t *SolutionsManager) GetState(ctx context.Context, id string, namespace string) (model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutions",
			"namespace": namespace,
		},
	}
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SolutionState{}, err
	}

	ret, err := getSolutionState(id, target.Body)
	if err != nil {
		return model.SolutionState{}, err
	}
	return ret, nil
}
