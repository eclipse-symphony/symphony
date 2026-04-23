/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package piccolo

import (
	"context"
	"os/exec"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestPiccoloTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := PiccoloTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestPiccoloTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := PiccoloTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestPiccoloTargetProviderInitEmptyConfig(t *testing.T) {
	config := PiccoloTargetProviderConfig{}
	provider := PiccoloTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}
func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := PiccoloTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}
func TestPiccoloTargetProviderApply(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000/",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)

	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "redis",
		Type: "container",
		Properties: map[string]interface{}{
			"workload.name": "redis-test",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestPiccoloTargetProviderGet(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	// Befor this, TestPiccoloTargetProviderApply must be called.
	TestPiccoloTargetProviderApply(t)
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "redis",
						Type: "container",
						Properties: map[string]interface{}{
							"workload.name": "redis-test",
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "redis",
				Type: "container",
				Properties: map[string]interface{}{
					"workload.name": "redis-test",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
	assert.Equal(t, "redis-test", components[0].Properties["workload.name"])
}
func TestPiccoloTargetProviderApplyDelete(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "redis",
		Type: "container",
		Properties: map[string]interface{}{
			"workload.name": "redis-test",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}
func TestApplyFailed(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)

	// invalid component properties
	component := model.ComponentSpec{
		Name: "",
		Type: "container",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}

	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}
func TestGetFailed(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	// Befor this, TestPiccoloTargetProviderApply must be called.
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "notApplied",
						Type: "container",
						Properties: map[string]interface{}{
							"workload.name": "notApplied-test",
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "notApplied",
				Type: "container",
				Properties: map[string]interface{}{
					"workload.name": "notApplied-test",
				},
			},
		},
	})
	assert.NotNil(t, err)
}
func TestDeleteFailed(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d -p 5000:5000 --name mock_piccolo hbai/piccolo-mock:latest
	cmd := exec.Command("podman", "ps", "-q", "-f", "name=mock_piccolo")
	output, err := cmd.Output()
	assert.Nil(t, err)
	if len(output) == 0 {
		t.Skip("Skipping because mock_picc container is not exist")
	}

	configMap := map[string]string{
		"name": "piccolo",
		"url":  "http://127.0.0.1:5000",
	}
	config, err := PiccoloTargetProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	provider := PiccoloTargetProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "notApplied",
		Type: "container",
		Properties: map[string]interface{}{
			"workload.name": "notApplied-test",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

func TestConformanceSuite(t *testing.T) {
	provider := &PiccoloTargetProvider{}
	err := provider.Init(PiccoloTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
