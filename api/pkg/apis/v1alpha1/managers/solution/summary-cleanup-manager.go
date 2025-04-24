/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

const (
	// DefaultSummaryRetentionDuration is the default time to cleanup deprecated summaries
	DefaultSummaryRetentionDuration = 180 * time.Hour * 24
)

type SummaryCleanupManager struct {
	SummaryManager
	SummaryRetentionDuration time.Duration
}

func (s *SummaryCleanupManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.SummaryManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}

	// Set activation cleanup interval after they are done. If not set, use default 180 days.
	if val, ok := config.Properties["RetentionDuration"]; ok {
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
			err = s.DeleteSummary(ctx, summary.SummaryId, "", false)
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
