/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package arm

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInitWithNil(t *testing.T) {
	provider := ArmTargetProvider{}
	err := provider.Init(nil)
	assert.NotNil(t, err)
}

func TestApply(t *testing.T) {
	provider := ArmTargetProvider{}
	err := provider.Init(ArmTargetProviderConfig{
		SubscriptionId: "",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test",
		Properties: map[string]interface{}{
			"resourceGroup": "test",
			"location":      "westus",
			"template": UrlOrJson{
				JSON: map[string]interface{}{
					"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
					"contentVersion": "1.0.0.0",
					"parameters":     map[string]interface{}{},
					"variables":      map[string]interface{}{},
					"resources": []interface{}{
						map[string]interface{}{
							"name":       "thestroageaaa",
							"type":       "Microsoft.Storage/storageAccounts",
							"apiVersion": "2021-04-01",
							"tags": map[string]string{
								"displayName": "thestroageaaa",
							},
							"location": "EastUS",
							"kind":     "StorageV2",
							"sku": map[string]string{
								"name": "Premium_LRS",
								"tier": "Premium",
							},
						},
					},
				},
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
			ObjectMeta: model.ObjectMeta{
				Name: "deepseek",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	components := []model.ComponentStep{
		{
			Action:    model.ComponentUpdate,
			Component: component,
		},
	}
	step := model.DeploymentStep{
		Components: components,
	}
	result, err := provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
	assert.Equal(t, result["test"].Status, v1alpha2.Updated)

	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}

	result, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
	assert.Equal(t, result["test"].Status, v1alpha2.Deleted)
}
