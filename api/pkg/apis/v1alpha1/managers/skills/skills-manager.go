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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"

	observability "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type SkillsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SkillsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *SkillsManager) DeleteSpec(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.AIGroup,
			"version":  "v1",
			"resource": "skills",
		},
	})
	return err
}

func (t *SkillsManager) UpsertSpec(ctx context.Context, name string, spec model.SkillSpec) error {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.AIGroup + "/v1",
				"kind":       "skill",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion": "%s/v1", "kind": "Skill", "metadata": {"name": "${{$skill()}}"}}`, model.AIGroup),
			"scope":    "",
			"group":    model.AIGroup,
			"version":  "v1",
			"resource": "skills",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

func (t *SkillsManager) ListSpec(ctx context.Context) ([]model.SkillState, error) {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "skills",
		},
	}
	models, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SkillState, 0)
	for _, t := range models {
		var rt model.SkillState
		rt, err = getSkillState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getSkillState(id string, body interface{}, etag string) (model.SkillState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]

	j, _ := json.Marshal(spec)
	var rSpec model.SkillSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.SkillState{}, err
	}
	//rSpec.Generation??
	state := model.SkillState{
		Id:   id,
		Spec: &rSpec,
	}
	return state, nil
}

func (t *SkillsManager) GetSpec(ctx context.Context, id string) (model.SkillState, error) {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "skills",
		},
	}
	m, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SkillState{}, err
	}

	ret, err := getSkillState(id, m.Body, m.ETag)
	if err != nil {
		return model.SkillState{}, err
	}
	return ret, nil
}
