package kubectl

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestKubectlTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestKubectlTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	provider.Init(config)
	// assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "bytes",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestReadYamlFromUrl(t *testing.T) {
	msgChan, errChan := readYamlFromUrl("https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml")
	totalSize := 0
	for {
		select {
		case data, ok := <-msgChan:
			assert.True(t, ok)
			totalSize += len(data)
		case err, ok := <-errChan:
			assert.True(t, ok)
			if err == io.EOF {
				assert.True(t, totalSize > 10000)
				return
			}
			assert.Nil(t, err)
		}
	}
}
func TestKubectlTargetProviderApply(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL_GATEKEEPER")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL_GATEKEEPER enviornment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "gatekeeper",
		},
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "gatekeepr",
							Type: "yaml.k8s",
							Properties: map[string]string{
								"yaml.url": "https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml",
							},
						},
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestKubectlTargetProviderApplyPolicy(t *testing.T) {
	testPolicy := os.Getenv("TEST_KUBECTL_GATEKEEPER_POLICY")
	if testPolicy == "" {
		t.Skip("Skipping because TEST_KUBECTL_GATEKEEPER_POLICY enviornment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "policies",
		},
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "policies",
							Type: "yaml.k8s",
							Properties: map[string]string{
								"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
							},
						},
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestKubectlTargetProviderDelete(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL_GATEKEEPER")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL_GATEKEEPER enviornment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "gatekeeper",
		},
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "gatekeepr1",
							Type: "yaml.k8s",
							Properties: map[string]string{
								"yaml.url": "https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml",
							},
						},
					},
				},
			},
		},
	}, nil)

	assert.Nil(t, err)
}
func TestKubectlTargetProviderDeletePolicies(t *testing.T) {
	testPolicy := os.Getenv("TEST_KUBECTL_GATEKEEPER_POLICY")
	if testPolicy == "" {
		t.Skip("Skipping because TEST_KUBECTL_GATEKEEPER_POLICY enviornment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "policies",
		},
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "policies",
							Type: "yaml.k8s",
							Properties: map[string]string{
								"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
							},
						},
					},
				},
			},
		},
	}, nil)

	assert.Nil(t, err)
}
