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
