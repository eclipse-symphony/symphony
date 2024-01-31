/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package devices

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

type DevicesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *DevicesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		log.Errorf(" M (Devices): failed to initialize manager %+v", err)
		return err
	}
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		log.Errorf(" M (Devices): failed to get state provider %+v", err)
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *DevicesManager) DeleteSpec(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): DeleteSpec name %s, traceId: %s", name, span.SpanContext().TraceID().String())

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "devices",
		},
	})
	if err != nil {
		log.Errorf(" M (Devices):failed to delete state %s, error: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}

func (t *DevicesManager) UpsertSpec(ctx context.Context, name string, spec model.DeviceSpec) error {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): UpsertSpec name %s, traceId: %s", name, span.SpanContext().TraceID().String())

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "device",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion": "%s/v1", "kind": "Device", "metadata": {"name": "${{$device()}}"}}`, model.FabricGroup),
			"scope":    "",
			"group":    model.FabricGroup,
			"version":  "v1",
			"resource": "devices",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to update state %s, error: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}

func (t *DevicesManager) ListSpec(ctx context.Context) ([]model.DeviceState, error) {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): ListSpec, traceId: %s", span.SpanContext().TraceID().String())

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "devices",
		},
	}
	solutions, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to list state, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	ret := make([]model.DeviceState, 0)
	for _, t := range solutions {
		var rt model.DeviceState
		rt, err = getDeviceState(t.ID, t.Body)
		if err != nil {
			log.Errorf(" M (Devices): ListSpec failed to get device state %s, error: %v, traceId: %s", t.ID, err, span.SpanContext().TraceID().String())
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getDeviceState(id string, body interface{}) (model.DeviceState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]

	j, _ := json.Marshal(spec)
	var rSpec model.DeviceSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.DeviceState{}, err
	}
	state := model.DeviceState{
		Id:   id,
		Spec: &rSpec,
	}
	return state, nil
}

func (t *DevicesManager) GetSpec(ctx context.Context, id string) (model.DeviceState, error) {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): GetSpec id %s, traceId: %s", id, span.SpanContext().TraceID().String())

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FabricGroup,
			"resource": "devices",
		},
	}
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to get state %s, error: %v, traceId: %s", id, err, span.SpanContext().TraceID().String())
		return model.DeviceState{}, err
	}

	ret, err := getDeviceState(id, target.Body)
	if err != nil {
		log.Errorf(" M (Devices): GetSpec failed to get device state, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return model.DeviceState{}, err
	}
	return ret, nil
}
