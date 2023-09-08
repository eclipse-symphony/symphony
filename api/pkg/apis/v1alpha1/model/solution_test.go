/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
	assert.Errorf(t, err, "parameter is not a SolutionSpec type")
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
