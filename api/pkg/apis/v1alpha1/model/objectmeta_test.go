/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"encoding/json"
	"testing"
)

func TestObjectMetaWithString(t *testing.T) {
	jsonString := `{"generation": "33"}`

	var newStage ObjectMeta
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "33"
	if newStage.Generation != targetTime {
		t.Fatalf("Generation is not match: %v", err)
	}
}

func TestObjectMetaWithNumber(t *testing.T) {
	jsonString := `{"generation": 33}`

	var newStage ObjectMeta
	var err = json.Unmarshal([]byte(jsonString), &newStage)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	targetTime := "33"
	if newStage.Generation != targetTime {
		t.Fatalf("Generation is not match: %v", err)
	}
}
