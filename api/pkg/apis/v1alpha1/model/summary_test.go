/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTargetResult(t *testing.T) {
	s := &SummarySpec{
		TargetResults: map[string]TargetResultSpec{
			"target1": {
				Status: "OK",
			},
			"target2": {
				Status: "OK",
			},
			"target3": {
				Status: "OK",
			},
		},
	}
	s.UpdateTargetResult("target2", TargetResultSpec{
		Status: "ERROR",
	})
	assert.Equal(t, "ERROR", s.TargetResults["target2"].Status)
	assert.Equal(t, 0, s.SuccessCount) //ver 0.48.1: UpdateTargetResult no longer updates success count
}

func TestUpdateTargetResultWithComponentResults(t *testing.T) {
	s := &SummarySpec{
		TargetResults: map[string]TargetResultSpec{
			"target1": {
				Status: "OK",
				ComponentResults: map[string]ComponentResultSpec{
					"component1": {
						Status:  v1alpha2.Accepted,
						Message: "Component 1 is accepted",
					},
				},
			},
		},
	}
	s.UpdateTargetResult("target1", TargetResultSpec{
		Status: "ERROR",
		ComponentResults: map[string]ComponentResultSpec{
			"component1": {
				Status:  v1alpha2.BadConfig,
				Message: "Component 1 is in bad config",
			},
		},
	})
	assert.Equal(t, "ERROR", s.TargetResults["target1"].Status)
	assert.Equal(t, v1alpha2.BadConfig, s.TargetResults["target1"].ComponentResults["component1"].Status)
	assert.Equal(t, "Component 1 is in bad config", s.TargetResults["target1"].ComponentResults["component1"].Message)
	assert.Equal(t, 0, s.SuccessCount) //ver 0.48.1: UpdateTargetResult no longer updates success count
}

func TestGenerateStatusMessage(t *testing.T) {
	tests := []struct {
		name     string
		summary  SummarySpec
		expected string
	}{
		{
			name: "AllAssignedDeployed true should return empty string",
			summary: SummarySpec{
				AllAssignedDeployed: true,
				SummaryMessage:      "Some message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {Status: v1alpha2.OK.String(), Message: "Target 1 status"},
				},
			},
			expected: "",
		},
		{
			name: "Empty summary message",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target failed",
					},
				},
			},
			expected: `Failed to deploy. Detailed status: target1: "Target failed"`,
		},
		{
			name: "Target list contains both OK and Failed status",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "Test message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.OK.String(),
						Message: "Deployment successful",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.OK,
								Message: "Component OK",
							},
						},
					},
					"target2": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Deployment failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.InternalError,
								Message: "Component failed",
							},
						},
					},
				},
			},
			expected: `Failed to deploy: Test message. Detailed status: target1: "Deployment successful" (target1.comp1: Component OK), target2: "Deployment failed" (target2.comp1: Component failed)`,
		},
		{
			name: "Target with component errors",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "Test message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.InternalError,
								Message: "Component error",
							},
						},
					},
				},
			},
			expected: `Failed to deploy: Test message. Detailed status: target1: "Target failed" (target1.comp1: Component error)`,
		},
		{
			name: "Empty summary message with multiple targets and components",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.OK.String(),
						Message: "Success",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.OK,
								Message: "OK",
							},
						},
					},
					"target2": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.InternalError,
								Message: "Component error",
							},
						},
					},
				},
			},
			expected: `Failed to deploy. Detailed status: target1: "Success" (target1.comp1: OK), target2: "Failed" (target2.comp1: Component error)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.summary.GenerateStatusMessage()
			if result != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}
