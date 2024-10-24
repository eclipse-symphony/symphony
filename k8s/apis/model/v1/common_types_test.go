/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"encoding/json"
	"testing"
)

func TestStageSpecSchedule(t *testing.T) {
	jsonString := `{"schedule": "2021-01-30T08:30:10+08:00"}`

	var newStage StageSpec
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "2021-01-30T08:30:10+08:00"
	if newStage.Schedule != targetTime {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
}
