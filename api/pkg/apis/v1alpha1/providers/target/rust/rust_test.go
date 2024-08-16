package rust

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockRustProviderGetValidationRule(t *testing.T) {
	config := ProviderConfig{}
	rustProvider, err := NewRustTargetProvider("mock", "./target/release/libmock.so")
	assert.Nil(t, err)
	defer rustProvider.Close()

	err = rustProvider.Init(config)
	assert.Nil(t, err)

	// Example usage of GetValidationRule
	rule := rustProvider.GetValidationRule(context.Background())
	assert.Equal(t, "example_type", rule.RequiredComponentType)
	assert.Equal(t, 1, len(rule.ComponentValidationRule.RequiredProperties))
	assert.Equal(t, "required_property_name", rule.ComponentValidationRule.RequiredProperties[0])
}
