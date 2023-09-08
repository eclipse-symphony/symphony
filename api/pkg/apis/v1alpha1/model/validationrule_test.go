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

func TestValidate(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		RequiredComponentType: "requiredComponentType",
		RequiredProperties: []string{
			"requiredProperties1",
		},
		RequiredMetadata: []string{
			"RequiredMetadata1",
		},
	}
	components := []ComponentSpec{
		{
			Type: "requiredComponentType",
			Properties: map[string]interface{}{
				"requiredProperties1": "requiredProperties1",
			},
			Metadata: map[string]string{
				"RequiredMetadata1": "RequiredMetadata1",
			},
		},
	}
	equal := validationRule.Validate(components)
	assert.Nil(t, equal)
}

func TestValidateNil(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		RequiredComponentType: "requiredComponentType",
		RequiredProperties: []string{
			"requiredProperties1",
			"requiredProperties2",
		},
		OptionalProperties: []string{
			"optionalProperties1",
			"optionalProperties2",
		},
		RequiredMetadata: []string{
			"RequiredMetadata1",
			"RequiredMetadata2",
		},
		OptionalMetadata: []string{
			"OptionalMetadata1",
			"OptionalMetadata2",
		},
	}

	equal := validationRule.Validate(nil)
	assert.Nil(t, equal)
}

func TestValidateCOA(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		RequiredComponentType: "requiredComponentType",
		RequiredProperties: []string{
			"requiredProperties1",
			"requiredProperties2",
		},
		OptionalProperties: []string{
			"optionalProperties1",
			"optionalProperties2",
		},
		RequiredMetadata: []string{
			"RequiredMetadata1",
			"RequiredMetadata2",
		},
		OptionalMetadata: []string{
			"OptionalMetadata1",
			"OptionalMetadata2",
		},
	}

	components := []ComponentSpec{
		{
			Type: "requiredComponentType",
			Metadata: map[string]string{
				"RequiredMetadata1": "RequiredMetadata1",
				"RequiredMetadata2": "RequiredMetadata2",
			},
		},
	}
	equal := validationRule.Validate(components)
	assert.Errorf(t, equal, "required property 'requiredProperties1' is missing")
}

func TestValidateMetadata(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		RequiredComponentType: "requiredComponentType",
		RequiredMetadata: []string{
			"RequiredMetadata1",
			"RequiredMetadata2",
		},
	}

	components := []ComponentSpec{
		{
			Type: "requiredComponentType",
			Properties: map[string]interface{}{
				"requiredProperties1": "requiredProperties1",
				"requiredProperties2": "requiredProperties2",
			},
		},
	}
	equal := validationRule.Validate(components)
	assert.Errorf(t, equal, "required property 'RequiredMetadata1' is missing")
}

func TestValidateComponentType(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		RequiredComponentType: "requiredComponentType",
	}

	components := []ComponentSpec{
		{
			Type: "requiredComponentType1",
		},
	}
	equal := validationRule.Validate(components)
	assert.Errorf(t, equal, "required property 'requiredComponentType' is missing")
}
