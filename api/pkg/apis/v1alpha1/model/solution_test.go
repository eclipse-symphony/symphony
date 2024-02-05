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
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestSolutionDeepEqualsOneEmpty(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	res, err := solution.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a SolutionSpec type")
	assert.False(t, res)
}

func TestSolutionDeepEqualsDisplayNameNotMatch(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName1",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsScopeNotMatch(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default1",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsMetadataKeyNotMatch(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo1": "bar"},
		Components:  []ComponentSpec{{}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsMetadataValueNotMatch(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar1"},
		Components:  []ComponentSpec{{}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestSolutionDeepEqualsComponentNameNotMatch(t *testing.T) {
	solution := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
	}
	other := SolutionSpec{
		DisplayName: "SolutionName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName1",
		}}}
	res, err := solution.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}
