/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

// TestKubectlTargetProviderConfigFromMapNil tests that passing nil to KubectlTargetProviderConfigFromMap returns a valid config
func TestInit(t *testing.T) {
	targetProvider := &MockTargetProvider{}
	err := targetProvider.Init(MockTargetProviderConfig{})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"id": "id",
	}
	provider := MockTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestMockTargetProviderApply(t *testing.T) {
	provider := &MockTargetProvider{}
	err := provider.Init(MockTargetProviderConfig{})
	assert.Nil(t, err)

	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "name",
			Scope: "default",
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: "update",
				Component: model.ComponentSpec{
					Name: "name",
				},
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: "delete",
				Component: model.ComponentSpec{
					Name: "name",
				},
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestMockTargetProviderGet(t *testing.T) {
	provider := &MockTargetProvider{}
	err := provider.Init(MockTargetProviderConfig{})
	assert.Nil(t, err)

	_, err = provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Nil(t, err)
}
