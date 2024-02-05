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
	assert.EqualError(t, equal, "required property 'requiredProperties1' is missing")
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
	assert.EqualError(t, equal, "required metadata 'RequiredMetadata1' is missing")
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
	assert.EqualError(t, equal, "provider requires component type 'requiredComponentType', but 'requiredComponentType1' is found instead")
}

func TestValidateInputs(t *testing.T) {
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
	inputs := map[string]interface{}{
		"requiredProperties1": "requiredProperties1",
	}
	equal := validationRule.ValidateInputs(inputs)
	assert.Nil(t, equal)

	inputs2 := map[string]interface{}{
		"requiredProperties": "requiredProperties",
	}
	equal = validationRule.ValidateInputs(inputs2)
	assert.EqualError(t, equal, "required property 'requiredProperties1' is missing")
}

func TestIsComponentChangedNoWildcard(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name: "a",
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

func TestIsComponentChanged_ChangeComponentNameIgnoreCase(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:            "a",
				IsComponentName: true,
				IgnoreCase:      true,
			},
		},
	}
	old := ComponentSpec{
		Name: "a",
	}
	new := ComponentSpec{
		Name: "A",
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.False(t, changed)
}

func TestIsComponentChanged_ChangeComponentNameNoIgnoreCase(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:            "a",
				IsComponentName: true,
			},
		},
	}
	old := ComponentSpec{
		Name: "a",
	}
	new := ComponentSpec{
		Name: "A",
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

func TestIsComponentChangedNoWildcard_Metadata(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionMetadata: []PropertyDesc{
			{
				Name: "a",
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "b",
		},
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "c",
		},
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

func TestIsChangedWildcard_Metadata(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionMetadata: []PropertyDesc{
			{
				Name: "*",
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "b",
		},
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "c",
		},
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

func TestComponentIsChanged_SkipMissingProperty(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: true,
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
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.False(t, changed)
}

func TestComponentIsChanged_SkipMissingMetadata(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionMetadata: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: true,
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "b",
		},
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.False(t, changed)
}

// FIX: validationrule.go#L132&134
func TestComponentIsChanged_MissingPropertyNotInOld(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: false,
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
		Properties: map[string]interface{}{
			"a": "b",
		},
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

// FIX: validationrule.go#L149&151
func TestComponentIsChanged_MissingMetadataNotInOld(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionMetadata: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: false,
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "b",
		},
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

// FIX: validationrule.go#L128&130
func TestComponentIsChanged_MissingProperty(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: false,
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
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

// FIX: validationrule.go#L145&147
func TestComponentIsChanged_MissingMetadata(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionMetadata: []PropertyDesc{
			{
				Name:          "a",
				SkipIfMissing: false,
			},
		},
	}
	old := ComponentSpec{
		Type: "requiredComponentType",
		Metadata: map[string]string{
			"a": "b",
		},
	}
	new := ComponentSpec{
		Type: "requiredComponentType",
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.True(t, changed)
}

// FIX: validationrule.go#L77
func TestComponentIsChanged_ComponentNameHasPrefix(t *testing.T) {
	// Create a new instance of our test struct
	validationRule := ValidationRule{
		ChangeDetectionProperties: []PropertyDesc{
			{
				Name:            "a",
				PrefixMatch:     true,
				IsComponentName: true,
			},
		},
	}
	old := ComponentSpec{
		Name: "a",
	}
	new := ComponentSpec{
		Name: "a1",
	}

	changed := validationRule.IsComponentChanged(old, new)
	assert.False(t, changed)
}
