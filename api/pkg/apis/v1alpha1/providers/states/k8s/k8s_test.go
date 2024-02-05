/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/stretchr/testify/assert"
)

func TestK8sStateProviderConfigFromMapNil(t *testing.T) {
	_, err := K8sStateProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sStateProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sStateProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "Bad",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	testEnabled := os.Getenv("TEST_MINIKUBE_ENABLED")
	if testEnabled == "" {
		t.Skip("Skipping because TEST_MINIKUBE_ENABLED enviornment variable is not set")
	}
	config := K8sStateProviderConfig{
		ConfigType: "path",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "bytes",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := K8sStateProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := K8sStateProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestUpsert(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata": map[string]interface{}{
					"name": "a1",
				},
				"spec": model.ActivationSpec{
					Campaign: "c1",
					Name:     "a1",
					Stage:    "s1",
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", id)
}
func TestList(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
			Body: model.ActivationSpec{
				Campaign: "c1",
				Name:     "a1",
				Stage:    "s1",
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "a1", entries[0].ID)

	assert.Nil(t, err)
	entries, _, err = provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]string{
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "a1", entries[0].ID)
}
func TestDelete(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a1",
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
}
func TestGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata": map[string]interface{}{
					"name": "a1",
				},
				"spec": model.ActivationSpec{
					Campaign: "c1",
					Name:     "a1",
					Stage:    "s1",
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	item, err := provider.Get(context.Background(), states.GetRequest{
		ID: "a1",
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", item.ID)
}

func TestUpsertWithState(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	activationStatus := model.ActivationStatus{
		Stage: "s2",
	}
	j, _ := json.Marshal(activationStatus)
	var dict map[string]interface{}
	json.Unmarshal(j, &dict)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata": map[string]interface{}{
					"name": "a1",
				},
				"spec": model.ActivationSpec{
					Campaign: "c1",
					Name:     "a1",
					Stage:    "s1",
				},
				"status": dict,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", id)

	entry, err := provider.Get(context.Background(), states.GetRequest{
		ID: "a1",
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})

	assert.Nil(t, err)
	j, _ = json.Marshal(entry.Body.(map[string]interface{})["status"])
	var rStatus model.ActivationStatus
	json.Unmarshal(j, &rStatus)
	assert.Equal(t, "s2", rStatus.Stage)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a1",
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
}
func TestUpsertWithStateOnly(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkActivationCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a2",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup,
				"kind":       "Activation",
				"metadata": map[string]interface{}{
					"name": "a2",
				},
				"status": map[string]interface{}{
					"properties": map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a2", id)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a2",
		Metadata: map[string]string{
			"scope":    "default",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
}

func runKubectl(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	return cmd.Output()
}

func checkActivationCRDApplied() error {
	// Check that the CRD is applied
	_, err := runKubectl("get", "crd", "activations.workflow.symphony")
	if err != nil {
		// apply the CRD api/pkg/apis/v1alpha1/providers/states/k8s/k8s_test.go
		ProjectPath := os.Getenv("REPOPATH")
		activationYamlPath := ProjectPath + "/k8s/config/oss/crd/bases/workflow.symphony_activations.yaml"
		if _, err := os.Stat(activationYamlPath); err != nil {
			return err
		}
		_, err = runKubectl("apply", "-f", activationYamlPath)
		if err != nil {
			return err
		}
		// Wait for the CRD to be applied
		time.Sleep(10 * time.Second)
	}
	return nil
}
