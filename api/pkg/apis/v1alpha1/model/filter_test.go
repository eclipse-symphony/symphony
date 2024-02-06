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

func TestFilterMatch(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	filter2 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := filter1.DeepEquals(filter2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestFilterDirectionNotMatch(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	filter2 := FilterSpec{
		Direction: "direction2",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := filter1.DeepEquals(filter2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestFilterTypeNotMatch(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	filter2 := FilterSpec{
		Direction: "direction1",
		Type:      "typ2",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := filter1.DeepEquals(filter2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestFilterTypeExtraParameter(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo": "bar",
		},
	}
	filter2 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo":     "bar",
			"another": "value",
		},
	}
	equal, err := filter1.DeepEquals(filter2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestFilterTypeMultiParameters(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo":     "bar",
			"another": "value",
			"third":   "value3",
		},
	}
	filter2 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo":     "bar",
			"another": "value",
			"third":   "value3",
		},
	}
	equal, err := filter1.DeepEquals(filter2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestFilterEqualOneEmpty(t *testing.T) {
	filter1 := FilterSpec{
		Direction: "direction1",
		Type:      "typ1",
		Parameters: map[string]string{
			"foo":     "bar",
			"another": "value",
			"third":   "value3",
		},
	}
	res, err := filter1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a FilterSpec type")
	assert.False(t, res)
}
