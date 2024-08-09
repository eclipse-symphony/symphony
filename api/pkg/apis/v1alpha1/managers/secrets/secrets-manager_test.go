package secrets

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	mocksecret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/stretchr/testify/assert"
)

func TestSecretsManager_Init(t *testing.T) {
	secretProvider := mocksecret.MockSecretProvider{}
	secretProvider.Init(mocksecret.MockSecretProviderConfig{})
	manager := SecretsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.secrets": "SecretProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["SecretProvider"] = &secretProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
}

func TestObjectFieldGetWithProviderSpecified(t *testing.T) {
	provider := mocksecret.MockSecretProvider{}
	err := provider.Init(mocksecret.MockSecretProviderConfig{})
	manager := SecretsManager{
		SecretProviders: map[string]secret.ISecretProvider{
			"mock": &provider,
		},
	}
	assert.Nil(t, err)
	val, err := manager.Get("mock::obj", "field", nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj>>field", val)
}

func TestObjectFieldGetWithOneProvider(t *testing.T) {
	provider := mocksecret.MockSecretProvider{}
	err := provider.Init(mocksecret.MockSecretProviderConfig{})
	manager := SecretsManager{
		SecretProviders: map[string]secret.ISecretProvider{
			"mock": &provider,
		},
	}
	val, err := manager.Get("obj", "field", nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj>>field", val)
}

func TestObjectFieldGetWithMoreProviders(t *testing.T) {
	provider1 := mocksecret.MockSecretProvider{}
	err := provider1.Init(mocksecret.MockSecretProviderConfig{})
	assert.Nil(t, err)
	provider2 := mocksecret.MockSecretProvider{}
	err = provider2.Init(mocksecret.MockSecretProviderConfig{})
	assert.Nil(t, err)
	manager := SecretsManager{
		SecretProviders: map[string]secret.ISecretProvider{
			"mock1": &provider1,
			"mock2": &provider2,
		},
		Precedence: []string{"mock1", "mock2"},
	}
	assert.Nil(t, err)
	val, err := manager.Get("obj", "field", nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj>>field", val)
}
