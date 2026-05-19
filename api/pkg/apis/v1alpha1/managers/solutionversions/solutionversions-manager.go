/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutionversions

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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type SolutionVersionsManager struct {
	managers.Manager
	StateProvider     states.IStateProvider
	needValidate      bool
	SolutionVersionValidator validation.SolutionVersionValidator
}

func (s *SolutionVersionsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
		// s.SolutionVersionValidator = validation.NewSolutionVersionValidator(s.solutionversionInstanceLookup, s.solutionversionContainerLookup, s.uniqueNameSolutionVersionLookup)
		s.SolutionVersionValidator = validation.NewSolutionVersionValidator(nil, nil, s.uniqueNameSolutionVersionLookup)
	}
	return nil
}

func (t *SolutionVersionsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("SolutionVersions Manager", ctx, &map[string]string{
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
			"group":     model.SolutionVersionGroup,
			"version":   "v1",
			"resource":  "solutionversions",
			"kind":      "SolutionVersion",
		},
	})
	return err
}

func (t *SolutionVersionsManager) UpsertState(ctx context.Context, name string, state model.SolutionVersionState) error {
	ctx, span := observability.StartSpan("SolutionVersions Manager", ctx, &map[string]string{
		"method": "UpsertState",
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
			state.ObjectMeta.Labels[constants.RootResource] = state.Spec.RootResource
		}
		if err = validation.ValidateCreateOrUpdateWrapper(ctx, &t.SolutionVersionValidator, state, oldState, getStateErr); err != nil {
			return err
		}
	}

	body := map[string]interface{}{
		"apiVersion": model.SolutionVersionGroup + "/v1",
		"kind":       "SolutionVersion",
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
			"group":     model.SolutionVersionGroup,
			"version":   "v1",
			"resource":  "solutionversions",
			"kind":      "SolutionVersion",
		},
	}

	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (t *SolutionVersionsManager) ListState(ctx context.Context, namespace string) ([]model.SolutionVersionState, error) {
	ctx, span := observability.StartSpan("SolutionVersions Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionVersionGroup,
			"resource":  "solutionversions",
			"namespace": namespace,
			"kind":      "SolutionVersion",
		},
	}
	var solutionversions []states.StateEntry
	solutionversions, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SolutionVersionState, 0)
	for _, t := range solutionversions {
		var rt model.SolutionVersionState
		rt, err = getSolutionVersionState(t.Body)
		if err != nil {
			return nil, err
		}
		rt.ObjectMeta.UpdateEtag(t.ETag)
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSolutionVersionState(body interface{}) (model.SolutionVersionState, error) {
	var solutionversionState model.SolutionVersionState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &solutionversionState)
	if err != nil {
		return model.SolutionVersionState{}, err
	}
	if solutionversionState.Spec == nil {
		solutionversionState.Spec = &model.SolutionVersionSpec{}
	}
	return solutionversionState, nil
}

func (t *SolutionVersionsManager) GetState(ctx context.Context, id string, namespace string) (model.SolutionVersionState, error) {
	ctx, span := observability.StartSpan("SolutionVersions Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionVersionGroup,
			"resource":  "solutionversions",
			"namespace": namespace,
			"kind":      "SolutionVersion",
		},
	}
	var entry states.StateEntry
	entry, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SolutionVersionState{}, err
	}
	var ret model.SolutionVersionState
	ret, err = getSolutionVersionState(entry.Body)
	if err != nil {
		return model.SolutionVersionState{}, err
	}
	ret.ObjectMeta.UpdateEtag(entry.ETag)
	return ret, nil
}

func (t *SolutionVersionsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.SolutionVersionValidator, state, err)
}

func (t *SolutionVersionsManager) solutionversionInstanceLookup(ctx context.Context, name string, namespace string) (bool, error) {
	instanceList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.Instance, namespace, map[string]string{constants.SolutionVersion: name}, 1)
	if err != nil {
		return false, err
	}
	return len(instanceList) > 0, nil
}

func (t *SolutionVersionsManager) solutionversionContainerLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.Solution, name, namespace)
}

func (t *SolutionVersionsManager) uniqueNameSolutionVersionLookup(ctx context.Context, displayName string, namespace string) (interface{}, error) {
	return states.GetObjectStateWithUniqueName(ctx, t.StateProvider, validation.SolutionVersion, displayName, namespace)
}
