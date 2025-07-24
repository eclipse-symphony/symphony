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

	k8sstate "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/states/k8s"
)

func TestInit(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	symphonyDevice := os.Getenv("SYMPHONY_DEVICE")
	if testK8s == "" || symphonyDevice == "" {
		t.Skip("Skipping because TEST_K8S or SYMPHONY_DEVICE enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
	err = provider.Report(symphonyDevice, "default", "fabric.symphony", "devices", "v1", map[string]string{
		"a": "ccc",
		"b": "ddd",
	}, false)
	assert.Nil(t, err)
}

func TestReportTargetProperty(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.InitWithMap(map[string]string{
		"name":      "test",
		"inCluster": "false",
	})
	assert.Nil(t, err)
	err = checkTargetCRDApplied()
	assert.Nil(t, err)

	k8sStateProvider := k8sstate.K8sStateProvider{}
	err = k8sStateProvider.Init(k8sstate.K8sStateProviderConfig{
		InCluster:  false,
		ConfigType: "path",
	})
	assert.Nil(t, err)

	id, err := k8sStateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "a1",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": "a1",
				},
				"spec": model.TargetSpec{},
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

	time.Sleep(5 * time.Second)
	err = provider.Report(id, "default", "fabric.symphony", "targets", "v1", map[string]string{
		"testkey": "testval",
		"status":  "Succeeded",
	}, true)
	assert.Nil(t, err)

	entry, err := k8sStateProvider.Get(context.Background(), states.GetRequest{
		ID: "a1",
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
		},
	})
	assert.Nil(t, err)
	status := entry.Body.(map[string]interface{})["status"]
	j, _ := json.Marshal(status)
	var rStatus model.TargetStatus
	err = json.Unmarshal(j, &rStatus)
	assert.Nil(t, err)
	assert.Equal(t, "Succeeded", rStatus.Status)
	assert.Equal(t, "testval", rStatus.Properties["testkey"])
}

func runKubectl(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	return cmd.Output()
}

func checkTargetCRDApplied() error {
	// Check that the CRD is applied
	_, err := runKubectl("get", "crd", "targets.fabric.symphony")
	if err != nil {
		// apply the CRD api/pkg/apis/v1alpha1/providers/states/k8s/k8s_test.go
		ProjectPath := os.Getenv("REPOPATH")
		targetYamlPath := ProjectPath + "/k8s/config/oss/crd/bases/fabric.symphony_targets.yaml"
		if _, err := os.Stat(targetYamlPath); err != nil {
			return err
		}
		_, err = runKubectl("apply", "-f", targetYamlPath)
		if err != nil {
			return err
		}
		// Wait for the CRD to be applied
		time.Sleep(10 * time.Second)
	}
	return nil
}
