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
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	certProvider "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// MockCertProvider implements both ICertProvider and IProvider interfaces for testing
type MockCertProvider struct {
}

// Implement IProvider interface
func (m *MockCertProvider) Init(config providers.IProviderConfig) error {
	return nil
}

// Implement ICertProvider interface
func (m *MockCertProvider) CreateCert(ctx context.Context, req certProvider.CertRequest) error {
	return nil
}

func (m *MockCertProvider) DeleteCert(ctx context.Context, targetName, namespace string) error {
	return nil
}

func (m *MockCertProvider) GetCert(ctx context.Context, targetName, namespace string) (*certProvider.CertResponse, error) {
	// Return mock certificate data that matches what the test expects
	return &certProvider.CertResponse{
		PublicKey:    "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJANlqGR0GwHpNMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNVBAMMCWxv\nY2FsaG9zdDAeFw0yMzA5MjIwOTE1MzRaFw0yNDA5MjEwOTE1MzRaMBQxEjAQBgNV\nBAMMCWxvY2FsaG9zdDBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7rsI/sNlQmFD+\nkab4TGXYXfVBnPJnZvajRvHxiTPfkTfNWVE/6LiYh8WNk6BW8jXMf5jf+DBSjvKW\n8P3VNhv5AgMBAAEwDQYJKoZIhvcNAQELBQADQQBJ4v4Y7HdXaajdRP3IgJyVgKQL\nIvdzP8qfmYCAf0+Dg4Gx8kfyze89/+P8dGwBCU6VzQGsv6K4FlT0gWg=\n-----END CERTIFICATE-----",
		PrivateKey:   "-----BEGIN PRIVATE KEY-----\nMIIBVAIBADANBgkqhkiG9w0BAQEFAASCAT4wggE6AgEAAkEAu67CP7DZUJHQ+pGm\n+Exl2F31QZzyZ2b2o0bx8Ykz35E3zVlRP+i4mIfFjZOgVvI1zH+Y3/gwUo7ylvD9\n1TYb+QIDAQABAkEAjLP5+VKam+XlSlJiuk8VSwZJvN1Ba2z3o7bq7J7z6KfkfWo3\nUvLL+bt+5YfzpxjzHur8YKK+n8KhSz5WPLwLsQIhAOO+7v7FL1a6K+FmJ2fPGadU\nqcY7FKjP3LTnh35pjNn1AiEA2gL7YKzsKmK2ZvJukM8eJSlrL7JJJLVcLhHmYzXa\nqr0CIGl+ADVLJiVZyCgJiXUgD7qq5Gi7CWGm2RJef8Gtn2aFAiBU5aAB/j3NKt7g\nkMHRnznYKBdb8tKUsQZgxWY1KXDoNwIgSPqzOgCpG6UNhG2jgL9JyGG7JJ1b7PfJ\nD8wEgEJWj8Y=\n-----END PRIVATE KEY-----",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		SerialNumber: "123456789",
	}, nil
}

func (m *MockCertProvider) CheckCertStatus(ctx context.Context, targetName, namespace string) (*certProvider.CertStatus, error) {
	return &certProvider.CertStatus{
		Ready:       true,
		Reason:      "Certificate is ready",
		Message:     "Mock certificate is ready for use",
		LastUpdate:  time.Now(),
		NextRenewal: time.Now().Add(7 * 24 * time.Hour),
	}, nil
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

	// Create mock certificate provider and initialize it
	mockCertProvider := &MockCertProvider{}
	mockCertProvider.Init(nil) // Initialize the provider

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
					"providers.cert":            "cert-provider",
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
			"mem-state":     &stateProvider,
			"cert-provider": mockCertProvider, // Add certificate provider to the providers map
		},
	}, &pubSubProvider)
	vendor.Config.Properties["useJobManager"] = "true"
	vendor.TargetsManager.TargetValidator = validation.NewTargetValidator(nil, nil)
	vendor.TargetsManager.SecretProvider = &secretProvider
	// Certificate provider should now be automatically set during Init() due to the providers map
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

	// Pre-create a mock certificate in ready state to simulate cert-manager behavior
	certObj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "target1-v1",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"secretName": "target1-v1-tls",
		},
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "True",
				},
			},
		},
	}

	// Store the mock certificate in state
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   "target1-v1",
			Body: certObj,
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}
	vendor.TargetsManager.StateProvider.Upsert(context.Background(), upsertRequest)

	resp = vendor.onGetCert(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Parameters: map[string]string{
			"__name": "target1-v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	var certResponse map[string]interface{}
	json.Unmarshal(resp.Body, &certResponse)
	assert.Contains(t, certResponse, "public")
	assert.Contains(t, certResponse, "private")
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
