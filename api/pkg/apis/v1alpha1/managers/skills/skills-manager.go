/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package skills

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

type SkillsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SkillsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *SkillsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugfCtx(ctx, " M (Skills): DeleteState, name: %s", name)
	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "skills",
			"kind":      "Skill",
		},
	})
	if err != nil {
		log.ErrorfCtx(ctx, " M (Skills): failed to delete state, name: %s, err: %v", name, err)
	}
	return err
}

func (t *SkillsManager) UpsertState(ctx context.Context, name string, state model.SkillState) error {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.DebugfCtx(ctx, " M (Skills): UpsertState, name: %s", name)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.AIGroup + "/v1",
				"kind":       "skill",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "skills",
			"kind":      "Skill",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Skills): failed to UpsertSpec, name: %s, err: %v", name, err)
		return err
	}
	return nil
}

func (t *SkillsManager) ListState(ctx context.Context, namespace string) ([]model.SkillState, error) {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugCtx(ctx, " M (Skills): ListState")
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.AIGroup,
			"resource":  "skills",
			"kind":      "Skill",
			"namespace": namespace,
		},
	}
	var models []states.StateEntry
	models, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Skills): failed to list state, err: %v", err)
		return nil, err
	}
	ret := make([]model.SkillState, 0)
	for _, t := range models {
		var rt model.SkillState
		rt, err = getSkillState(t.Body)
		if err != nil {
			log.ErrorfCtx(ctx, " M (Models): failed to get skill state, err: %v", err)
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSkillState(body interface{}) (model.SkillState, error) {
	var skillState model.SkillState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &skillState)
	if err != nil {
		return model.SkillState{}, err
	}
	if skillState.Spec == nil {
		skillState.Spec = &model.SkillSpec{}
	}
	return skillState, nil
}

func (t *SkillsManager) GetState(ctx context.Context, name string, namespace string) (model.SkillState, error) {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugfCtx(ctx, " M (Skills): GetState, name: %s", name)
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.AIGroup,
			"resource":  "skills",
			"namespace": namespace,
			"kind":      "Skill",
		},
	}
	var m states.StateEntry
	m, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Skills): failed to get state, name: %s, err: %v", name, err)
		return model.SkillState{}, err
	}
	var ret model.SkillState
	ret, err = getSkillState(m.Body)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Skills): failed to get skill state, name: %s, err: %v", name, err)
		return model.SkillState{}, err
	}
	return ret, nil
}
