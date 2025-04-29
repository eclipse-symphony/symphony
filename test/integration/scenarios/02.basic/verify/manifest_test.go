/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	testManifests = []string{
		"../manifest/oss/solution-container.yaml",
		"../manifest/oss/target.yaml",
		"../manifest/oss/solution.yaml",
		"../manifest/oss/instance.yaml",
	}
)

func TestDryRunAndActiveState(t *testing.T) {
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	for _, manifest := range testManifests {
		_, err := DeployManifests(manifest, namespace, "true", "active")
		require.NoError(t, err)
	}
	testBasic_InstanceStatus(t, "0")
	testBasic_TargetStatus(t, "0")
	testBasic_VerifyPodsExist(t, []string{}, []string{"nginx", "testapp", namespace + "instance"})

	_, err := DeployManifests("../manifest/oss/target.yaml", namespace, "false", "active")
	require.NoError(t, err)
	testBasic_InstanceStatus(t, "0")
	testBasic_TargetStatus(t, "1")
	testBasic_VerifyPodsExist(t, []string{"nginx"}, []string{"testapp", namespace + "instance"})

	_, err = DeployManifests("../manifest/oss/instance.yaml", namespace, "false", "active")
	require.NoError(t, err)
	testBasic_InstanceStatus(t, "1")
	testBasic_TargetStatus(t, "1")
	testBasic_VerifyPodsExist(t, []string{"nginx", "testapp", namespace + "instance"}, []string{})

	output, err := DeployManifests("../manifest/oss/instance.yaml", namespace, "true", "active")
	require.Error(t, err)
	require.Contains(t, string(output), "The instance is already deployed. Cannot change isDryRun from false to true.")

	output, err = DeployManifests("../manifest/oss/target.yaml", namespace, "true", "active")
	require.Error(t, err)
	require.Contains(t, string(output), "The target is already deployed. Cannot change isDryRun from false to true.")

	_, err = DeployManifests("../manifest/oss/instance.yaml", namespace, "false", "inactive")
	require.NoError(t, err)
	testBasic_InstanceStatus(t, "0")
	testBasic_VerifyPodsExist(t, []string{"nginx"}, []string{"testapp", namespace + "instance"})

	_, err = DeployManifests("../manifest/oss/instance.yaml", namespace, "false", "active")
	require.NoError(t, err)
	testBasic_InstanceStatus(t, "1")
	testBasic_VerifyPodsExist(t, []string{"nginx", "testapp", namespace + "instance"}, []string{})

	err = CleanUpSymphonyObjects(namespace)
	if err != nil {
		t.Errorf("Failed to clean up Symphony objects Error: %v", err)
	}
}

// Verify target has correct status
func testBasic_TargetStatus(t *testing.T, successCount string) {
	// Verify targets
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "fabric.symphony",
		Version: "v1",
		Kind:    "Target",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		sleepDuration, _ := time.ParseDuration("10s")
		time.Sleep(sleepDuration)
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "fabric.symphony",
			Version:  "v1",
			Resource: "targets",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one target")

		targetState := getTargetState(resources.Items[0])
		fmt.Printf("Current target status: %v\n", targetState)
		require.NotEqual(t, "Failed", targetState.Status.ProvisioningStatus.Status, "target should not be in failed state")

		if success := targetState.Status.Deployed; targetState.Status.ProvisioningStatus.Status == "Succeeded" && strconv.FormatInt(int64(success), 10) == successCount {
			break
		}
	}
}

// Verify instance has correct status
func testBasic_InstanceStatus(t *testing.T, successCount string) {
	// Verify instances
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		sleepDuration, _ := time.ParseDuration("10s")
		time.Sleep(sleepDuration)
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "solution.symphony",
			Version:  "v1",
			Resource: "instances",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one instance")
		instance := getInstanceState(resources.Items[0])
		fmt.Printf("Current instance status: %v\n", instance)
		require.NotEqual(t, "Failed", instance.Status.ProvisioningStatus.Status, "instance should not be in failed state")
		// TODO: check success count
		if success := instance.Status.Deployed; instance.Status.ProvisioningStatus.Status == "Succeeded" && strconv.FormatInt(int64(success), 10) == successCount {
			break
		}
	}
}

// Verify that the pods we expect are running in the namespace
// Lists pods from the cluster and then verifies that the
// expected strings are found in the list.
func testBasic_VerifyPodsExist(t *testing.T, toFind []string, NotFound []string) {
	// Get kube client
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	i := 0
	for {
		i++
		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods(namespace+"scope").List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		notFound := make(map[string]bool)
		for _, s := range NotFound {
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, s) && pod.DeletionTimestamp == nil {
					require.Fail(t, "Pod found that should not be created", "Pod: %v", pod.Name)
				}
			}
		}
		for _, s := range toFind {
			found := false
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, s) && pod.Status.Phase == "Running" {
					found = true
					break
				}
			}

			if !found {
				notFound[s] = true
			}
		}

		if len(notFound) == 0 {
			fmt.Println("All pods found!")
			break
		} else {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods: %v\n", notFound)
			}
		}
	}
}

