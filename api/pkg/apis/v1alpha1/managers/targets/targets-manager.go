/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package targets

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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/registry"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type TargetsManager struct {
	managers.Manager
	StateProvider    states.IStateProvider
	RegistryProvider registry.IRegistryProvider
}

func (s *TargetsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *TargetsManager) DeleteSpec(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var rootResource string
	var version string
	var id string
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		rootResource = parts[0]
		version = parts[1]
		id = rootResource + "-" + version
	} else {
		id = name
	}

	log.Info("  M (Target manager): delete state >>>>>>>>>>>>>>>>>>>>parts  %v, %v", rootResource, version)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"namespace":    namespace,
			"group":        model.FabricGroup,
			"version":      "v1",
			"resource":     "targets",
			"kind":         "Target",
			"rootResource": rootResource,
		},
	})
	return err
}

func (t *TargetsManager) UpsertState(ctx context.Context, name string, state model.TargetState) error {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})

	log.Info("  M (Target manager): upsert state >>>>>>>>>>>>>>>>>>>>parts  %v ", name)

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	var rootResource string
	var version string
	var refreshLabels bool
	if state.Spec != nil {
		log.Info("  M (Solution manager): debug upsert state >>>>>>>>>>>>>>>>>>>>  %v, %v, %v", state.Spec.Version, state.Spec.RootResource, name)

		if state.Spec.Version != "" {
			version = state.Spec.Version
		}
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
		if !versionLabelExists || !rootLabelExists {
			log.Info("  M (Target manager): update labels to true >>>>>>>>>>>>>>>>>>>>  %v, %v", rootResource, version)

			state.ObjectMeta.Labels["rootResource"] = rootResource
			state.ObjectMeta.Labels["version"] = version
			refreshLabels = true
		}
	}

	body := map[string]interface{}{
		"apiVersion": model.FabricGroup + "/v1",
		"kind":       "Target",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: state.Spec.Generation,
		},
		Metadata: map[string]interface{}{
			"namespace":     state.ObjectMeta.Namespace,
			"group":         model.FabricGroup,
			"version":       "v1",
			"resource":      "targets",
			"kind":          "Target",
			"rootResource":  rootResource,
			"refreshLabels": strconv.FormatBool(refreshLabels),
		},
	}

	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	return err
}

// Caller need to explicitly set namespace in current.Metadata!
func (t *TargetsManager) ReportState(ctx context.Context, current model.TargetState) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "ReportState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: current.ObjectMeta.Name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": current.ObjectMeta.Namespace,
		},
	}

	var target states.StateEntry
	target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		observ_utils.CloseSpanWithError(span, &err)
		return model.TargetState{}, err
	}

	var targetState model.TargetState
	bytes, _ := json.Marshal(target.Body)
	err = json.Unmarshal(bytes, &targetState)
	if err != nil {
		observ_utils.CloseSpanWithError(span, &err)
		return model.TargetState{}, err
	}

	for k, v := range current.Status.Properties {
		if targetState.Status.Properties == nil {
			targetState.Status.Properties = make(map[string]string)
		}
		targetState.Status.Properties[k] = v
	}
	targetState.Status.LastModified = current.Status.LastModified

	target.Body = targetState

	updateRequest := states.UpsertRequest{
		Value: target,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": current.ObjectMeta.Namespace,
			"kind":      "Target",
		},
		Options: states.UpsertOption{
			UpdateStateOnly: true,
		},
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return model.TargetState{}, err
	}
	return targetState, nil
}
func (t *TargetsManager) ListState(ctx context.Context, namespace string) ([]model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	var targets []states.StateEntry
	targets, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.TargetState, 0)
	for _, t := range targets {
		var rt model.TargetState
		rt, err = getTargetState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getTargetState(body interface{}, etag string) (model.TargetState, error) {
	var targetState model.TargetState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &targetState)
	if err != nil {
		return model.TargetState{}, err
	}
	if targetState.Spec == nil {
		targetState.Spec = &model.TargetSpec{}
	}
	targetState.Spec.Generation = etag
	return targetState, nil
}

func (t *TargetsManager) GetState(ctx context.Context, id string, namespace string) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	var target states.StateEntry
	target, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	var ret model.TargetState
	ret, err = getTargetState(target.Body, target.ETag)
	if err != nil {
		return model.TargetState{}, err
	}
	return ret, nil
}

func (t *TargetsManager) GetLatestState(ctx context.Context, id string, namespace string) (model.TargetState, error) {
	ctx, span := observability.StartSpan("Targets Manager", ctx, &map[string]string{
		"method": "GetLatest",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info("  M (Target manager): debug get latest state >>>>>>>>>>>>>>>>>>>>  %v, %v", id, namespace)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FabricGroup,
			"resource":  "targets",
			"namespace": namespace,
			"kind":      "Target",
		},
	}
	target, err := t.StateProvider.GetLatest(ctx, getRequest)
	if err != nil {
		return model.TargetState{}, err
	}

	ret, err := getTargetState(target.Body, target.ETag)
	if err != nil {
		return model.TargetState{}, err
	}
	return ret, nil
}
