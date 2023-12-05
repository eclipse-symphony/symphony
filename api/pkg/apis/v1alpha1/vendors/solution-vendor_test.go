/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/managers/solution"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/mock"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createVendor() SolutionVendor {
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := solution.SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		StateProvider: stateProvider,
	}
	vendor := SolutionVendor{
		SolutionManager: &manager,
	}
	return vendor
}
func createDockerDeployment(id string) model.DeploymentSpec {
	return model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "instance-docker",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "coma",
					Properties: map[string]interface{}{
						"container.image": "redis",
					},
				},
			},
		},
		Assignments: map[string]string{
			"docker": "{coma}",
		},
		Targets: map[string]model.TargetSpec{
			"docker": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "providers.target.docker",
								Config: map[string]string{
									"name": id,
								},
							},
						},
					},
				},
			},
		},
	}
}
func createDeployment2Mocks1Target(id string) model.DeploymentSpec {
	return model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "instance1",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "a",
					Type: "mock",
				},
				{
					Name: "b",
					Type: "mock",
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetSpec{
			"T1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "mock",
								Provider: "providers.target.mock",
								Config: map[string]string{
									"id": id,
								},
							},
						},
					},
				},
			},
		},
	}
}
func TestGetInstances(t *testing.T) {
	vendor := createVendor()
	deployment := createDeployment2Mocks1Target(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var summary model.SummarySpec
	err := json.Unmarshal(resp.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, "OK", summary.TargetResults["T1"].Status)
}
func TestApply(t *testing.T) {
	vendor := createVendor()
	deployment := createDeployment2Mocks1Target(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var summary model.SummarySpec
	err := json.Unmarshal(resp.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, 1, summary.TargetCount)
	assert.Equal(t, "OK", summary.TargetResults["T1"].Status)
}
func TestRemove(t *testing.T) {
	vendor := createVendor()
	deployment := createDeployment2Mocks1Target(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var summary model.SummarySpec
	err := json.Unmarshal(resp.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, 1, summary.TargetCount)
	assert.Equal(t, false, summary.Skipped)
}
func TestReconcileDocker(t *testing.T) {
	testDocker := os.Getenv("TEST_DOCKER_RECONCILE")
	if testDocker == "" {
		t.Skip("Skipping because TEST_DOCKER_RECONCILE environment variable is not set")
	}
	var summary model.SummarySpec
	vendor := createVendor()

	// deploy
	deployment := createDockerDeployment(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
		Parameters: map[string]string{
			"delete": "true",
		},
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	json.Unmarshal(resp.Body, &summary)
	assert.False(t, summary.Skipped)
}
func TestReconcile(t *testing.T) {
	var summary model.SummarySpec
	vendor := createVendor()

	// deploy
	deployment := createDeployment2Mocks1Target(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	json.Unmarshal(resp.Body, &summary)
	assert.False(t, summary.Skipped)

	// try deploy agin, this should be skipped
	resp = vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	json.Unmarshal(resp.Body, &summary)
	assert.True(t, summary.Skipped)

	//now update the deployment and add one more component
	deployment.Solution.Components = append(deployment.Solution.Components, model.ComponentSpec{Name: "c", Type: "mock"})
	deployment.Assignments["T1"] = "{a}{b}{c}"
	data, _ = json.Marshal(deployment)

	//now deploy agian, this should trigger a new deployment
	resp = vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &summary)
	assert.Nil(t, err)
	assert.False(t, summary.Skipped)

	//now apply the deployment again, this should be skipped
	resp = vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	json.Unmarshal(resp.Body, &summary)
	assert.True(t, summary.Skipped)

	//now update again to remove the first component
	deployment.Solution.Components = deployment.Solution.Components[1:]
	deployment.Assignments["T1"] = "{b}{c}"
	data, _ = json.Marshal(deployment)

	//now check if update is needed again
	resp = vendor.onReconcile(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	json.Unmarshal(resp.Body, &summary)
	assert.False(t, summary.Skipped)
}
