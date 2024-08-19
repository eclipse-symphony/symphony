package rust

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestMockRustProviderGetValidationRule(t *testing.T) {
	config := RustTargetProviderConfig{}
	rustProvider, err := NewRustTargetProvider("mock", "./target/release/libmock.so")
	assert.Nil(t, err)
	defer rustProvider.Close()

	err = rustProvider.Init(config)
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
	config := RustTargetProviderConfig{}
	rustProvider, err := NewRustTargetProvider("mock", "./target/release/libmock.so")
	assert.Nil(t, err)
	defer rustProvider.Close()

	err = rustProvider.Init(config)
	assert.Nil(t, err)

	// Create a mock deployment spec and component steps
	deployment := model.DeploymentSpec{}
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
	assert.Equal(t, "example_metadata_key", component.Metadata["example_metadata_key"])
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
