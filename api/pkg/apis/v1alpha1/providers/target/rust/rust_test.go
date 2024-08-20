/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestMockRustProviderGetValidationRule(t *testing.T) {
	config := RustTargetProviderConfig{
		Name:    "mock",
		LibFile: "./target/release/libmock.so",
		LibHash: "26e68667de3d7bfd5ff758191ce4a231800a93f52e1751c10fa0afc7811893cd",
	}
	rustProvider := &RustTargetProvider{}
	err := rustProvider.Init(config)
	assert.Nil(t, err)

	// Example usage of GetValidationRule
	rule := rustProvider.GetValidationRule(context.Background())

	assert.Equal(t, "example_type", rule.RequiredComponentType)

	// Component Validation Rule
	assert.Equal(t, "example_type", rule.ComponentValidationRule.RequiredComponentType)
	assert.Equal(t, 1, len(rule.ComponentValidationRule.ChangeDetectionProperties))
	assert.Equal(t, "example_property", rule.ComponentValidationRule.ChangeDetectionProperties[0].Name)
	assert.Equal(t, true, rule.ComponentValidationRule.ChangeDetectionProperties[0].IgnoreCase)
	assert.Equal(t, true, rule.ComponentValidationRule.ChangeDetectionProperties[0].SkipIfMissing)
	assert.Equal(t, true, rule.ComponentValidationRule.ChangeDetectionProperties[0].PrefixMatch)
	assert.Equal(t, true, rule.ComponentValidationRule.ChangeDetectionProperties[0].IsComponentName)

	assert.Equal(t, 1, len(rule.ComponentValidationRule.ChangeDetectionMetadata))
	assert.Equal(t, "example_metadata_property", rule.ComponentValidationRule.ChangeDetectionMetadata[0].Name)

	assert.Equal(t, 1, len(rule.ComponentValidationRule.RequiredProperties))
	assert.Equal(t, "required_property_name", rule.ComponentValidationRule.RequiredProperties[0])

	assert.Equal(t, 1, len(rule.ComponentValidationRule.OptionalProperties))
	assert.Equal(t, "optional_property_name", rule.ComponentValidationRule.OptionalProperties[0])

	assert.Equal(t, 1, len(rule.ComponentValidationRule.RequiredMetadata))
	assert.Equal(t, "required_metadata_name", rule.ComponentValidationRule.RequiredMetadata[0])

	assert.Equal(t, 1, len(rule.ComponentValidationRule.OptionalMetadata))
	assert.Equal(t, "optional_metadata_name", rule.ComponentValidationRule.OptionalMetadata[0])

	// Sidecar Validation Rule
	assert.Equal(t, "sidecar_example_type", rule.SidecarValidationRule.RequiredComponentType)
	assert.Equal(t, 1, len(rule.SidecarValidationRule.ChangeDetectionProperties))
	assert.Equal(t, "sidecar_example_property", rule.SidecarValidationRule.ChangeDetectionProperties[0].Name)

	assert.Equal(t, true, rule.AllowSidecar)
	assert.Equal(t, true, rule.ScopeIsolation)
	assert.Equal(t, true, rule.InstanceIsolation)
}

func TestMockRustProviderGet(t *testing.T) {
	config := RustTargetProviderConfig{
		Name:    "mock",
		LibFile: "./target/release/libmock.so",
		LibHash: "26e68667de3d7bfd5ff758191ce4a231800a93f52e1751c10fa0afc7811893cd",
	}
	rustProvider := &RustTargetProvider{}
	err := rustProvider.Init(config)
	assert.Nil(t, err)

	// Create a mock TargetState to populate the Targets map
	targetState := model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      "example_target",
			Namespace: "default",
		},
		Status: model.TargetStatus{},
		Spec: &model.TargetSpec{
			DisplayName:   "Example Target",
			Scope:         "example_scope",
			Metadata:      map[string]string{"example_target_metadata_key": "example_target_metadata_value"},
			Properties:    map[string]string{"example_target_property_key": "example_target_property_value"},
			Components:    nil,
			Constraints:   "example_constraints",
			Topologies:    nil,
			ForceRedeploy: false,
		},
	}

	// Initialize the Targets map in DeploymentSpec
	deployment := model.DeploymentSpec{
		Targets: map[string]model.TargetState{
			"example_target": targetState,
		},
	}

	// Mock references (empty for this example)
	references := []model.ComponentStep{}

	// Call the Get method
	components, err := rustProvider.Get(context.Background(), deployment, references)
	assert.Nil(t, err)

	// Validate the returned component specifications
	assert.Equal(t, 1, len(components))

	// Validate the properties of the returned ComponentSpec
	component := components[0]
	assert.Equal(t, "example_component", component.Name)
	assert.Equal(t, "example_type", component.Type)
	assert.Equal(t, "example_constraint", component.Constraints)

	// Validate metadata
	assert.NotNil(t, component.Metadata)
	assert.Equal(t, "example_metadata_value", component.Metadata["example_metadata_key"])

	// Validate properties
	assert.NotNil(t, component.Properties)
	assert.Equal(t, "example_property_value", component.Properties["example_property_key"])

	// Validate parameters
	assert.NotNil(t, component.Parameters)
	assert.Equal(t, "example_parameter_value", component.Parameters["example_parameter_key"])

	// Validate routes
	assert.Equal(t, 1, len(component.Routes))
	route := component.Routes[0]
	assert.Equal(t, "example_route", route.Route)
	assert.Equal(t, "example_type", route.Type)
	assert.Equal(t, "example_route_property_value", route.Properties["example_route_property_key"])

	// Validate sidecars
	assert.Equal(t, 1, len(component.Sidecars))
	sidecar := component.Sidecars[0]
	assert.Equal(t, "example_sidecar", sidecar.Name)
	assert.Equal(t, "example_type", sidecar.Type)
	assert.Equal(t, "example_sidecar_property_value", sidecar.Properties["example_sidecar_property_key"])
}

func TestMockRustProviderApply(t *testing.T) {
	config := RustTargetProviderConfig{
		Name:    "mock",
		LibFile: "./target/release/libmock.so",
		LibHash: "26e68667de3d7bfd5ff758191ce4a231800a93f52e1751c10fa0afc7811893cd",
	}
	rustProvider := &RustTargetProvider{}
	err := rustProvider.Init(config)
	assert.Nil(t, err)

	// Create a mock deployment spec and deployment step
	deployment := model.DeploymentSpec{
		SolutionName: "test_solution",
		Targets:      map[string]model.TargetState{},
	}
	step := model.DeploymentStep{
		Target: "test_target",
		Components: []model.ComponentStep{
			{
				Action: model.ComponentUpdate,
				Component: model.ComponentSpec{
					Name: "example_component",
					Type: "example_type",
				},
			},
		},
		Role:    "test_role",
		IsFirst: true,
	}

	// Call the Apply method
	result, err := rustProvider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	// Validate the returned component result
	assert.Equal(t, 1, len(result))

	// Validate the properties of the returned ComponentResultSpec
	componentResult := result["example_component"]
	assert.Equal(t, v1alpha2.OK, componentResult.Status)
	assert.Equal(t, "Component example_component applied successfully", componentResult.Message)
}
