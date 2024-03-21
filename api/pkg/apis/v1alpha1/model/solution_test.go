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

func TestSolutionDeepEquals(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestSolutionDeepEqualsOneEmpty(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solution.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a SolutionState type")
	assert.False(t, res)
}

func TestSolutionDeepEqualsDisplayNameNotMatch(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName1",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsNamespaceNotMatch(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default1",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsMetadataKeyNotMatch(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo1": "bar"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsMetadataValueNotMatch(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar1"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsComponentNameNotMatch(t *testing.T) {
	solution := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Metadata: map[string]string{"foo": "bar"},
		},
	}
	other := SolutionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionSpec{
			DisplayName: "SolutionName",
			Scope:       "Default",
			Components: []ComponentSpec{{
				Name: "ComponentName1",
			}},
			Metadata: map[string]string{"foo": "bar"},
		},
	}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}
