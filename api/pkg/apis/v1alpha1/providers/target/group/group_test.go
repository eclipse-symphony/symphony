package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := GroupTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

func TestGroupTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := GroupTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestGroupTargetProviderInitEmptyConfig(t *testing.T) {
	config := GroupTargetProviderConfig{}
	provider := GroupTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}
