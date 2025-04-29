/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

type SummaryManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SummaryManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(ctx, config, providers)
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

func (s *SummaryManager) GetDeploymentState(ctx context.Context, instance string, namespace string) *model.SolutionManagerDeploymentState {
	state, err := s.StateProvider.Get(ctx, states.GetRequest{
		ID: instance,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentState,
		},
	})
	if err == nil {
		var managerState model.SolutionManagerDeploymentState
		jData, _ := json.Marshal(state.Body)
		err = json.Unmarshal(jData, &managerState)
		if err == nil {
			return &managerState
		}
	}
	log.InfofCtx(ctx, " M (Summary): failed to get previous state for instance %s in namespace %s: %+v", instance, namespace, err)
	return nil
}

func (s *SummaryManager) UpsertDeploymentState(ctx context.Context, instance string, namespace string, deployment model.DeploymentSpec, mergedState model.DeploymentState) error {
	_, err := s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID: instance,
			Body: model.SolutionManagerDeploymentState{
				Spec:  deployment,
				State: mergedState,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentState,
		},
	})
	return err
}

func (s *SummaryManager) DeleteDeploymentState(ctx context.Context, instance string, namespace string) error {
	err := s.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: instance,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  DeploymentState,
		},
	})
	return err
}

func (s *SummaryManager) GetSummary(ctx context.Context, summaryId string, name string, namespace string) (model.SummaryResult, error) {
	ctx, span := observability.StartSpan("Summary Manager", ctx, &map[string]string{
		"method": "GetSummary",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, " M (Summary): get summary, name: %s, summaryId: %s, namespace: %s", name, summaryId, namespace)

	var state states.StateEntry
	state, err = s.StateProvider.Get(ctx, states.GetRequest{
		ID: summaryId,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})
	if err != nil && api_utils.IsNotFound(err) && name != "" {
		// if get summary by guid not found, try to get the summary by name
		log.InfofCtx(ctx, " M (Summary): failed to get deployment summary[%s] by summaryId, error: %+v. Try to get summary by name", summaryId, err)
		state, err = s.StateProvider.Get(ctx, states.GetRequest{
			ID: fmt.Sprintf("%s-%s", "summary", name),
			Metadata: map[string]interface{}{
				"namespace": namespace,
				"group":     model.SolutionGroup,
				"version":   "v1",
				"resource":  Summary,
			},
		})
	}
	if err != nil {
		log.ErrorfCtx(ctx, " M (Summary): failed to get deployment summary[%s]: %+v", summaryId, err)
		return model.SummaryResult{}, err
	}

	var result model.SummaryResult
	jData, _ := json.Marshal(state.Body)
	err = json.Unmarshal(jData, &result)
	if err != nil {
		log.ErrorfCtx(ctx, " M (Summary): failed to deserailze deployment summary[%s]: %+v", summaryId, err)
		return model.SummaryResult{}, err
	}

	log.InfofCtx(ctx, " M (Summary): get summary, key: %s, namespace: %s, summary: %+v", summaryId, namespace, result)
	return result, nil
}

func (s *SummaryManager) ListSummary(ctx context.Context, namespace string) ([]model.SummaryResult, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"resource":  "Summary",
			"group":     model.SolutionGroup,
		},
	}
	var entries []states.StateEntry
	entries, _, err := s.StateProvider.List(ctx, listRequest)
	if err != nil {
		return []model.SummaryResult{}, nil
	}

	var summaries []model.SummaryResult
	for _, entry := range entries {
		var result model.SummaryResult
		jData, _ := json.Marshal(entry.Body)
		err = json.Unmarshal(jData, &result)
		if err == nil {
			result.SummaryId = entry.ID
			summaries = append(summaries, result)
		}
	}
	return summaries, nil
}

func (s *SummaryManager) UpsertSummary(ctx context.Context, summaryId string, generation string, hash string, summary model.SummarySpec, state model.SummaryState, namespace string) error {
	oldSummary, err := s.GetSummary(ctx, summaryId, "", namespace)
	if err != nil && !v1alpha2.IsNotFound(err) {
		log.ErrorfCtx(ctx, " M (Summary): failed to get previous summary: %+v", err)
		return err
	} else if err == nil {
		if summary.JobID != "" && oldSummary.Summary.JobID != "" {
			var newId, oldId int64
			newId, err = strconv.ParseInt(summary.JobID, 10, 64)
			if err != nil {
				log.ErrorfCtx(ctx, " M (Summary): failed to parse new job id: %+v", err)
				return v1alpha2.NewCOAError(err, "failed to parse new job id", v1alpha2.BadRequest)
			}
			oldId, err = strconv.ParseInt(oldSummary.Summary.JobID, 10, 64)
			if err == nil && oldId > newId {
				errMsg := fmt.Sprintf("old job id %d is greater than new job id %d", oldId, newId)
				log.ErrorfCtx(ctx, " M (Summary): %s", errMsg)
				return v1alpha2.NewCOAError(err, errMsg, v1alpha2.BadRequest)
			}
		} else {
			log.WarnfCtx(ctx, " M (Summary): JobIDs are both empty, skip id check")
		}
	}
	_, err = s.StateProvider.Upsert(ctx, states.UpsertRequest{
		Value: states.StateEntry{
			ID: summaryId,
			Body: model.SummaryResult{
				Summary:        summary,
				Generation:     generation,
				Time:           time.Now().UTC(),
				State:          state,
				DeploymentHash: hash,
			},
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})
	return err
}

func (s *SummaryManager) DeleteSummary(ctx context.Context, summaryId string, namespace string, softDelete bool) error {
	ctx, span := observability.StartSpan("Summary Manager", ctx, &map[string]string{
		"method": "DeleteSummary",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if softDelete {
		// soft delete summary
		log.InfofCtx(ctx, " M (Summary): soft delete summary, key: %s, namespace: %s", summaryId, namespace)

		result, err := s.GetSummary(ctx, fmt.Sprintf("%s-%s", "summary", summaryId), "", namespace)
		if err != nil {
			log.InfofCtx(ctx, " M (Summary): failed to soft delete summary due to get summary error: %s", err.Error())
			return err
		}

		result.Summary.Removed = true
		return s.UpsertSummary(ctx,
			fmt.Sprintf("%s-%s", "summary", summaryId),
			result.Generation,
			result.DeploymentHash,
			result.Summary,
			result.State,
			namespace,
		)
	}

	log.InfofCtx(ctx, " M (Summary): delete summary, key: %s, namespace: %s", summaryId, namespace)

	err = s.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: summaryId,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})

	if err != nil {
		if api_utils.IsNotFound(err) {
			log.DebugfCtx(ctx, " M (Summary): DeleteSummary NoutFound, id: %s, namespace: %s", summaryId, namespace)
			return nil
		}
		log.ErrorfCtx(ctx, " M (Summary): failed to get summary[%s]: %+v", summaryId, err)
		return err
	}

	return nil
}
