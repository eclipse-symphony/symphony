/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateString(t *testing.T) {
	stateStringMap := map[State]string{
		OK:                 "OK",
		Accepted:           "Accepted",
		BadRequest:         "Bad Request",
		Unauthorized:       "Unauthorized",
		NotFound:           "Not Found",
		MethodNotAllowed:   "Method Not Allowed",
		Conflict:           "Conflict",
		InternalError:      "Internal Error",
		BadConfig:          "Bad Config",
		MissingConfig:      "Missing Config",
		InvalidArgument:    "Invalid Argument",
		APIRedirect:        "API Redirect",
		FileAccessError:    "File Access Error",
		SerializationError: "Serialization Error",
		DeleteRequested:    "Delete Requested",
		UpdateFailed:       "Update Failed",
		DeleteFailed:       "Delete Failed",
		ValidateFailed:     "Validate Failed",
		Updated:            "Updated",
		Deleted:            "Deleted",
		Running:            "Running",
		Paused:             "Paused",
		Done:               "Done",
		Delayed:            "Delayed",
		Untouched:          "Untouched",
		NotImplemented:     "Not Implemented",
	}
	for state, expected := range stateStringMap {
		assert.Equal(t, expected, state.String())
	}

	assert.Equal(t, "Unknown State: 10000", State(10000).String())
}
