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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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

	// Initialize Kubernetes dynamic client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.ErrorCtx(ctx, "Failed to create Kubernetes config: %v", err)
		return []error{err}
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.ErrorCtx(ctx, "Failed to create dynamic client: %v", err)
		return []error{err}
	}

	// Get resource counts
	resourceCounts, err := s.getResourceCounts(ctx, dynamicClient)
	if err != nil {
		log.ErrorCtx(ctx, "Failed to get resource counts: %v", err)
		return []error{err}
	}

	// Log the counts as a single string
	log.InfofCtx(ctx, fmt.Sprintf(
		"Summary: Found %d instances, %d targets, %d solutions, and %d solution containers across all namespaces",
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
func (s *SummaryCleanupManager) getResourceCounts(ctx context.Context, client dynamic.Interface) (map[string]int, error) {
	// Define the GVRs (GroupVersionResource) for the resources
	resourceGVRs := map[string]schema.GroupVersionResource{
		"instances":          {Group: "solution.symphony", Version: "v1", Resource: "instances"},
		"targets":            {Group: "fabric.symphony", Version: "v1", Resource: "targets"},
		"solutions":          {Group: "solution.symphony", Version: "v1", Resource: "solutions"},
		"solutioncontainers": {Group: "solution.symphony", Version: "v1", Resource: "solutioncontainers"},
	}

	// Get all namespaces
	namespaces, err := client.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	log.InfofCtx(ctx, fmt.Sprintf("Found %d namespaces", len(namespaces.Items)))

	// Initialize counters
	resourceCounts := make(map[string]int)

	// Iterate through namespaces and count resources
	for _, ns := range namespaces.Items {
		namespace := ns.GetName()
		log.InfofCtx(ctx, fmt.Sprintf("Checking namespace: %s", namespace))

		for resourceName, gvr := range resourceGVRs {
			resourceCounts[resourceName] += countResources(ctx, client, gvr, namespace, resourceName)
		}
	}

	return resourceCounts, nil
}

// countResources is a helper function to count resources in a namespace
func countResources(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, resourceName string) int {
	resources, err := client.Resource(gvr).Namespace(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		log.InfofCtx(ctx, fmt.Sprintf("Failed to list %s in namespace %s: %v", resourceName, namespace, err))
		return 0
	}
	log.InfofCtx(ctx, fmt.Sprintf("Namespace: %s, Found %d %s", namespace, len(resources.Items), resourceName))
	return len(resources.Items)
}

func (s *SummaryCleanupManager) Reconcil() []error {
	return nil
}
