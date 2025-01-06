/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

const (
	// DefaultSummaryRetentionDuration is the default time to cleanup deprecated summaries
	// DefaultSummaryRetentionDuration = 180 * time.Hour * 24
	DefaultSummaryRetentionDuration = 60 * time.Second * 5
)

type SummaryCleanupManager struct {
	managers.Manager
	StateProvider            states.IStateProvider
	SummaryRetentionDuration time.Duration
}

func (s *SummaryCleanupManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

	// Set activation cleanup interval after they are done. If not set, use default 180 days.
	if val, ok := config.Properties["SummaryRetentionDuration"]; ok {
		s.SummaryRetentionDuration, err = time.ParseDuration(val)
		if err != nil {
			return v1alpha2.NewCOAError(nil, "SummaryRetentionDuration cannot be parsed, please enter a valid duration", v1alpha2.BadConfig)
		} else if s.SummaryRetentionDuration < 0 {
			return v1alpha2.NewCOAError(nil, "SummaryRetentionDuration cannot be negative", v1alpha2.BadConfig)
		}
	} else {
		s.SummaryRetentionDuration = DefaultSummaryRetentionDuration
	}

	log.Info("M (Summary Cleanup): Initialize SummaryRetentionDuration as " + s.SummaryRetentionDuration.String())
	return nil
}

func (s *SummaryCleanupManager) Enabled() bool {
	return true
}

func (s *SummaryCleanupManager) Poll() []error {
	// TODO: initialize the context with id correctly
	ctx, span := observability.StartSpan("Summary Cleanup Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.InfoCtx(ctx, "M (Summary Cleanup): Polling summaries")
	summaries, err := s.ListSummary(ctx, "")
	if err != nil {
		return []error{err}
	}
	ret := []error{}
	for _, summary := range summaries {
		// Check the upsert time of summary.
		duration := time.Since(summary.Time)
		if duration > s.SummaryRetentionDuration {
			log.InfofCtx(ctx, "M (Summary Cleanup): Deleting summary %s since it has deprecated for %s", summary.SummaryId, duration.String())
			err = s.DeleteSummary(ctx, summary.SummaryId, "")
			if err != nil {
				ret = append(ret, err)
			}
		}
	}
	return ret
}

func (s *SummaryCleanupManager) Reconcil() []error {
	return nil
}

func (s *SummaryCleanupManager) ListSummary(ctx context.Context, namespace string) ([]model.SummaryResult, error) {
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

func (s *SummaryCleanupManager) DeleteSummary(ctx context.Context, id string, namespace string) error {
	log.InfofCtx(ctx, " M (SummaryCleanup): delete summary, key: %s, namespace: %s", id, namespace)

	err := s.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  Summary,
		},
	})

	if err != nil {
		if api_utils.IsNotFound(err) {
			log.DebugfCtx(ctx, " M (SummaryCleanup): DeleteSummary NoutFound, id: %s, namespace: %s", id, namespace)
			return nil
		}
		log.ErrorfCtx(ctx, " M (SummaryCleanup): failed to get summary[%s]: %+v", id, err)
		return err
	}

	return nil
}
