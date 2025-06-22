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

func TestGenerateErrorMessage(t *testing.T) {
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
					"target1": {Status: v1alpha2.OK.String(), Message: "Target 1 failed"},
				},
			},
			expected: "",
		},
		{
			name: "Target error without component errors",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "Test message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target failed",
					},
					"target2": {
						Status:  v1alpha2.OK.String(),
						Message: "Success",
					},
				},
			},
			expected: `Failed to deploy instance: Test message. Target errors: target1: "Target failed"`,
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
								Status:  v1alpha2.InternalError, // Failed state
								Message: "Component error",
							},
						},
					},
					"target2": {
						Status:  v1alpha2.OK.String(),
						Message: "Success",
					},
				},
			},
			expected: `Failed to deploy instance: Test message. Target errors: target1: "Target failed" (target1.comp1: Component error)`,
		},
		{
			name: "Multiple targets with errors and component errors",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "Test message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target 1 failed",
					},
					"target2": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target 2 failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.InternalError, // Failed state
								Message: "Component 1 error",
							},
							"comp2": {
								Status:  v1alpha2.InternalError, // Failed state
								Message: "Component 2 error",
							},
						},
					},
					"target3": {
						Status:  v1alpha2.OK.String(),
						Message: "Success",
					},
				},
			},
			expected: `Failed to deploy instance: Test message. Target errors: target1: "Target 1 failed", target2: "Target 2 failed" (target2.comp1: Component 1 error) (target2.comp2: Component 2 error)`,
		},
		{
			name: "Multiple targets all with errors and mixed component states",
			summary: SummarySpec{
				AllAssignedDeployed: false,
				SummaryMessage:      "Test message",
				TargetResults: map[string]TargetResultSpec{
					"target1": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target 1 failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.OK, // OK state
								Message: "OK",
							},
							"comp2": {
								Status:  v1alpha2.InternalError, // Failed state
								Message: "Component error",
							},
						},
					},
					"target2": {
						Status:  v1alpha2.InternalError.String(),
						Message: "Target 2 failed",
						ComponentResults: map[string]ComponentResultSpec{
							"comp1": {
								Status:  v1alpha2.InternalError, // Failed state
								Message: "Another error",
							},
						},
					},
				},
			},
			expected: `Failed to deploy instance: Test message. Target errors: target1: "Target 1 failed" (target1.comp2: Component error), target2: "Target 2 failed" (target2.comp1: Another error)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.summary.GenerateErrorMessage()
			if result != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}
