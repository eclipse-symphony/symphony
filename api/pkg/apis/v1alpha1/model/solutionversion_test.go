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

func TestSolutionVersionDeepEquals(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestSolutionVersionDeepEqualsOneEmpty(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a SolutionVersionState type")
	assert.False(t, res)
}

func TestSolutionVersionDeepEqualsDisplayNameNotMatch(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName1",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionVersionDeepEqualsNamespaceNotMatch(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default1",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionVersionDeepEqualsMetadataKeyNotMatch(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo1": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionVersionDeepEqualsMetadataValueNotMatch(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components:  []ComponentSpec{{}},
			Metadata:    map[string]string{"foo": "bar1"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionVersionDeepEqualsComponentNameNotMatch(t *testing.T) {
	solutionversion := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Metadata: map[string]string{"foo": "bar"},
		},
	}
	other := SolutionVersionState{
		ObjectMeta: ObjectMeta{
			Namespace: "Default",
		},
		Spec: &SolutionVersionSpec{
			DisplayName: "SolutionVersionName",
			Components: []ComponentSpec{{
				Name: "ComponentName1",
			}},
			Metadata: map[string]string{"foo": "bar"},
		},
	}
	res, err := solutionversion.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}
