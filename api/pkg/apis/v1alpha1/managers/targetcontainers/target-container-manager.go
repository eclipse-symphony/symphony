/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package targetcontainers

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

type TargetContainersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *TargetContainersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *TargetContainersManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("TargetContainersManager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targetcontainers",
			"kind":      "TargetContainer",
		},
	})
	return err
}

func (t *TargetContainersManager) UpsertState(ctx context.Context, name string, state model.TargetContainerState) error {
	ctx, span := observability.StartSpan("TargetContainersManager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "TargetContainer",
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
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targetcontainers",
			"kind":      "TargetContainer",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *TargetContainersManager) ListState(ctx context.Context, namespace string) ([]model.TargetContainerState, error) {
	ctx, span := observability.StartSpan("TargetContainersManager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targetcontainers",
			"namespace": namespace,
			"kind":      "TargetContainer",
		},
	}
	var targetcontainers []states.StateEntry
	targetcontainers, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.TargetContainerState, 0)
	for _, t := range targetcontainers {
		var rt model.TargetContainerState
		rt, err = getTargetContainerState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getTargetContainerState(body interface{}, etag string) (model.TargetContainerState, error) {
	var TargetContainerState model.TargetContainerState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &TargetContainerState)
	if err != nil {
		return model.TargetContainerState{}, err
	}
	if TargetContainerState.Spec == nil {
		TargetContainerState.Spec = &model.TargetContainerSpec{}
	}
	return TargetContainerState, nil
}

func (t *TargetContainersManager) GetState(ctx context.Context, id string, namespace string) (model.TargetContainerState, error) {
	ctx, span := observability.StartSpan("TargetContainersManager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targetcontainers",
			"namespace": namespace,
			"kind":      "TargetContainer",
		},
	}
	var Target states.StateEntry
	Target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.TargetContainerState{}, err
	}
	var ret model.TargetContainerState
	ret, err = getTargetContainerState(Target.Body, Target.ETag)
	if err != nil {
		return model.TargetContainerState{}, err
	}
	return ret, nil
}
