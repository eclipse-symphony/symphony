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

func TestDeviceMatch(t *testing.T) {
	device1 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	device2 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	equal, err := device1.DeepEquals(device2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestDeviceMatchOneEmpty(t *testing.T) {
	device1 := DeviceSpec{
		DisplayName: "displayName",
	}
	res, err := device1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a DeviceSpec type")
	assert.False(t, res)
}

func TestDeviceDisplayNameNotMatch(t *testing.T) {
	device1 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	device2 := DeviceSpec{
		DisplayName: "displayName1",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	equal, err := device1.DeepEquals(device2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestDevicePropertiesNotMatch(t *testing.T) {
	device1 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	device2 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo1": "bar1",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	equal, err := device1.DeepEquals(device2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestDeviceBindingsNotMatch(t *testing.T) {
	device1 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role",
		}},
	}
	device2 := DeviceSpec{
		DisplayName: "displayName",
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{{
			Role: "role1",
		}},
	}
	equal, err := device1.DeepEquals(device2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
