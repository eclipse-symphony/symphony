package adu

import (
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := ADUTargetProvider{}
	err := provider.Init(nil)
	assert.NotNil(t, err)
}

func TestConformanceSuite(t *testing.T) {
	provider := &ADUTargetProvider{}
	err := provider.Init(ADUTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
