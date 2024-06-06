/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package instancecontainers

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

type InstanceContainersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *InstanceContainersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *InstanceContainersManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("InstanceContainersManager", ctx, &map[string]string{
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
			"resource":  "instancecontainers",
			"kind":      "InstanceContainer",
		},
	})
	return err
}

func (t *InstanceContainersManager) UpsertState(ctx context.Context, name string, state model.InstanceContainerState) error {
	ctx, span := observability.StartSpan("InstanceContainersManager", ctx, &map[string]string{
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
		"kind":       "InstanceContainer",
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
			"resource":  "instancecontainers",
			"kind":      "InstanceContainer",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *InstanceContainersManager) ListState(ctx context.Context, namespace string) ([]model.InstanceContainerState, error) {
	ctx, span := observability.StartSpan("InstanceContainersManager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "instancecontainers",
			"namespace": namespace,
			"kind":      "InstanceContainer",
		},
	}
	var instanceContainers []states.StateEntry
	instanceContainers, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.InstanceContainerState, 0)
	for _, t := range instanceContainers {
		var rt model.InstanceContainerState
		rt, err = getInstanceContainerState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getInstanceContainerState(body interface{}, etag string) (model.InstanceContainerState, error) {
	var InstanceContainerState model.InstanceContainerState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &InstanceContainerState)
	if err != nil {
		return model.InstanceContainerState{}, err
	}
	if InstanceContainerState.Spec == nil {
		InstanceContainerState.Spec = &model.InstanceContainerSpec{}
	}
	return InstanceContainerState, nil
}

func (t *InstanceContainersManager) GetState(ctx context.Context, id string, namespace string) (model.InstanceContainerState, error) {
	ctx, span := observability.StartSpan("InstanceContainersManager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.SolutionGroup,
			"resource":  "instancecontainers",
			"namespace": namespace,
			"kind":      "InstanceContainer",
		},
	}
	var instance states.StateEntry
	instance, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.InstanceContainerState{}, err
	}
	var ret model.InstanceContainerState
	ret, err = getInstanceContainerState(instance.Body, instance.ETag)
	if err != nil {
		return model.InstanceContainerState{}, err
	}
	return ret, nil
}
