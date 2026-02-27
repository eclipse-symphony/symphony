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
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	mockconfig "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	memorykeylock "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/keylock/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	mocksecret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	coalogcontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createSolutionVendor() SolutionVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	configProvider := mockconfig.MockConfigProvider{}
	configProvider.Init(mockconfig.MockConfigProviderConfig{})
	secretProvider := mocksecret.MockSecretProvider{}
	secretProvider.Init(mocksecret.MockSecretProviderConfig{})
	keyLockProvider := memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{})
	vendor := SolutionVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "solution-manager",
				Type: "managers.symphony.solution",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
					"providers.config":          "mock-config",
					"providers.secret":          "mock-secret",
					"providers.keylock":         "mem-keylock",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
					"mem-keylock": {
						Type:   "providers.keylock.memory",
						Config: memorykeylock.MemoryKeyLockProviderConfig{},
					},
					"mock-config": {
						Type:   "providers.config.mock",
						Config: mockconfig.MockConfigProviderConfig{},
					},
					"mock-secret": {
						Type:   "providers.secret.mock",
						Config: mocksecret.MockSecretProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"solution-manager": {
			"mem-state":   &stateProvider,
			"mem-keylock": &keyLockProvider,
			"mock-config": &configProvider,
			"mock-secret": &secretProvider,
		},
	}, nil)
	return vendor
}

func createDockerDeployment(id string) model.DeploymentSpec {
	return model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "instance-docker",
				Annotations: map[string]string{
					"Guid": uuid.New().String(),
				},
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "coma",
						Properties: map[string]interface{}{
							"container.image": "redis",
						},
					},
				},
			},
		},
		Assignments: map[string]string{
			"docker": "{coma}",
		},
		Targets: map[string]model.TargetState{
			"docker": {
				Spec: &model.TargetSpec{
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
		},
	}
}

func createDeployment2Mocks1Target(id string) model.DeploymentSpec {
	return model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "instance1",
				Annotations: map[string]string{
					"Guid": uuid.New().String(),
				},
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
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
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
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
		},
	}
}
func TestSolutionEndpoints(t *testing.T) {
	vendor := createSolutionVendor()
	vendor.Route = "solution"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 4, len(endpoints))
}

func TestSolutionInfo(t *testing.T) {
	vendor := createSolutionVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestSolutionGetInstances(t *testing.T) {
	vendor := createSolutionVendor()
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

	resp = vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var components []model.ComponentSpec
	err = json.Unmarshal(resp.Body, &components)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(components))
	assert.Equal(t, "a", components[0].Name)
	assert.Equal(t, "mock", components[0].Type)
	assert.Equal(t, "b", components[1].Name)
	assert.Equal(t, "mock", components[1].Type)
}
func TestSolutionApply(t *testing.T) {
	vendor := createSolutionVendor()
	deployment := createDeployment2Mocks1Target(uuid.New().String())
	data, _ := json.Marshal(deployment)
	resp := vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var components []model.ComponentSpec
	err := json.Unmarshal(resp.Body, &components)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	resp = vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var summary model.SummarySpec
	err = json.Unmarshal(resp.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, 1, summary.TargetCount)
	assert.Equal(t, "OK", summary.TargetResults["T1"].Status)

	resp = vendor.onApplyDeployment(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Body:    data,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	err = json.Unmarshal(resp.Body, &components)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(components))
}
func TestSolutionRemove(t *testing.T) {
	vendor := createSolutionVendor()
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
func TestSolutionReconcileDocker(t *testing.T) {
	testDocker := os.Getenv("TEST_DOCKER_RECONCILE")
	if testDocker == "" {
		t.Skip("Skipping because TEST_DOCKER_RECONCILE environment variable is not set")
	}
	var summary model.SummarySpec
	vendor := createSolutionVendor()

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
func TestSolutionReconcile(t *testing.T) {
	var summary model.SummarySpec
	vendor := createSolutionVendor()

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
	deployment.Solution.Spec.Components = append(deployment.Solution.Spec.Components, model.ComponentSpec{Name: "c", Type: "mock"})
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
	deployment.Solution.Spec.Components = deployment.Solution.Spec.Components[1:]
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
func TestSolutionQueue(t *testing.T) {
	vendor := createSolutionVendor()
	resp := vendor.onQueue(v1alpha2.COARequest{
		Method:     fasthttp.MethodGet,
		Parameters: map[string]string{},
		Context:    context.Background(),
	})
	assert.Equal(t, v1alpha2.BadRequest, resp.State)

	resp = vendor.onQueue(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"instance": "instance1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.NotFound, resp.State)

}
func TestSolutionQueueInstanceUpdate(t *testing.T) {
	vendor := createSolutionVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	succeededCount := 0
	sig := make(chan bool)
	ctx := context.TODO()
	correlationId := uuid.New().String()
	resourceId := uuid.New().String()
	ctx = coalogcontexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)
	vendor.Context.Subscribe("job", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			assert.NotEqual(t, ctx, event.Context)
			assert.NotNil(t, event.Context)
			diagCtx, ok := event.Context.Value(coalogcontexts.DiagnosticLogContextKey).(*coalogcontexts.DiagnosticLogContext)
			assert.True(t, ok)
			assert.NotNil(t, diagCtx)
			assert.Equal(t, correlationId, diagCtx.GetCorrelationId())
			assert.Equal(t, resourceId, diagCtx.GetResourceId())

			var job v1alpha2.JobData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &job)
			assert.Nil(t, err)
			assert.Equal(t, "instance", event.Metadata["objectType"])
			assert.Equal(t, "scope1", event.Metadata["namespace"])
			assert.Equal(t, "instance1", job.Id)
			assert.Equal(t, v1alpha2.JobUpdate, job.Action)
			succeededCount += 1
			sig <- true
			return nil
		},
	})
	resp := vendor.onQueue(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Parameters: map[string]string{
			"instance":  "instance1",
			"target":    "false",
			"namespace": "scope1",
		},
		Context: ctx,
	})
	<-sig
	assert.Equal(t, v1alpha2.OK, resp.State)
	// wait for the job to be processed
	time.Sleep(time.Second)
	assert.Equal(t, 1, succeededCount)
}
func TestSolutionQueueTargetUpdate(t *testing.T) {
	vendor := createSolutionVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	sig := make(chan bool)
	succeededCount := 0
	vendor.Context.Subscribe("job", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			var job v1alpha2.JobData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &job)
			assert.Nil(t, err)
			assert.Equal(t, "target", event.Metadata["objectType"])
			assert.Equal(t, "scope1", event.Metadata["namespace"])
			assert.Equal(t, "target1", job.Id)
			assert.Equal(t, v1alpha2.JobDelete, job.Action)
			succeededCount += 1
			sig <- true
			return nil
		},
	})
	resp := vendor.onQueue(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Parameters: map[string]string{
			"instance":  "target1",
			"target":    "true",
			"namespace": "scope1",
			"delete":    "true",
		},
		Context: context.Background(),
	})
	<-sig
	assert.Equal(t, v1alpha2.OK, resp.State)
	// wait for the job to be processed
	time.Sleep(time.Second)
	assert.Equal(t, 1, succeededCount)
}
