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
	"strconv"
	"strings"

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

var sLog = logger.NewLogger("coa.runtime")

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

	var rootResource string
	var version string
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		rootResource = parts[0]
		version = parts[1]
	} else {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Solution name is invalid in the request (%s)", name), v1alpha2.BadRequest)
	}

	sLog.Info("  M (Solution manager): delete state >>>>>>>>>>>>>>>>>>>>parts  %v, %v", rootResource, version)

	id := rootResource + "-" + version
	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"namespace":    namespace,
			"group":        model.SolutionGroup,
			"version":      "v1",
			"resource":     "solutions",
			"kind":         "Solution",
			"rootResource": rootResource,
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

	sLog.Info("  M (Solution manager): debug upsert state >>>>>>>>>>>>>>>>>>>>  %v, %v, %v", state.Spec.Version, state.Spec.RootResource, name)
	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	var rootResource string
	version := state.Spec.Version
	if state.Spec.RootResource == "" && version != "" {
		suffix := "-" + version
		rootResource = strings.TrimSuffix(name, suffix)
	} else {
		rootResource = state.Spec.RootResource
	}

	if state.ObjectMeta.Labels == nil {
		state.ObjectMeta.Labels = make(map[string]string)
	}

	_, versionLabelExists := state.ObjectMeta.Labels["version"]
	_, rootLabelExists := state.ObjectMeta.Labels["rootResource"]
	refreshLabels := false
	if !versionLabelExists || !rootLabelExists {
		sLog.Info("  M (Solution manager): update labels to true >>>>>>>>>>>>>>>>>>>>  %v, %v", rootResource, version)

		state.ObjectMeta.Labels["rootResource"] = rootResource
		state.ObjectMeta.Labels["version"] = version
		refreshLabels = true
	}

	sLog.Info("  M (Solution manager): debug refresh >>>>>>>>>>>>>>>>>>>>  %v, %v, %v", refreshLabels, versionLabelExists, rootLabelExists)

	body := map[string]interface{}{
		"apiVersion": model.SolutionGroup + "/v1",
		"kind":       "Solution",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
		},
		Metadata: map[string]interface{}{
			"namespace":     state.ObjectMeta.Namespace,
			"group":         model.SolutionGroup,
			"version":       "v1",
			"resource":      "solutions",
			"kind":          "Solution",
			"rootResource":  rootResource,
			"refreshLabels": strconv.FormatBool(refreshLabels),
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

	sLog.Info("  M (Solution manager): debug get state >>>>>>>>>>>>>>>>>>>>  %v, %v", id, namespace)

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

func (t *SolutionsManager) GetLatestState(ctx context.Context, id string, namespace string) (model.SolutionState, error) {
	ctx, span := observability.StartSpan("Solutions Manager", ctx, &map[string]string{
		"method": "GetLatest",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  M (Solution manager): debug get latest state >>>>>>>>>>>>>>>>>>>>  %v, %v", id, namespace)

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
	target, err := t.StateProvider.GetLatest(ctx, getRequest)
	if err != nil {
		return model.SolutionState{}, err
	}

	ret, err := getSolutionState(target.Body)
	if err != nil {
		return model.SolutionState{}, err
	}
	return ret, nil
}
