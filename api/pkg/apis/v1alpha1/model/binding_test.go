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
	assert.EqualError(t, err, "parameter is not a BindingSpec type")
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
