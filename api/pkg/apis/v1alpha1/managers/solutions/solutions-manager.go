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

func (t *SolutionsManager) DeleteSpec(ctx context.Context, name string, scope string) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    scope,
			"group":    model.SolutionGroup,
			"version":  "v1",
			"resource": "solutions",
		},
	})
	return err
}

func (t *SolutionsManager) UpsertSpec(ctx context.Context, name string, spec model.SolutionSpec, scope string) error {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.SolutionGroup + "/v1",
				"kind":       "Solution",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Solution", "metadata": {"name": "${{$solution()}}"}}`, model.SolutionGroup),
			"scope":    scope,
			"group":    model.SolutionGroup,
			"version":  "v1",
			"resource": "solutions",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (t *SolutionsManager) ListSpec(ctx context.Context, scope string) ([]model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.SolutionGroup,
			"resource": "solutions",
			"scope":    scope,
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
	scope, exist := dict["scope"]
	var s string
	if !exist {
		s = "default"
	} else {
		s = scope.(string)
	}

	state := model.SolutionState{
		Id:    id,
		Scope: s,
		Spec:  &rSpec,
	}
	return state, nil
}

func (t *SolutionsManager) GetSpec(ctx context.Context, id string, scope string) (model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.SolutionGroup,
			"resource": "solutions",
			"scope":    scope,
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
