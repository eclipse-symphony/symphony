/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var lock sync.Mutex

var log = logger.NewLogger("coa.runtime")

type ActivationsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *ActivationsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
	return nil
}

func (m *ActivationsManager) GetState(ctx context.Context, name string, namespace string) (model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": namespace,
			"kind":      "Activation",
		},
	}
	var entry states.StateEntry
	entry, err = m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.ActivationState{}, err
	}
	var ret model.ActivationState
	ret, err = getActivationState(entry.Body, entry.ETag)
	if err != nil {
		return model.ActivationState{}, err
	}
	return ret, nil
}

func getActivationState(body interface{}, etag string) (model.ActivationState, error) {
	var activationState model.ActivationState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &activationState)
	if err != nil {
		return model.ActivationState{}, err
	}
	if activationState.Spec == nil {
		activationState.Spec = &model.ActivationSpec{}
	}
	activationState.ObjectMeta.Generation = etag
	if activationState.Status == nil {
		activationState.Status = &model.ActivationStatus{}
	}
	return activationState, nil
}

func (m *ActivationsManager) UpsertState(ctx context.Context, name string, state model.ActivationState) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
			ETag: state.ObjectMeta.Generation,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *ActivationsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	return err
}

func (t *ActivationsManager) ListState(ctx context.Context, namespace string) ([]model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": namespace,
			"kind":      "Activation",
		},
	}
	var activations []states.StateEntry
	activations, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.ActivationState, 0)
	for _, t := range activations {
		var rt model.ActivationState
		rt, err = getActivationState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
func (t *ActivationsManager) ReportStatus(ctx context.Context, name string, namespace string, current model.ActivationStatus) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ReportStatus",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	lock.Lock()
	defer lock.Unlock()

	var activationState model.ActivationState
	activationState, err = t.GetState(ctx, name, namespace)
	if err != nil {
		return err
	}

	current.UpdateTime = time.Now().Format(time.RFC3339) // TODO: is this correct? Shouldn't it be reported?
	activationState.Status = &current

	var entry states.StateEntry
	entry.ID = activationState.ObjectMeta.Name
	entry.Body = activationState

	upsertRequest := states.UpsertRequest{
		Value: entry,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": activationState.ObjectMeta.Namespace,
			"kind":      "Activation",
		},
		Options: states.UpsertOption{
			UpdateStatusOnly: true,
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *ActivationsManager) ReportStageStatus(ctx context.Context, name string, namespace string, current model.StageStatus) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ReportStageStatus",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	lock.Lock()
	defer lock.Unlock()

	var activationState model.ActivationState
	activationState, err = t.GetState(ctx, name, namespace)
	if err != nil {
		return err
	}

	activationState.Status.UpdateTime = time.Now().Format(time.RFC3339) // TODO: is this correct? Shouldn't it be reported?

	err = mergeStageStatus(&activationState, current)
	if err != nil {
		return err
	}

	var entry states.StateEntry
	entry.ID = activationState.ObjectMeta.Name
	entry.Body = activationState

	upsertRequest := states.UpsertRequest{
		Value: entry,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.WorkflowGroup,
			"resource":  "activations",
			"namespace": activationState.ObjectMeta.Namespace,
			"kind":      "Activation",
		},
		Options: states.UpsertOption{
			UpdateStatusOnly: true,
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func mergeStageStatus(activationState *model.ActivationState, current model.StageStatus) error {
	if current.Outputs["__site"] == nil {
		// The StageStatus is triggered locally
		if activationState.Status.StageHistory == nil {
			activationState.Status.StageHistory = make([]model.StageStatus, 0)
		}
		if len(activationState.Status.StageHistory) == 0 {
			activationState.Status.StageHistory = append(activationState.Status.StageHistory, current)
		} else if activationState.Status.StageHistory[len(activationState.Status.StageHistory)-1].Stage != current.Stage {
			activationState.Status.StageHistory = append(activationState.Status.StageHistory, current)
		} else {
			activationState.Status.StageHistory[len(activationState.Status.StageHistory)-1] = current
		}
	} else {
		// The StageStatus is triggered remotely
		if activationState.Status.StageHistory == nil || len(activationState.Status.StageHistory) == 0 {
			err := v1alpha2.NewCOAError(nil, "activation status doesn't has a parent stage history", v1alpha2.BadRequest)
			return err
		}
		parentStageStatus := &activationState.Status.StageHistory[len(activationState.Status.StageHistory)-1]
		stage := utils.FormatAsString(current.Outputs["__stage"])
		if stage != parentStageStatus.Stage {
			err := v1alpha2.NewCOAError(nil, "remote job result doesn't match the latest stage, discard the result", v1alpha2.BadRequest)
			return err
		}
		site := utils.FormatAsString(current.Outputs["__site"])
		parentStageStatus.Outputs[fmt.Sprintf("%s.__status", site)] = current.Status.String()
		output := map[string]interface{}{}
		for k, v := range current.Outputs {
			// remove outputs for internal tracking use
			if !strings.HasPrefix(k, "__") {
				output[k] = v
			}
		}
		outputBytes, _ := json.Marshal(output)
		parentStageStatus.Outputs[fmt.Sprintf("%s.__output", site)] = string(outputBytes)
		parentStageStatus.Status = v1alpha2.Done
		for k, v := range parentStageStatus.Outputs {
			if strings.HasSuffix(k, "__status") {
				if v != v1alpha2.Done.String() {
					parentStageStatus.Status = v1alpha2.Paused
					break
				}
			}
		}
		parentStageStatus.StatusMessage = parentStageStatus.Status.String()
	}

	latestStage := &activationState.Status.StageHistory[len(activationState.Status.StageHistory)-1]
	if latestStage.Status == v1alpha2.Done && latestStage.NextStage == "" {
		activationState.Status.Status = v1alpha2.Done
	} else {
		activationState.Status.Status = v1alpha2.Running
	}
	activationState.Status.StatusMessage = activationState.Status.Status.String()
	return nil
}
