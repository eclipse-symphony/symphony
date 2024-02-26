/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

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
	assert.Equal(t, 0, s.SuccessCount) //ver 0.48.1: UpdateTargetResult no longer updates success count
}