// Helper for finding the status
func getTargetState(resource unstructured.Unstructured) model.TargetState {
	data, err := json.Marshal(resource.Object)
	if err != nil {
		return model.TargetState{}
	}
	var instance model.TargetState
	err = json.Unmarshal(data, &instance)
	if err != nil {
		return model.TargetState{}
	}
	return instance
}

func getInstanceState(resource unstructured.Unstructured) model.InstanceState {
	data, err := json.Marshal(resource.Object)
	if err != nil {
		fmt.Printf("Failed to marshal resource: %v\n", resource)
		return model.InstanceState{}
	}
	var instance model.InstanceState
	err = json.Unmarshal(data, &instance)
	if err != nil {
		fmt.Printf("Failed to unmarshal resource: %v\n", resource)
		return model.InstanceState{}
	}
	return instance
}

func DeployManifests(fileName string, namespace string, dryrun string, activestate string) ([]byte, error) {
	if namespace != "default" {
		// Create non-default namespace if not exist
		output, err := exec.Command("kubectl", "get", "namespace", namespace).CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "not found") {
				// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
				output, err = exec.Command("kubectl", "create", "namespace", namespace).CombinedOutput()
				if err != nil {
					return output, err
				}
			} else {
				return output, err
			}
		}
	}

	fullPath, err := filepath.Abs(fileName)
	if err != nil {
		return []byte(fullPath), err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return data, err
	}
	stringYaml := string(data)
	if testhelpers.IsTestInAzure() {
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONREFNAME",
			"/subscriptions/af54d2ce-0dcb-48f8-9d2d-ff9c53e48c8d/resourcegroups/test-rg/providers/microsoft.edge/targets/TARGETNAME/solutions/SOLUTIONCONTAINERNAME/versions/SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "TARGETREFNAME", "/subscriptions/af54d2ce-0dcb-48f8-9d2d-ff9c53e48c8d/resourcegroups/test-rg/providers/microsoft.edge/targets/TARGETNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME-v-INSTANCENAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME-v-SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME")
	} else {
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONREFNAME", namespace+"solution:SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "TARGETREFNAME", "TARGETNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEFULLNAME", "INSTANCENAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONFULLNAME", "SOLUTIONCONTAINERNAME-v-SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERFULLNAME", "SOLUTIONCONTAINERNAME")
	}
	stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERNAME", namespace+"solution")
	stringYaml = strings.ReplaceAll(stringYaml, "INSTANCENAME", namespace+"instance")
	stringYaml = strings.ReplaceAll(stringYaml, "SCOPENAME", namespace+"scope")
	stringYaml = strings.ReplaceAll(stringYaml, "TARGETNAME", namespace+"target")
	stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONNAME", "version1")
	stringYaml = strings.ReplaceAll(stringYaml, "DRYRUN", dryrun)
	stringYaml = strings.ReplaceAll(stringYaml, "ACTIVESTATE", activestate)

	err = testhelpers.WriteYamlStringsToFile(stringYaml, "./test.yaml")
	if err != nil {
		return []byte{}, err
	}
	output, err := exec.Command("kubectl", "apply", "-f", "./test.yaml", "-n", namespace).CombinedOutput()
	// os.Remove("./test.yaml")
	if err != nil {
		return output, err
	}
	return []byte{}, nil
}

func CleanUpSymphonyObjects(namespace string) error {
	// Get all instances in the namespace
	output, err := shellcmd.Command(fmt.Sprintf("kubectl get instances.solution.symphony -n %s -o name", namespace)).Output()
	if err != nil {
		return fmt.Errorf("failed to list instances: %v", err)
	}

	// Split the output into individual instance names
	instances := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, instance := range instances {
		if instance == "" {
			continue
		}
		// Delete each instance
		err := shellcmd.Command(fmt.Sprintf("kubectl delete %s -n %s", instance, namespace)).Run()
		if err != nil {
			return fmt.Errorf("failed to delete instance %s: %v", instance, err)
		}
	}

	// Repeat similar logic for targets and solutions if needed
	output, err = shellcmd.Command(fmt.Sprintf("kubectl get targets.fabric.symphony -n %s -o name", namespace)).Output()
	if err != nil {
		return fmt.Errorf("failed to list targets: %v", err)
	}

	targets := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, target := range targets {
		if target == "" {
			continue
		}
		err := shellcmd.Command(fmt.Sprintf("kubectl delete %s -n %s", target, namespace)).Run()
		if err != nil {
			return fmt.Errorf("failed to delete target %s: %v", target, err)
		}
	}

	output, err = shellcmd.Command(fmt.Sprintf("kubectl get solutions.solution.symphony -n %s -o name", namespace)).Output()
	if err != nil {
		return fmt.Errorf("failed to list solutions: %v", err)
	}

	solutions := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, solution := range solutions {
		if solution == "" {
			continue
		}
		err := shellcmd.Command(fmt.Sprintf("kubectl delete %s -n %s", solution, namespace)).Run()
		if err != nil {
			return fmt.Errorf("failed to delete solution %s: %v", solution, err)
		}
	}
	return nil
}
