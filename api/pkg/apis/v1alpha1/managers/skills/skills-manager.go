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

	log.Debugf(" M (Skills): DeleteSpec, name: %s, traceId: %s", name, span.SpanContext().TraceID().String())
	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": "",
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "skills",
			"kind":      "Skill",
		},
	})
	if err != nil {
		log.Errorf(" M (Skills): failed to DeleteSpec, name: %s, err: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
	}
	return err
}

func (t *SkillsManager) UpsertSpec(ctx context.Context, name string, spec model.SkillSpec) error {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Debugf(" M (Skills): UpsertSpec, name: %s, traceId: %s", name, span.SpanContext().TraceID().String())
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
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion": "%s/v1", "kind": "Skill", "metadata": {"name": "${{$skill()}}"}}`, model.AIGroup),
			"namespace": "",
			"group":     model.AIGroup,
			"version":   "v1",
			"resource":  "skills",
			"kind":      "Skill",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.Errorf(" M (Skills): failed to UpsertSpec, name: %s, err: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}

func (t *SkillsManager) ListSpec(ctx context.Context) ([]model.SkillState, error) {
	ctx, span := observability.StartSpan("Skills Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Debugf(" M (Skills): ListSpec, traceId: %s", span.SpanContext().TraceID().String())
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "skills",
		},
	}
	models, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		log.Errorf(" M (Skills): failed to ListSpec, err: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	ret := make([]model.SkillState, 0)
	for _, t := range models {
		var rt model.SkillState
		rt, err = getSkillState(t.ID, t.Body, t.ETag)
		if err != nil {
			log.Errorf(" M (Models): failed to getSkillState, err: %v, traceId: %s", err, span.SpanContext().TraceID().String())
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

	log.Debugf(" M (Skills): GetSpec, name: %s, traceId: %s", id, span.SpanContext().TraceID().String())
	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "skills",
		},
	}
	m, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.Errorf(" M (Skills): failed to GetSpec, name: %s, err: %v, traceId: %s", id, err, span.SpanContext().TraceID().String())
		return model.SkillState{}, err
	}

	ret, err := getSkillState(id, m.Body, m.ETag)
	if err != nil {
		log.Errorf(" M (Skills): failed to getSkillState, name: %s, err: %v, traceId: %s", id, err, span.SpanContext().TraceID().String())
		return model.SkillState{}, err
	}
	return ret, nil
}
