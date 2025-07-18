/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/stretchr/testify/assert"
)

func TestObjectMetaWithString(t *testing.T) {
	jsonString := `{"etag": "33"}`

	var newStage ObjectMeta
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "33"
	if newStage.ETag != targetTime {
		t.Fatalf("Generation is not match: %v", err)
	}
}

func TestObjectMetaWithNumber(t *testing.T) {
	jsonString := `{"etag": 33}`

	var newStage ObjectMeta
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "33"
	if newStage.ETag != targetTime {
		t.Fatalf("Generation is not match: %v", err)
	}
}

func TestPreserveSystemMetadata_EmptyCurrentMetadata(t *testing.T) {
	// Test case: Current metadata has nil annotations and labels
	current := ObjectMeta{
		Name:      "test-object",
		Namespace: "default",
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "correlation-123",
			constants.GuidKey:               "guid-456",
			constants.SummaryJobIdKey:       "job-789",
		},
		Labels: map[string]string{
			constants.DisplayName: "Test Display",
			constants.Solution:    "test-solution",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify annotations were preserved
	assert.NotNil(t, current.Annotations)
	assert.Equal(t, "correlation-123", current.Annotations[constants.AzureCorrelationIdKey])
	assert.Equal(t, "guid-456", current.Annotations[constants.GuidKey])
	assert.Equal(t, "job-789", current.Annotations[constants.SummaryJobIdKey])

	// Verify labels were preserved
	assert.NotNil(t, current.Labels)
	assert.Equal(t, "Test Display", current.Labels[constants.DisplayName])
	assert.Equal(t, "test-solution", current.Labels[constants.Solution])
}

func TestPreserveSystemMetadata_ExistingAnnotationsNotOverwritten(t *testing.T) {
	// Test case: Current metadata has existing annotations that should not be overwritten
	current := ObjectMeta{
		Name: "test-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "existing-correlation",
			"custom-annotation":             "custom-value",
		},
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "new-correlation",
			constants.GuidKey:               "new-guid",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify existing annotation was not overwritten
	assert.Equal(t, "existing-correlation", current.Annotations[constants.AzureCorrelationIdKey])
	// Verify new annotation was added
	assert.Equal(t, "new-guid", current.Annotations[constants.GuidKey])
	// Verify custom annotation was preserved
	assert.Equal(t, "custom-value", current.Annotations["custom-annotation"])
}

func TestPreserveSystemMetadata_ExistingLabelsNotOverwritten(t *testing.T) {
	// Test case: Current metadata has existing labels that should not be overwritten
	current := ObjectMeta{
		Name: "test-object",
		Labels: map[string]string{
			constants.DisplayName: "Existing Display",
			"custom-label":        "custom-value",
		},
	}

	source := ObjectMeta{
		Name: "source-object",
		Labels: map[string]string{
			constants.DisplayName: "New Display",
			constants.Solution:    "new-solution",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify existing label was not overwritten
	assert.Equal(t, "Existing Display", current.Labels[constants.DisplayName])
	// Verify new label was added
	assert.Equal(t, "new-solution", current.Labels[constants.Solution])
	// Verify custom label was preserved
	assert.Equal(t, "custom-value", current.Labels["custom-label"])
}

func TestPreserveSystemMetadata_AnnotationsByPostfixes(t *testing.T) {
	// Test case: Preserve annotations that match system reserved postfixes
	current := ObjectMeta{
		Name: "test-object",
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			"instance.symphony/started-at": "2023-01-01T00:00:00Z",
			"target.symphony/started-at":   "2023-01-02T00:00:00Z",
			"custom-annotation":            "should-not-be-preserved",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify postfix annotations were preserved
	assert.Equal(t, "2023-01-01T00:00:00Z", current.Annotations["instance.symphony/started-at"])
	assert.Equal(t, "2023-01-02T00:00:00Z", current.Annotations["target.symphony/started-at"])
	// Verify custom annotation was not preserved
	assert.Empty(t, current.Annotations["custom-annotation"])
}

func TestPreserveSystemMetadata_MixedScenario(t *testing.T) {
	// Test case: Complex scenario with mixed existing and new metadata
	current := ObjectMeta{
		Name: "test-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "existing-correlation",
			"custom-annotation":             "custom-value",
		},
		Labels: map[string]string{
			constants.DisplayName: "Existing Display",
			"custom-label":        "custom-value",
		},
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "new-correlation",
			constants.GuidKey:               "new-guid",
			"instance.symphony/started-at":  "2023-01-01T00:00:00Z",
			"custom-annotation":             "should-not-overwrite",
		},
		Labels: map[string]string{
			constants.DisplayName: "New Display",
			constants.Solution:    "new-solution",
			"custom-label":        "should-not-overwrite",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify annotations
	assert.Equal(t, "existing-correlation", current.Annotations[constants.AzureCorrelationIdKey]) // Not overwritten
	assert.Equal(t, "new-guid", current.Annotations[constants.GuidKey])                           // Added
	assert.Equal(t, "2023-01-01T00:00:00Z", current.Annotations["instance.symphony/started-at"])  // Added
	assert.Equal(t, "custom-value", current.Annotations["custom-annotation"])                     // Preserved

	// Verify labels
	assert.Equal(t, "Existing Display", current.Labels[constants.DisplayName]) // Not overwritten
	assert.Equal(t, "new-solution", current.Labels[constants.Solution])        // Added
	assert.Equal(t, "custom-value", current.Labels["custom-label"])            // Preserved
}

func TestPreserveSystemMetadata_EmptySourceMetadata(t *testing.T) {
	// Test case: Source metadata is empty
	current := ObjectMeta{
		Name: "test-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "existing-correlation",
		},
		Labels: map[string]string{
			constants.DisplayName: "Existing Display",
		},
	}

	source := ObjectMeta{
		Name: "source-object",
	}

	current.PreserveSystemMetadata(source)

	// Verify current metadata remains unchanged
	assert.Equal(t, "existing-correlation", current.Annotations[constants.AzureCorrelationIdKey])
	assert.Equal(t, "Existing Display", current.Labels[constants.DisplayName])
}

func TestPreserveSystemMetadata_NilSourceAnnotations(t *testing.T) {
	// Test case: Source has nil annotations
	current := ObjectMeta{
		Name: "test-object",
	}

	source := ObjectMeta{
		Name: "source-object",
		Labels: map[string]string{
			constants.DisplayName: "Test Display",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify annotations map was created but empty
	assert.NotNil(t, current.Annotations)
	assert.Empty(t, current.Annotations)

	// Verify labels were preserved
	assert.NotNil(t, current.Labels)
	assert.Equal(t, "Test Display", current.Labels[constants.DisplayName])
}

func TestPreserveSystemMetadata_NilSourceLabels(t *testing.T) {
	// Test case: Source has nil labels
	current := ObjectMeta{
		Name: "test-object",
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey: "correlation-123",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify annotations were preserved
	assert.NotNil(t, current.Annotations)
	assert.Equal(t, "correlation-123", current.Annotations[constants.AzureCorrelationIdKey])

	// Verify labels map was created but empty
	assert.NotNil(t, current.Labels)
	assert.Empty(t, current.Labels)
}

func TestPreserveSystemMetadata_AllSystemReservedAnnotations(t *testing.T) {
	// Test case: Verify all system reserved annotations are handled
	current := ObjectMeta{
		Name: "test-object",
	}

	source := ObjectMeta{
		Name: "source-object",
		Annotations: map[string]string{
			constants.AzureCorrelationIdKey:        "correlation-123",
			constants.AzureCloudLocationKey:        "eastus",
			constants.AzureEdgeLocationKey:         "edge-location",
			constants.AzureOperationIdKey:          "operation-456",
			constants.AzureDeleteOperationKey:      "delete-op",
			constants.AzureNameIdKey:               "azure-name",
			constants.AzureResourceIdKey:           "/subscriptions/123/resourceGroups/rg",
			constants.AzureSystemDataKey:           "system-data",
			constants.AzureTenantIdKey:             "tenant-123",
			constants.RunningAzureCorrelationIdKey: "running-correlation",
			constants.SummaryJobIdKey:              "summary-job",
			constants.GuidKey:                      "guid-789",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify all system reserved annotations were preserved
	assert.Equal(t, "correlation-123", current.Annotations[constants.AzureCorrelationIdKey])
	assert.Equal(t, "eastus", current.Annotations[constants.AzureCloudLocationKey])
	assert.Equal(t, "edge-location", current.Annotations[constants.AzureEdgeLocationKey])
	assert.Equal(t, "operation-456", current.Annotations[constants.AzureOperationIdKey])
	assert.Equal(t, "delete-op", current.Annotations[constants.AzureDeleteOperationKey])
	assert.Equal(t, "azure-name", current.Annotations[constants.AzureNameIdKey])
	assert.Equal(t, "/subscriptions/123/resourceGroups/rg", current.Annotations[constants.AzureResourceIdKey])
	assert.Equal(t, "system-data", current.Annotations[constants.AzureSystemDataKey])
	assert.Equal(t, "tenant-123", current.Annotations[constants.AzureTenantIdKey])
	assert.Equal(t, "running-correlation", current.Annotations[constants.RunningAzureCorrelationIdKey])
	assert.Equal(t, "summary-job", current.Annotations[constants.SummaryJobIdKey])
	assert.Equal(t, "guid-789", current.Annotations[constants.GuidKey])
}

func TestPreserveSystemMetadata_AllSystemReservedLabels(t *testing.T) {
	// Test case: Verify all system reserved labels are handled
	current := ObjectMeta{
		Name: "test-object",
	}

	source := ObjectMeta{
		Name: "source-object",
		Labels: map[string]string{
			constants.Campaign:       "test-campaign",
			constants.DisplayName:    "Test Display",
			constants.ProviderName:   "test-provider",
			constants.ManagerMetaKey: "test-manager",
			constants.ParentName:     "parent-resource",
			constants.RootResource:   "root-resource",
			constants.Solution:       "test-solution",
			constants.StagedTarget:   "staged-target",
			constants.StatusMessage:  "status-message",
			constants.Target:         "test-target",
		},
	}

	current.PreserveSystemMetadata(source)

	// Verify all system reserved labels were preserved
	assert.Equal(t, "test-campaign", current.Labels[constants.Campaign])
	assert.Equal(t, "Test Display", current.Labels[constants.DisplayName])
	assert.Equal(t, "test-provider", current.Labels[constants.ProviderName])
	assert.Equal(t, "test-manager", current.Labels[constants.ManagerMetaKey])
	assert.Equal(t, "parent-resource", current.Labels[constants.ParentName])
	assert.Equal(t, "root-resource", current.Labels[constants.RootResource])
	assert.Equal(t, "test-solution", current.Labels[constants.Solution])
	assert.Equal(t, "staged-target", current.Labels[constants.StagedTarget])
	assert.Equal(t, "status-message", current.Labels[constants.StatusMessage])
	assert.Equal(t, "test-target", current.Labels[constants.Target])
}
