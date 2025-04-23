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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutioncontainers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

const (
	// DefaultSummaryRetentionDuration is the default time to cleanup deprecated summaries
	// DefaultSummaryRetentionDuration = 180 * time.Hour * 24
	DefaultResourceCountRetentionDuration = 60 * time.Second * 5
)

type ResourceCountManager struct {
	SolutionsManager          *solutions.SolutionsManager
	TargetsManager            *targets.TargetsManager
	InstancesManager          *instances.InstancesManager
	SolutionContainersManager *solutioncontainers.SolutionContainersManager
	StateProvider             states.IStateProvider
}

func (s *ResourceCountManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.SolutionsManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}
	err = s.TargetsManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}
	err = s.InstancesManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}
	err = s.SolutionContainersManager.Init(ctx, config, providers)
	if err != nil {
		return err
	}
	return nil
}

func (s *ResourceCountManager) Poll() []error {
	// TODO: initialize the context with id correctly
	ctx, span := observability.StartSpan("Resource Count Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.InfoCtx(ctx, "M (Inventory Count): Polling summaries")

	// Initialize a map to store resource counts
	resourceCounts := make(map[string]int)

	// Call TargetsManager.ListState
	targetStates, err := s.TargetsManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list targets: %v", err)
		return []error{err}
	}
	resourceCounts["targets"] = len(targetStates)

	// Call InstancesManager.ListState (if available)
	instanceStates, err := s.InstancesManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list instances: %v", err)
		return []error{err}
	}
	resourceCounts["instances"] = len(instanceStates)

	// Call SolutionsManager.ListState (if available)
	solutionStates, err := s.SolutionsManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list solutions: %v", err)
		return []error{err}
	}
	resourceCounts["solutions"] = len(solutionStates)

	// Call SolutionContainersManager.ListState (if available)
	solutionContainerStates, err := s.SolutionContainersManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list solution containers: %v", err)
		return []error{err}
	}
	resourceCounts["solutioncontainers"] = len(solutionContainerStates)

	// Log the counts as a single string
	log.InfofCtx(ctx, fmt.Sprintf(
		"M (Inventory Count): Summary: Found %d instances, %d targets, %d solutions, and %d solution containers across all namespaces",
		resourceCounts["instances"], resourceCounts["targets"], resourceCounts["solutions"], resourceCounts["solutioncontainers"],
	))

	return nil
}

func (s *ResourceCountManager) Shutdown(ctx context.Context) error {
	log.InfoCtx(ctx, "Shutting down ResourceCountManager")
	// 如果有需要清理的资源，可以在这里处理
	return nil
}

// func (s *ResourceCountManager) Reconcil() []error {
// 	return nil
// }

// // getResourceCounts is a helper function to get resource counts across all namespaces
// func (s *ResourceCountManager) getResourceCounts(ctx context.Context) (map[string]int, error) {
// 	// Define the resource types and their metadata
// 	resourceMetadata := []struct {
// 		Resource string
// 		Group    string
// 		Version  string
// 	}{
// 		{"instances", "solution.symphony", "v1"},
// 		{"targets", "fabric.symphony", "v1"},
// 		{"solutions", "solution.symphony", "v1"},
// 		{"solutioncontainers", "solution.symphony", "v1"},
// 	}

// 	// Initialize a map to store resource counts
// 	resourceCounts := make(map[string]int)

// 	// Iterate over each resource type and call the StateProvider's List method
// 	for _, metadata := range resourceMetadata {
// 		listRequest := states.ListRequest{
// 			Metadata: map[string]interface{}{
// 				"group":    metadata.Group,
// 				"version":  metadata.Version,
// 				"resource": metadata.Resource,
// 			},
// 		}

// 		// Use the provider's List method to fetch resources
// 		entities, _, err := s.StateProvider.List(ctx, listRequest)
// 		if err != nil {
// 			log.ErrorfCtx(ctx, "Failed to list %s: %v", metadata.Resource, err)
// 			return nil, fmt.Errorf("failed to list %s: %w", metadata.Resource, err)
// 		}

// 		// Store the count of resources in the map
// 		resourceCounts[metadata.Resource] = len(entities)
// 	}

// 	return resourceCounts, nil
// }

// func (s *SummaryCleanupManager) Reconcil() []error {
// 	return nil
// }
