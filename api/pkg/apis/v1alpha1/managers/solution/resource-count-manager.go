/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"fmt"

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

type ResourceCountManager struct {
	solutions.SolutionsManager
	targets.TargetsManager
	instances.InstancesManager
	solutioncontainers.SolutionContainersManager
	StateProvider states.IStateProvider
}

func (s *ResourceCountManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	if err := s.SolutionsManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.TargetsManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.InstancesManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.SolutionContainersManager.Init(ctx, config, providers); err != nil {
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

func (s *ResourceCountManager) Enabled() bool {
	return true
}
func (s *ResourceCountManager) Reconcil() []error {
	return nil
}

func (s *ResourceCountManager) Shutdown(ctx context.Context) error {
	// Explicitly call the Shutdown method of the desired embedded struct(s)
	if err := s.SolutionsManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.TargetsManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.InstancesManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.SolutionContainersManager.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
