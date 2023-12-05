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

func TestIsChangedWildcard(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name: "*",
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
		Properties: map[string]interface{}{
			"a": "b",
		},
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
		Properties: map[string]interface{}{
			"a": "c",
		},
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
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
