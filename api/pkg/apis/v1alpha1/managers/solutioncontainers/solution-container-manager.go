/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutioncontainers

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
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type SolutionContainersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SolutionContainersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *SolutionContainersManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("SolutionContainersManager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutioncontainers",
			"kind":      "SolutionContainer",
		},
	})
	return err
}

func (t *SolutionContainersManager) UpsertState(ctx context.Context, name string, state model.SolutionContainerState) error {
	ctx, span := observability.StartSpan("SolutionContainersManager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "SolutionContainer",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: "",
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutioncontainers",
			"kind":      "SolutionContainer",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *SolutionContainersManager) ListState(ctx context.Context, namespace string) ([]model.SolutionContainerState, error) {
	ctx, span := observability.StartSpan("SolutionContainersManager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutioncontainers",
			"namespace": namespace,
			"kind":      "SolutionContainer",
		},
	}
	var solutioncontainers []states.StateEntry
	solutioncontainers, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SolutionContainerState, 0)
	for _, t := range solutioncontainers {
		var rt model.SolutionContainerState
		rt, err = getSolutionContainerState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSolutionContainerState(body interface{}, etag string) (model.SolutionContainerState, error) {
	var SolutionContainerState model.SolutionContainerState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &SolutionContainerState)
	if err != nil {
		return model.SolutionContainerState{}, err
	}
	if SolutionContainerState.Spec == nil {
		SolutionContainerState.Spec = &model.SolutionContainerSpec{}
	}
	return SolutionContainerState, nil
}

func (t *SolutionContainersManager) GetState(ctx context.Context, id string, namespace string) (model.SolutionContainerState, error) {
	ctx, span := observability.StartSpan("SolutionContainersManager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "solutioncontainers",
			"namespace": namespace,
			"kind":      "SolutionContainer",
		},
	}
	var Solution states.StateEntry
	Solution, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SolutionContainerState{}, err
	}
	var ret model.SolutionContainerState
	ret, err = getSolutionContainerState(Solution.Body, Solution.ETag)
	if err != nil {
		return model.SolutionContainerState{}, err
	}
	return ret, nil
}
