/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/cert"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// MockCertManager implements a mock certificate manager for testing
type MockCertManager struct{}

func (m *MockCertManager) GetWorkingCert(ctx context.Context, targetName, namespace string) (string, string, error) {
	// Return mock certificate data
	public := "-----BEGIN CERTIFICATE----- mock-public-cert-data -----END CERTIFICATE-----"
	private := "-----BEGIN PRIVATE KEY----- mock-private-key-data -----END PRIVATE KEY-----"
	return public, private, nil
}

func (m *MockCertManager) CreateWorkingCert(ctx context.Context, targetName, namespace string) error {
	return nil
}

func (m *MockCertManager) DeleteWorkingCert(ctx context.Context, targetName, namespace string) error {
	return nil
}

func (m *MockCertManager) CheckCertificateReady(ctx context.Context, targetName, namespace string) (bool, error) {
	return true, nil
}

func TestTargetsEndpoints(t *testing.T) {
	vendor := createTargetsVendor()
	vendor.Route = "targets"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 8, len(endpoints))
}

func TestTargetsInfo(t *testing.T) {
	vendor := createTargetsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func createTargetsVendor() TargetsVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	secretProvider := mock.MockSecretProvider{}
	secretProvider.Init(mock.MockSecretProviderConfig{Name: "test-secret"})

	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := TargetsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "targets-manager",
				Type: "managers.symphony.targets",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"targets-manager": {
			"mem-state": &stateProvider,
		},
	}, &pubSubProvider)
	vendor.Config.Properties["useJobManager"] = "true"
	vendor.TargetsManager.TargetValidator = validation.NewTargetValidator(nil, nil)
	vendor.TargetsManager.SecretProvider = &secretProvider

	// Set up mock CertManager - create a real CertManager but with mock providers
	mockCertManager := &cert.CertManager{
		StateProvider:  &stateProvider,
		SecretProvider: &secretProvider,
	}
	vendor.CertManager = mockCertManager

	return vendor
}
func TestTargetsOnRegistry(t *testing.T) {
	vendor := createTargetsVendor()
	target := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target1-v1",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config: map[string]string{
								"id": uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":       "target1-v1",
			"with-binding": "staging",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	var targets model.TargetState
	json.Unmarshal(resp.Body, &targets)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, "target1-v1", targets.ObjectMeta.Name)
	assert.Equal(t, 1, len(targets.Spec.Topologies))

	resp = vendor.onRegistry(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	var targetsList []model.TargetState
	json.Unmarshal(resp.Body, &targetsList)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, 1, len(targetsList))

	resp = vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

// func TestTargetsOnBootstrap(t *testing.T) {
// 	vendor := createTargetsVendor()
// 	authRequest := AuthRequest{
// 		UserName: "symphony-test",
// 		Password: "",
// 	}
// 	data, _ := json.Marshal(authRequest)
// 	resp := vendor.onBootstrap(v1alpha2.COARequest{
// 		Method:  fasthttp.MethodPost,
// 		Body:    data,
// 		Context: context.Background(),
// 	})
// 	assert.Equal(t, v1alpha2.OK, resp.State)
// 	var authResponse AuthResponse
// 	json.Unmarshal(resp.Body, &authResponse)
// 	assert.NotNil(t, authResponse.AccessToken)
// 	assert.Equal(t, "Bearer", authResponse.TokenType)
// }

func TestTargetsOnStatus(t *testing.T) {
	vendor := createTargetsVendor()

	target := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target1-v1",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config: map[string]string{
								"id": uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":       "target1-v1",
			"with-binding": "staging",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	dict := map[string]interface{}{
		"status": map[string]interface{}{
			"properties": map[string]string{
				"testkey": "testvalue",
			},
		},
	}
	data, _ = json.Marshal(dict)

	resp = vendor.onStatus(v1alpha2.COARequest{
		Method: fasthttp.MethodPut,
		Body:   data,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	var targetState model.TargetState
	json.Unmarshal(resp.Body, &targetState)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, "testvalue", targetState.Status.Properties["testkey"])
}
func TestTargetsOnHeartbeats(t *testing.T) {
	vendor := createTargetsVendor()

	target := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target1-v1",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config: map[string]string{
								"id": uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":       "target1-v1",
			"with-binding": "staging",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onHeartBeat(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	var targetState model.TargetState
	json.Unmarshal(resp.Body, &targetState)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.NotNil(t, targetState.Status.Properties["ping"])
}
func TestTargetsOnGetCert(t *testing.T) {
	vendor := createTargetsVendor()

	// Register a target first
	target := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target1-v1",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config: map[string]string{
								"id": uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":       "target1-v1",
			"with-binding": "staging",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	ctx := context.Background()
	targetName := "target1-v1"
	namespace := "default"

	// Use CertManager.CreateWorkingCert to create certificate
	err := vendor.CertManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Test the onGetCert endpoint
	resp = vendor.onGetCert(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Parameters: map[string]string{
			"__name": targetName,
		},
		Context: ctx,
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	// Verify response contains certificate data
	var certResponse map[string]interface{}
	json.Unmarshal(resp.Body, &certResponse)
	assert.Contains(t, certResponse, "public")
	assert.Contains(t, certResponse, "private")

	// Verify certificate data is not empty and follows MockSecretProvider behavior
	// MockSecretProvider returns "secretName>>fieldName" format
	public := certResponse["public"].(string)
	private := certResponse["private"].(string)
	assert.NotEmpty(t, public)
	assert.NotEmpty(t, private)
	assert.Contains(t, public, "target1-v1-tls>>tls.crt")
	assert.Contains(t, private, "target1-v1-tls>>tls.key")
}

func TestTargetsOnUpdateTopology(t *testing.T) {
	vendor := createTargetsVendor()

	target := model.TargetState{
		Spec: &model.TargetSpec{
			DisplayName: "target1-v1",
			Topologies: []model.TopologySpec{
				{
					Bindings: []model.BindingSpec{
						{
							Role:     "mock",
							Provider: "providers.target.mock",
							Config: map[string]string{
								"id": uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":       "target1-v1",
			"with-binding": "staging",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	topology := model.TopologySpec{
		Bindings: []model.BindingSpec{
			{
				Role:     "updated-mock",
				Provider: "providers.target.updated-mock",
				Config: map[string]string{
					"id":     uuid.New().String(),
					"update": "true",
				},
			},
		},
	}
	topologyData, _ := json.Marshal(topology)

	resp = vendor.onUpdateTopology(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   topologyData,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	var updateResponse map[string]interface{}
	json.Unmarshal(resp.Body, &updateResponse)
	assert.Equal(t, "topology updated successfully", updateResponse["result"])
}

func TestTargetWrongMethod(t *testing.T) {
	vendor := createTargetsVendor()
	resp := vendor.onRegistry(v1alpha2.COARequest{
		Method:  fasthttp.MethodPut,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
	resp = vendor.onStatus(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)

	resp = vendor.onHeartBeat(v1alpha2.COARequest{
		Method:  fasthttp.MethodPut,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)

	resp = vendor.onGetCert(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)

	resp = vendor.onUpdateTopology(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
