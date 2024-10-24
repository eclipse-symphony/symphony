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

func TestActivationUpsert(t *testing.T) {
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
					Stage:    "s1",
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", id)
}
func TestActivationList(t *testing.T) {
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
				Stage:    "s1",
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "a1", entries[0].ID)

	assert.Nil(t, err)
	entries, _, err = provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "a1", entries[0].ID)
}
func TestActivationDelete(t *testing.T) {
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
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a1",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
}
func TestActivationGet(t *testing.T) {
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
					Stage:    "s1",
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	item, err := provider.Get(context.Background(), states.GetRequest{
		ID: "a1",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", item.ID)
}

func TestActivationUpsertWithState(t *testing.T) {
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
		StageHistory: []model.StageStatus{
			{
				Stage: "s2",
			},
		},
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
					Stage:    "s1",
				},
				"status": dict,
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a1", id)

	entry, err := provider.Get(context.Background(), states.GetRequest{
		ID: "a1",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})

	assert.Nil(t, err)
	j, _ = json.Marshal(entry.Body.(map[string]interface{})["status"])
	var rStatus model.ActivationStatus
	json.Unmarshal(j, &rStatus)
	assert.Equal(t, "s2", rStatus.StageHistory[0].Stage)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a1",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
}
func TestActivationUpsertWithStateOnly(t *testing.T) {
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
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "${{$activation()}}"}}`, model.WorkflowGroup),
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a2", id)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "a2",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})
	assert.Nil(t, err)
}

func TestCatalogSpecFilter(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkCatalogCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "c2",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup,
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": "c2",
				},
				"spec": model.CatalogSpec{
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
		FilterType:  "spec",
		FilterValue: `[?(@.properties.foo=="bar")]`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "c2",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)
}
func TestCatalogLabelFilter(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkCatalogCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "c2",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup,
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": "c2",
					"labels": map[string]interface{}{
						"foo":  "bar",
						"foo2": "bar2",
					},
				},
				"spec": model.CatalogSpec{
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
		FilterType:  "label",
		FilterValue: `foo=bar,foo2=bar2`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "c2",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)
}
func TestCatalogFieldFilter(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkCatalogCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "c2",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup,
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": "c2",
				},
				"spec": model.CatalogSpec{
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
		FilterType:  "field",
		FilterValue: `metadata.name==c2`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "c2",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)
}

func TestCatalogStatusFilter(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkCatalogCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "c2",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup,
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": "c2",
				},
				"spec": model.CatalogSpec{
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{ // An update is issued as in initial insert status is ignored. TODO: Isn't that a bug?
		Value: states.StateEntry{
			ID: "c2",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup,
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": "c2",
				},
				"spec": model.CatalogSpec{
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)

	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
		FilterType:  "status",
		FilterValue: `[?(@.properties.foo=="bar")]`,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "c2",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	assert.Nil(t, err)
}

func runKubectl(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	return cmd.Output()
}

func checkCatalogCRDApplied() error {
	// Check that the CRD is applied
	_, err := runKubectl("get", "crd", "catalogs.federation.symphony")
	if err != nil {
		// apply the CRD api/pkg/apis/v1alpha1/providers/states/k8s/k8s_test.go
		ProjectPath := os.Getenv("REPOPATH")
		activationYamlPath := ProjectPath + "/k8s/config/oss/crd/bases/federation.symphony_catalogs.yaml"
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

func checkTargetCRDApplied() error {
	// Check that the CRD is applied
	_, err := runKubectl("get", "crd", "targets.fabric.symphony")
	if err != nil {
		// apply the CRD api/pkg/apis/v1alpha1/providers/states/k8s/k8s_test.go
		ProjectPath := os.Getenv("REPOPATH")
		activationYamlPath := ProjectPath + "/k8s/config/oss/crd/bases/fabric.symphony_targets.yaml"
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

func TestTargetUpsert(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s123",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s123", id)
}

func TestTargetList(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: model.TargetSpec{
				Properties: map[string]string{
					"foo": "bar2",
				},
			},
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	entries, _, err := provider.List(context.Background(), states.ListRequest{
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "s123", entries[0].ID)
}

func TestTargetDelete(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "s123",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
}
func TestTargetGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s123",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s123",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	item, err := provider.Get(context.Background(), states.GetRequest{
		ID: "s123",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s123", item.ID)
}
func TestTargetUpSertWithState(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	id, err := provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s234",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s234",
				},
				"spec": model.TargetSpec{
					Properties: map[string]string{
						"foo": "bar2",
					},
				},
				"status": map[string]interface{}{
					"properties": map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "s234", id)
}
func TestTargetUpSertWithStateOnly(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S_STATE")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S_STATE enviornment variable is not set")
	}
	err := checkTargetCRDApplied()
	assert.Nil(t, err)
	provider := K8sStateProvider{}
	err = provider.Init(K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)
	_, err = provider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "s234",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup,
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "s234",
				},
				"status": map[string]interface{}{
					"properties": map[string]string{
						"foo": "bar2",
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Target", "metadata": {"name": "${{$target()}}"}}`, model.FabricGroup),
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	// Target update status will fail since ProvisioningStatus is not set
	assert.NotNil(t, err)

	err = provider.Delete(context.Background(), states.DeleteRequest{
		ID: "s234",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})
	assert.Nil(t, err)
}
