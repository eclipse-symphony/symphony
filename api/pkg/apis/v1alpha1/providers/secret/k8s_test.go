package secret

import (
	"context"
	"os"
	"testing"

	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestK8sStateProviderConfigFromMapNil(t *testing.T) {
	_, err := K8sSecretProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sStateProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sSecretProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := K8sSecretProviderConfig{
		ConfigType: "Bad",
	}
	provider := K8sSecretProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	testEnabled := os.Getenv("TEST_MINIKUBE_ENABLED")
	if testEnabled == "" {
		t.Skip("Skipping because TEST_MINIKUBE_ENABLED enviornment variable is not set")
	}
	config := K8sSecretProviderConfig{
		ConfigType: "path",
	}
	provider := K8sSecretProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := K8sSecretProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := K8sSecretProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := K8sSecretProviderConfig{
		ConfigType: "bytes",
	}
	provider := K8sSecretProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := K8sSecretProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := K8sSecretProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestK8sSecretProvider_Get(t *testing.T) {
	// Create a fake client to mock API calls
	clientset := fake.NewSimpleClientset()

	// Create a test secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"test-field": []byte("test-value"),
		},
	}

	// Create the secret using the fake client
	_, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secret, metav1.CreateOptions{})
	assert.Nil(t, err)

	// Create a K8sSecretProvider
	provider := &K8sSecretProvider{
		Clientset: clientset,
		Config:    K8sSecretProviderConfig{},
	}

	// Call the Get method
	evalContext := coa_utils.EvaluationContext{
		Namespace: "default",
	}
	value, err := provider.Read("test-secret", "test-field", evalContext)
	assert.Nil(t, err)

	// Check the value
	assert.Equal(t, "test-value", value)
}
