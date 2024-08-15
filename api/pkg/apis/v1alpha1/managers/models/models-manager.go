/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package models

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

type ModelsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *ModelsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *ModelsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Models Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugfCtx(ctx, " M (Models): DeleteState, name: %s", name)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "models",
			"kind":      "Model",
		},
	})

	if err != nil {
		log.ErrorfCtx(ctx, " M (Models): failed to delete state, name: %s, err: %v", name, err)
	}
	return err
}

func (t *ModelsManager) UpsertState(ctx context.Context, name string, state model.ModelState) error {
	ctx, span := observability.StartSpan("Models Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Models): UpsertState, name: %s", name)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.AIGroup + "/v1",
				"kind":       "model",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "models",
			"kind":      "Model",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Models): failed to UpsertSpec, name: %s, err: %v", name, err)
		return err
	}
	return nil
}

func (t *ModelsManager) ListState(ctx context.Context, namespace string) ([]model.ModelState, error) {
	ctx, span := observability.StartSpan("Models Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugCtx(ctx, " M (Models): ListState")
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.AIGroup,
			"resource":  "models",
			"kind":      "Model",
			"namespace": namespace,
		},
	}
	var models []states.StateEntry
	models, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Models): failed to ListState, err: %v", err)
		return nil, err
	}
	ret := make([]model.ModelState, 0)
	for _, t := range models {
		var rt model.ModelState
		rt, err = getModelState(t.Body)
		if err != nil {
			log.ErrorfCtx(ctx, " M (Models): failed to getModelState, err: %v", err)
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getModelState(body interface{}) (model.ModelState, error) {
	var modelState model.ModelState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &modelState)
	if err != nil {
		return model.ModelState{}, err
	}
	if modelState.Spec == nil {
		modelState.Spec = &model.ModelSpec{}
	}
	return modelState, nil
}

func (t *ModelsManager) GetState(ctx context.Context, name string, namespace string) (model.ModelState, error) {
	ctx, span := observability.StartSpan("Models Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugfCtx(ctx, " M (Models): GetState, name: %s", name)
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.AIGroup,
			"resource":  "models",
			"namespace": namespace,
			"kind":      "Model",
		},
	}
	var m states.StateEntry
	m, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Models): failed to GetSpec, name: %s, err: %v", name, err)
		return model.ModelState{}, err
	}

	var ret model.ModelState
	ret, err = getModelState(m.Body)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Models): failed to getModelState, name: %s, err: %v", name, err)
		return model.ModelState{}, err
	}
	return ret, nil
}
