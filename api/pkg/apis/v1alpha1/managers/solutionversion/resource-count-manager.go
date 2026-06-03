/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutionversion

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/instances"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/solutionversions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/targets"
	vendorCtx "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ResourceCountManager struct {
	solutionversions.SolutionVersionsManager
	targets.TargetsManager
	instances.InstancesManager
	solutions.SolutionsManager
}

func (s *ResourceCountManager) Init(ctx *vendorCtx.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	if err := s.SolutionVersionsManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.TargetsManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.InstancesManager.Init(ctx, config, providers); err != nil {
		return err
	}
	if err := s.SolutionsManager.Init(ctx, config, providers); err != nil {
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

	// Call SolutionVersionsManager.ListState (if available)
	solutionversionStates, err := s.SolutionVersionsManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list solutionversions: %v", err)
		return []error{err}
	}
	resourceCounts["solutionversions"] = len(solutionversionStates)

	// Call SolutionsManager.ListState (if available)
	solutionversionContainerStates, err := s.SolutionsManager.ListState(ctx, "")
	if err != nil {
		log.ErrorCtx(ctx, "Failed to list solutionversion containers: %v", err)
		return []error{err}
	}
	resourceCounts["solutions"] = len(solutionversionContainerStates)

	// Log the counts as a single string
	log.InfofCtx(ctx, fmt.Sprintf(
		"M (Inventory Count): Summary: Found %d instances, %d targets, %d solutionversions, and %d solutionversion containers across all namespaces",
		resourceCounts["instances"], resourceCounts["targets"], resourceCounts["solutionversions"], resourceCounts["solutions"],
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
	if err := s.SolutionVersionsManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.TargetsManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.InstancesManager.Shutdown(ctx); err != nil {
		return err
	}
	if err := s.SolutionsManager.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
