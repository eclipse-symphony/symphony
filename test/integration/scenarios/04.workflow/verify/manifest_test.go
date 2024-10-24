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
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	CampaignNotExistActivation = "test/integration/scenarios/04.workflow/manifest/activation-campaignnotexist.yaml"

	WithStageActivation = "test/integration/scenarios/04.workflow/manifest/activation-stage.yaml"
)

// Verify catalog created
func TestBasic_Catalogs(t *testing.T) {
	fmt.Printf("Checking Catalogs\n")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "federation.symphony",
		Version: "v1",
		Kind:    "Catalog",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogs",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		catalogs := []string{}
		for _, item := range resources.Items {
			catalogs = append(catalogs, item.GetName())
		}
		fmt.Printf("Catalogs: %v\n", catalogs)
		if len(resources.Items) == 7 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify catalog created
func TestBasic_Campaign(t *testing.T) {
	fmt.Printf("Checking Campaign\n")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "workflow.symphony",
		Version: "v1",
		Kind:    "Campaign",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "workflow.symphony",
			Version:  "v1",
			Resource: "campaigns",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		if len(resources.Items) == 1 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

func TestBasic_ActivationStatus(t *testing.T) {
	fmt.Printf("Checking Activation\n")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "workflow.symphony",
		Version: "v1",
		Kind:    "Activation",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "workflow.symphony",
			Version:  "v1",
			Resource: "activations",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one activation")

		bytes, _ := json.Marshal(resources.Items[0].Object)
		var state model.ActivationState
		err = json.Unmarshal(bytes, &state)
		require.NoError(t, err)
		status := state.Status.Status
		fmt.Printf("Current activation status: %s\n", status)
		if status == v1alpha2.Done {
			require.Equal(t, 3, len(state.Status.StageHistory))
			require.Equal(t, "wait", state.Status.StageHistory[0].Stage)
			require.Equal(t, "list", state.Status.StageHistory[0].NextStage)
			require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
			require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[0].StatusMessage)
			require.Equal(t, "catalogs", state.Status.StageHistory[0].Inputs["objectType"])
			require.Equal(t, []interface{}{"sitecatalog:v1", "sitecatalog2:v1", "siteapp:v1", "sitek8starget:v1", "siteinstance:v1"}, state.Status.StageHistory[0].Inputs["names"].([]interface{}))
			require.Equal(t, "catalogs", state.Status.StageHistory[0].Outputs["objectType"])
			require.Equal(t, "list", state.Status.StageHistory[1].Stage)
			require.Equal(t, "deploy", state.Status.StageHistory[1].NextStage)
			require.Equal(t, v1alpha2.Done, state.Status.StageHistory[1].Status)
			require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[1].StatusMessage)
			require.Equal(t, "catalogs", state.Status.StageHistory[1].Inputs["objectType"])
			require.Equal(t, true, state.Status.StageHistory[1].Inputs["namesOnly"])
			require.Equal(t, []interface{}{"siteapp-v-v1", "sitecatalog-v-v1", "sitecatalog2-v-v1", "siteinstance-v-v1", "sitek8starget-v-v1"}, state.Status.StageHistory[1].Outputs["items"].([]interface{}))
			require.Equal(t, "catalogs", state.Status.StageHistory[1].Outputs["objectType"])
			require.Equal(t, "deploy", state.Status.StageHistory[2].Stage)
			require.Equal(t, "", state.Status.StageHistory[2].NextStage)
			require.Equal(t, v1alpha2.Done, state.Status.StageHistory[2].Status)
			require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[2].StatusMessage)
			require.Equal(t, []interface{}{"siteapp-v-v1", "sitecatalog-v-v1", "sitecatalog2-v-v1", "siteinstance-v-v1", "sitek8starget-v-v1"}, state.Status.StageHistory[2].Inputs["names"].([]interface{}))
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify target has correct status
func TestBasic_TargetStatus(t *testing.T) {
	fmt.Printf("Checking Target\n")
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
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "fabric.symphony",
			Version:  "v1",
			Resource: "targets",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one target")

		status := getStatus(resources.Items[0])
		fmt.Printf("Current target status: %s\n", status)
		require.NotEqual(t, "Failed", status, "target should not be in failed state")
		if status == "Succeeded" {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify target has correct labels
func TestAdvance_TargetLabel(t *testing.T) {
	fmt.Printf("Checking Target\n")
	namespace := os.Getenv("NAMESPACE")
	labelingEnabled := os.Getenv("labelingEnabled")
	if namespace == "" {
		namespace = "default"
	}
	expectedResult := "nolabel"
	if labelingEnabled == "true" {
		expectedResult = "localtest"
	}
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	resource, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "fabric.symphony",
		Version:  "v1",
		Resource: "targets",
	}).Namespace(namespace).Get(context.Background(), "sitek8starget", metav1.GetOptions{})
	require.NoError(t, err)

	result := getLabels(*resource)
	fmt.Printf("The target is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)
}

// Verify instance has correct status
func TestBasic_InstanceStatus(t *testing.T) {
	fmt.Printf("Checking Instances\n")
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "solution.symphony",
			Version:  "v1",
			Resource: "instances",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one instance")

		status := getStatus(resources.Items[0])
		fmt.Printf("Current instance status: %s\n", status)
		require.NotEqual(t, "Failed", status, "instance should not be in failed state")
		if status == "Succeeded" {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify instance has correct labels
func TestAdvance_InstanceLabel(t *testing.T) {
	fmt.Printf("Checking Target\n")
	namespace := os.Getenv("NAMESPACE")
	labelingEnabled := os.Getenv("labelingEnabled")
	if namespace == "" {
		namespace = "default"
	}
	expectedResult := "nolabel"
	if labelingEnabled == "true" {
		expectedResult = "localtest"
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	resource, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "instances",
	}).Namespace(namespace).Get(context.Background(), "siteinstance", metav1.GetOptions{})
	require.NoError(t, err)

	result := getLabels(*resource)
	fmt.Printf("The instance is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)
}

// Verify solution has correct labels
func TestAdvance_SolutionLabel(t *testing.T) {
	fmt.Printf("Checking Target\n")
	namespace := os.Getenv("NAMESPACE")
	labelingEnabled := os.Getenv("labelingEnabled")
	if namespace == "" {
		namespace = "default"
	}
	expectedResult := "nolabel"
	if labelingEnabled == "true" {
		expectedResult = "localtest"
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	resource, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "solutions",
	}).Namespace(namespace).Get(context.Background(), "siteapp-v-v1", metav1.GetOptions{})
	require.NoError(t, err)

	result := getLabels(*resource)
	fmt.Printf("The solution is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)

	annotations := getAnnotations(*resource)
	fmt.Printf("The instance is annotated with: %s\n", annotations)
	require.Equal(t, expectedResult, annotations)

	resource, err = dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "solutioncontainers",
	}).Namespace(namespace).Get(context.Background(), "siteapp", metav1.GetOptions{})
	require.NoError(t, err)

	result = getLabels(*resource)
	fmt.Printf("The solution container is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)
}

// Verify Catalog has correct labels
func TestAdvance_CatalogLabel(t *testing.T) {
	fmt.Printf("Checking Target\n")
	namespace := os.Getenv("NAMESPACE")
	labelingEnabled := os.Getenv("labelingEnabled")
	if namespace == "" {
		namespace = "default"
	}
	expectedResult := "nolabel"
	if labelingEnabled == "true" {
		expectedResult = "localtest"
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	resource, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "federation.symphony",
		Version:  "v1",
		Resource: "catalogs",
	}).Namespace(namespace).Get(context.Background(), "webappconfig-v-v1", metav1.GetOptions{})
	require.NoError(t, err)

	result := getLabels(*resource)
	fmt.Printf("The catalog is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)

	annotations := getAnnotations(*resource)
	fmt.Printf("The instance is annotated with: %s\n", annotations)
	require.Equal(t, expectedResult, annotations)

	resource, err = dyn.Resource(schema.GroupVersionResource{
		Group:    "federation.symphony",
		Version:  "v1",
		Resource: "catalogcontainers",
	}).Namespace(namespace).Get(context.Background(), "webappconfig", metav1.GetOptions{})
	require.NoError(t, err)

	result = getLabels(*resource)
	fmt.Printf("The catalog container is labeled with: %s\n", result)
	require.Equal(t, expectedResult, result)
}

// Verify that the pods we expect are running in the namespace
// Lists pods from the cluster and then verifies that the
// expected strings are found in the list.
func TestBasic_VerifyPodsExist(t *testing.T) {
	fmt.Printf("Checking Pod Status\n")
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
		pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		// Verify that the pods we expect are running
		toFind := []string{"web-app"}

		notFound := make(map[string]bool)
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
func getStatus(resource unstructured.Unstructured) string {
	status, ok := resource.Object["status"].(map[string]interface{})
	if ok {
		props, ok := status["provisioningStatus"].(map[string]interface{})
		if ok {
			statusString, ok := props["status"].(string)
			if ok {
				return statusString
			}
		}
	}

	return ""
}

// Helper for finding the labels
func getLabels(resource unstructured.Unstructured) string {
	labels := resource.GetLabels()
	if labels != nil {
		labelValue, ok := labels["localtest"]
		if ok {
			if labelValue == "localtest" {
				return labelValue
			} else {
				return "wronglabel"
			}
		} else {
			return "nolabel"
		}
	} else {
		return "nolabel"
	}
}

// Helper for finding the annotations
func getAnnotations(resource unstructured.Unstructured) string {
	annos := resource.GetAnnotations()
	name := resource.GetName()
	if annos != nil && name != "" {
		azureName, ok := annos["management.azure.com/azureName"]
		if ok {
			parts := strings.Split(name, "-v-")
			if azureName == parts[1] {
				return "localtest"
			} else {
				return "wrongAnnotationName"
			}
		} else {
			return "wrongAnnotationName"
		}
	} else {
		return "nolabel"
	}
}

func TestFaultScenario(t *testing.T) {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../../"
	}
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	CampaignNotExistActivationAbs := filepath.Join(repoPath, CampaignNotExistActivation)
	output, err := exec.Command("kubectl", "apply", "-f", CampaignNotExistActivationAbs, "-n", namespace).CombinedOutput()
	assert.Contains(t, string(output), "campaign reference must be a valid Campaign object in the same namespace")
	assert.NotNil(t, err, "fault test failed for non-existing campaign")
	WithStageActivationAbs := filepath.Join(repoPath, WithStageActivation)
	output, err = exec.Command("kubectl", "apply", "-f", WithStageActivationAbs, "-n", namespace).CombinedOutput()
	assert.Contains(t, string(output), "spec is immutable: stage doesn't match")
	assert.NotNil(t, err, "fault test failed for non-existing campaign")
}
