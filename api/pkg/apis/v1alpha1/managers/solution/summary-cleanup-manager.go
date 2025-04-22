/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

const (
	// DefaultSummaryRetentionDuration is the default time to cleanup deprecated summaries
	// DefaultSummaryRetentionDuration = 180 * time.Hour * 24
	DefaultSummaryRetentionDuration = 60 * time.Second * 5
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

	// Get resource counts
	resourceCounts, err := s.getResourceCounts(ctx)
	if err != nil {
		log.ErrorCtx(ctx, "Failed to get resource counts: %v", err)
		return []error{err}
	}

	// Log the counts as a single string
	log.InfofCtx(ctx, fmt.Sprintf(
		"M (Summary Cleanup): Summary: Found %d instances, %d targets, %d solutions, and %d solution containers across all namespaces",
		resourceCounts["instances"], resourceCounts["targets"], resourceCounts["solutions"], resourceCounts["solutioncontainers"],
	))

	// Proceed with summary cleanup logic
	summaries, err := s.ListSummary(ctx, "")
	if err != nil {
		return []error{err}
	}
	ret := []error{}
	for _, summary := range summaries {
		// Check the upsert time of summary.
		duration := time.Since(summary.Time)
		if duration > s.SummaryRetentionDuration {
			log.InfofCtx(ctx, fmt.Sprintf(
				"M (Summary Cleanup): Deleting summary %s since it has deprecated for %s",
				summary.SummaryId, duration.String(),
			))

			err = s.DeleteSummary(ctx, summary.SummaryId, "", false)
			if err != nil {
				ret = append(ret, err)
			}
		}
	}
	return ret
}

// getResourceCounts is a helper function to get resource counts across all namespaces
func (s *SummaryCleanupManager) getResourceCounts(ctx context.Context) (map[string]int, error) {
	// Define the resource types and their metadata
	resourceMetadata := []struct {
		Resource string
		Group    string
		Version  string
	}{
		{"instances", "solution.symphony", "v1"},
		{"targets", "fabric.symphony", "v1"},
		{"solutions", "solution.symphony", "v1"},
		{"solutioncontainers", "solution.symphony", "v1"},
	}

	// Initialize a map to store resource counts
	resourceCounts := make(map[string]int)

	// Iterate over each resource type and call the StateProvider's List method
	for _, metadata := range resourceMetadata {
		listRequest := states.ListRequest{
			Metadata: map[string]interface{}{
				"group":    metadata.Group,
				"version":  metadata.Version,
				"resource": metadata.Resource,
			},
		}

		// Use the provider's List method to fetch resources
		entities, _, err := s.StateProvider.List(ctx, listRequest)
		if err != nil {
			log.ErrorfCtx(ctx, "Failed to list %s: %v", metadata.Resource, err)
			return nil, fmt.Errorf("failed to list %s: %w", metadata.Resource, err)
		}

		// Store the count of resources in the map
		resourceCounts[metadata.Resource] = len(entities)
	}

	return resourceCounts, nil
}

func (s *SummaryCleanupManager) Reconcil() []error {
	return nil
}
