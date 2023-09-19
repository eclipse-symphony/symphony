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

func TestBindingMatch(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	binding2 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestBindingMatchOneEmpty(t *testing.T) {
	binding1 := BindingSpec{
		Role: "role",
	}
	res, err := binding1.DeepEquals(nil)
	assert.Errorf(t, err, "parameter is not a BindingSpec type")
	assert.False(t, res)
}

func TestBindingRoleNotMatch(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	binding2 := BindingSpec{
		Role:     "role1",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestBindingProviderNotMatch(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	binding2 := BindingSpec{
		Role:     "role",
		Provider: "provider1",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestBindingConfigNotMatch(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	binding2 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar1",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestBindingExtraConfigProperties(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	binding2 := BindingSpec{
		Role:     "role1",
		Provider: "provider",
		Config: map[string]string{
			"foo":  "bar",
			"foo2": "bar2",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestBindingMissingFilter(t *testing.T) {
	binding1 := BindingSpec{
		Role:     "role",
		Provider: "provider",
		Config: map[string]string{
			"foo":  "bar",
			"foo2": "bar2",
		},
	}
	binding2 := BindingSpec{
		Role:     "role1",
		Provider: "provider",
		Config: map[string]string{
			"foo": "bar",
		},
	}
	equal, err := binding1.DeepEquals(binding2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
