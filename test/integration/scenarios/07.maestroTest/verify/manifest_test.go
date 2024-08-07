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
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Verify target has correct status
func Test_VerifyTargetStatus(t *testing.T) {
	namespace := getNamespace()

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

// Verify instance has correct status
func Test_VerifyInstanceStatus(t *testing.T) {
	namespace := getNamespace()

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

// Verify pod exists in hello-world sample
func TestBasic_VerifyPodsExist(t *testing.T) {
	sampleName := getNamespace()

	if (sampleName == "sample-hello-world") {
		// Get kube client
		kubeClient, err := testhelpers.KubeClient()
		require.NoError(t, err)

		i := 0
		for {
			i++

			// List all pods in the deployed namespace
			pods, err := kubeClient.CoreV1().Pods("sample-k8s-scope").List(context.Background(), metav1.ListOptions{})
			require.NoError(t, err)

			found := false
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, "sample-prometheus-instance") && pod.Status.Phase == "Running" {
					found = true
					break
				}
			}

			if found {
				break
			} else {
				time.Sleep(time.Second * 5)

				if i%12 == 0 {
					fmt.Printf("Waiting for pods: sample-prometheus-instance\n")
				}
			}
		}
	}
}

// Verify catelog exists in staged sample
func TestBasic_VerifyCatelogExist(t *testing.T) {
	sampleName := getNamespace()

	if (sampleName == "sample-staged") {
		cfg, err := testhelpers.RestConfig()
		require.NoError(t, err)
	
		dyn, err := dynamic.NewForConfig(cfg)
		require.NoError(t, err)
	
		for {
			resources, err := dyn.Resource(schema.GroupVersionResource{
				Group:    "federation.symphony",
				Version:  "v1",
				Resource: "catalogs",
			}).Namespace("default").List(context.TODO(), metav1.ListOptions{})
			require.NoError(t, err)

			found := false
			for _, catalog := range resources.Items {
				if strings.Contains(catalog.GetName(), "sample-staged-instance") {
					found = true
					break
				}
			}

			if found {
				break
			} else {
				sleepDuration, _ := time.ParseDuration("30s")
				time.Sleep(sleepDuration)
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

// Helper for getting namespace
func getNamespace() string {
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	return namespace
}