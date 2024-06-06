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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

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

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Solution",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	if state.Spec != nil {
		rootResource := state.Spec.RootResource
		if rootResource != "" {
			log.Debugf(" M (Solutions): solution root resource: %s, solution: %s", rootResource, name)
			resourceName := "solutioncontainers"
			kind := "SolutionContainer"
			containerMetadata := map[string]interface{}{
				"version":   "v1",
				"group":     model.SolutionGroup,
				"resource":  resourceName,
				"namespace": state.ObjectMeta.Namespace,
				"kind":      kind,
			}
			getRequest := states.GetRequest{
				ID:       rootResource,
				Metadata: containerMetadata,
			}
			_, err = t.StateProvider.Get(ctx, getRequest)
			if err != nil {
				log.Debugf(" M (Solutions): get solution container %s, err %v", rootResource, err)
				cErr, ok := err.(v1alpha2.COAError)
				if ok && cErr.State == v1alpha2.NotFound {
					containerBody := map[string]interface{}{
						"apiVersion": model.SolutionGroup + "/v1",
						"kind":       kind,
						"metadata":   model.ObjectMeta{Namespace: state.ObjectMeta.Namespace, Name: rootResource},
						"spec":       model.SolutionContainerSpec{},
					}
					containerUpsertRequest := states.UpsertRequest{
						Value: states.StateEntry{
							ID:   rootResource,
							Body: containerBody,
						},
						Metadata: containerMetadata,
					}
					_, err = t.StateProvider.Upsert(ctx, containerUpsertRequest)
					if err != nil {
						log.Errorf(" M (Solutions): failed to create solution container %s, namespace: %v, err %v", rootResource, state.ObjectMeta.Namespace, err)
						return err
					}
				}
			}
		}
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
