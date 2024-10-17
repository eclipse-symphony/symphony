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

func TestUpdateFailedTargetResultToSucceed(t *testing.T) {
	s := &SummarySpec{
		TargetResults: map[string]TargetResultSpec{
			"target1": {
				Status: "ERROR",
				ComponentResults: map[string]ComponentResultSpec{
					"component1": {
						Status:  v1alpha2.BadConfig,
						Message: "Component 1 is in bad config",
					},
				},
			},
		},
	}
	s.UpdateTargetResult("target1", TargetResultSpec{
		Status: "OK",
		ComponentResults: map[string]ComponentResultSpec{
			"component1": {
				Status:  v1alpha2.Accepted,
				Message: "Component 1 is accepted",
			},
		},
	})
	assert.Equal(t, "OK", s.TargetResults["target1"].Status)
	assert.Equal(t, v1alpha2.Accepted, s.TargetResults["target1"].ComponentResults["component1"].Status)
	assert.Equal(t, "Component 1 is accepted", s.TargetResults["target1"].ComponentResults["component1"].Message)
}
