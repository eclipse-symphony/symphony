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
	if err != nil {
		log.Errorf(" M (Devices): failed to get state provider %+v", err)
		return err
	}
	s.StateProvider = stateprovider
	return nil
}

func (t *DevicesManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): DeleteState name %s, traceId: %s", name, span.SpanContext().TraceID().String())

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "devices",
			"kind":      "Device",
		},
	})
	if err != nil {
		log.Errorf(" M (Devices):failed to delete state %s, error: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}

func (t *DevicesManager) UpsertState(ctx context.Context, name string, state model.DeviceState) error {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): UpsertState name %s, traceId: %s", name, span.SpanContext().TraceID().String())

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "device",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "devices",
			"kind":      "Device",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to update state %s, error: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return err
	}
	return nil
}

func (t *DevicesManager) ListState(ctx context.Context, namespace string) ([]model.DeviceState, error) {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): ListState, traceId: %s", span.SpanContext().TraceID().String())

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "devices",
			"kind":      "Device",
			"namespace": namespace,
		},
	}
	devices, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to list state, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	ret := make([]model.DeviceState, 0)
	for _, t := range devices {
		var rt model.DeviceState
		rt, err = getDeviceState(t.ID, t.Body)
		if err != nil {
			log.Errorf(" M (Devices): ListState failed to get device state %s, error: %v, traceId: %s", t.ID, err, span.SpanContext().TraceID().String())
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getDeviceState(id string, body interface{}) (model.DeviceState, error) {
	dict := body.(map[string]interface{})

	//read spec
	spec := dict["spec"]
	j, _ := json.Marshal(spec)
	var rSpec model.DeviceSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.DeviceState{}, err
	}

	//read metadata
	metadata := dict["metadata"]
	j, _ = json.Marshal(metadata)
	var rMetadata model.ObjectMeta
	err = json.Unmarshal(j, &rMetadata)
	if err != nil {
		return model.DeviceState{}, err
	}

	//read status
	status := dict["status"]
	j, _ = json.Marshal(status)
	var rStatus map[string]string
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.DeviceState{}, err
	}

	state := model.DeviceState{
		ObjectMeta: rMetadata,
		Spec:       &rSpec,
		Status:     rStatus,
	}
	return state, nil
}

func (t *DevicesManager) GetState(ctx context.Context, name string, namespace string) (model.DeviceState, error) {
	ctx, span := observability.StartSpan("Devices Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	log.Infof(" M (Devices): GetState id %s, traceId: %s", name, span.SpanContext().TraceID().String())

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "devices",
			"namespace": namespace,
			"kind":      "Device",
		},
	}
	entry, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		log.Errorf(" M (Devices): failed to get state %s, error: %v, traceId: %s", name, err, span.SpanContext().TraceID().String())
		return model.DeviceState{}, err
	}

	ret, err := getDeviceState(name, entry.Body)
	if err != nil {
		log.Errorf(" M (Devices): GetSpec failed to get device state, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return model.DeviceState{}, err
	}
	return ret, nil
}
