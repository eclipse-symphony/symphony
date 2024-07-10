/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const ()

type (
	TestCase struct {
		// Name gives the brief introduction of each test case
		Name string

		// Target is Symphony manifest to test, e.g. solution/target
		Target string

		// ComponentsToAdd specifies the components to be added to the symphony manifest
		ComponentsToAdd []string

		// PodsToVerify specifies the pods need to be running
		PodsToVerify []string

		// DeletedPodsToVerify specifies the pods need to be deleted
		DeletedPodsToVerify []string
	}
)

var (
	// manifestTemplateFolder includes manifest templates with empty components to deploy
	manifestTemplateFolder = "../manifestTemplates"
	// testManifestsFolder includes temporary manifest files for each test run. set in .gitignore
	testManifestsFolder = "../manifestForTestingOnly"
)

var (
	// Manifest templates
	containerManifestTemplates = map[string]string{
		"solution-container": fmt.Sprintf("%s/%s/solution-container.yaml", manifestTemplateFolder, "oss"),
	}

	manifestTemplates = map[string]string{
		"target":   fmt.Sprintf("%s/%s/target.yaml", manifestTemplateFolder, "oss"),
		"instance": fmt.Sprintf("%s/%s/instance.yaml", manifestTemplateFolder, "oss"),
		"solution": fmt.Sprintf("%s/%s/solution.yaml", manifestTemplateFolder, "oss"),
	}

	// Manifests to deploy
	testManifests = map[string]string{
		"target":   fmt.Sprintf("%s/%s/target.yaml", testManifestsFolder, "oss"),
		"instance": fmt.Sprintf("%s/%s/instance.yaml", testManifestsFolder, "oss"),
		"solution": fmt.Sprintf("%s/%s/solution.yaml", testManifestsFolder, "oss"),
	}

	testCases = []TestCase{
		{
			Name:                "Initial Symphony Target Deployment",
			Target:              "target",
			ComponentsToAdd:     []string{"e4k", "e4k-broker"},
			PodsToVerify:        []string{"azedge-dmqtt-backend", "azedge-dmqtt-frontend"},
			DeletedPodsToVerify: []string{},
		},
		{
			Name:                "Update Symphony Target to add bluefin-extension",
			Target:              "target",
			ComponentsToAdd:     []string{"e4k", "e4k-broker", "bluefin-extension"},
			PodsToVerify:        []string{"azedge-dmqtt-backend", "azedge-dmqtt-frontend", "bluefin-operator-controller"},
			DeletedPodsToVerify: []string{},
		},
		{
			Name:                "Update Symphony Solution to add bluefin-instance and bluefin-pipeline",
			Target:              "solution",
			ComponentsToAdd:     []string{"bluefin-instance", "bluefin-pipeline"},
			PodsToVerify:        []string{"azedge-dmqtt-backend", "azedge-dmqtt-frontend", "bluefin-operator-controller", "bluefin-scheduler-0", "bluefin-runner-worker-0"},
			DeletedPodsToVerify: []string{},
		},
		{
			Name:                "Update Symphony Solution to remove bluefin-instance and bluefin-pipeline",
			Target:              "solution",
			ComponentsToAdd:     []string{},
			PodsToVerify:        []string{"azedge-dmqtt-backend", "azedge-dmqtt-frontend", "bluefin-operator-controller"},
			DeletedPodsToVerify: []string{"bluefin-scheduler-0", "bluefin-runner-worker-0"},
		},
		{
			Name:                "Update Symphony Target to remove bluefin-extension and e4k",
			Target:              "target",
			ComponentsToAdd:     []string{},
			PodsToVerify:        []string{},
			DeletedPodsToVerify: []string{"azedge-dmqtt-backend", "azedge-dmqtt-frontend", "bluefin-operator-controller"},
		},
	}
)

func TestScenario_Update_AllNamespaces(t *testing.T) {
	namespace := os.Getenv("NAMESPACE")
	if namespace != "default" {
		// Create non-default namespace if not exist
		err := shellcmd.Command(fmt.Sprintf("kubectl get namespace %s", namespace)).Run()
		if err != nil {
			// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
			err = shellcmd.Command(fmt.Sprintf("kubectl create namespace %s", namespace)).Run()
			require.NoError(t, err)
		}
	}
	Scenario_Update(t, namespace)
}

func Scenario_Update(t *testing.T, namespace string) {
	// Deploy base manifests
	for _, manifest := range containerManifestTemplates {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", fullPath, namespace)).Run()
		require.NoError(t, err)
	}
	for _, manifest := range manifestTemplates {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", fullPath, namespace)).Run()
		require.NoError(t, err)
	}
	for _, test := range testCases {
		fmt.Printf("[Test case]: %s\n", test.Name)

		// Construct the manifests
		err := testhelpers.BuildManifestFile(
			fmt.Sprintf("%s/%s", manifestTemplateFolder, "oss"),
			fmt.Sprintf("%s/%s", testManifestsFolder, "oss"), test.Target, test.ComponentsToAdd)
		require.NoError(t, err)

		// Deploy the modified manifests
		for _, manifest := range testManifests {
			fullPath, err := filepath.Abs(manifest)
			require.NoError(t, err)
			// skip deploying unchanged manifest to test instance Watch logic
			// i.e. target and solution changes should trigger instance reconciler
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				continue
				// fullPath, err = filepath.Abs(manifestTemplates[k])
				// require.NoError(t, err)
			}
			err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", fullPath, namespace)).Run()
			require.NoError(t, err)
		}

		verifyTargetStatus(t, test, namespace)
		verifyInstanceStatus(t, test, namespace)
		verifyPodsExist(t, test, test.PodsToVerify)
		verifyPodsDeleted(t, test, test.DeletedPodsToVerify)
	}
}

// Verify target has correct status
func verifyTargetStatus(t *testing.T, test TestCase, namespace string) {
	// Verify targets
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
		require.NotEqual(t, "Failed", status, fmt.Sprintf("%s: Target should not be in failed state", test.Name))
		if status == "Succeeded" {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify instance has correct status
func verifyInstanceStatus(t *testing.T, test TestCase, namespace string) {
	// Verify instances
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
		require.NotEqual(t, "Failed", status, fmt.Sprintf("%s: Instance should not be in failed state", test.Name))
		if status == "Succeeded" {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify that the pods we expect are running in the namespace
// Lists pods from the cluster and then verifies that the
// expected strings are found in the list.
func verifyPodsExist(t *testing.T, test TestCase, toFind []string) {
	// Get kube client
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)

	i := 0
	for {
		i++
		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods("alice-springs").List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		// Verify that the pods we expect are running
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

// Verify that the pods we expect are deleted in the namespace
// Lists pods from the cluster and then verifies that the
// expected strings are no longer found in the list.
func verifyPodsDeleted(t *testing.T, test TestCase, toFind []string) {
	// Get kube client
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)

	i := 0
	for {
		i++
		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods("alice-springs").List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		// Verify that the pods we expect are deleted
		waitingForDeletion := make(map[string]bool)
		for _, s := range toFind {
			found := false
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, s) {
					found = true
					break
				}
			}

			if found {
				waitingForDeletion[s] = true
			}
		}

		if len(waitingForDeletion) == 0 {
			fmt.Println("All pods deleted!")
			break
		} else {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods to be deleted: %v\n", waitingForDeletion)
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
