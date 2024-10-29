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
