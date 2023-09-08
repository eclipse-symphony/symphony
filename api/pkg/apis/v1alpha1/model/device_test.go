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
	assert.Errorf(t, err, "parameter is not a DeviceSpec type")
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
