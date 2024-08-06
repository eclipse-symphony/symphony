/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"context"
	"fmt"
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

// Test config
var (
	NAMESPACE = "sample-k8s-scope"
)

// Verify target has correct status
func Test_VerifyTargetStatus(t *testing.T) {
	// Verify target
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "fabric.symphony",
			Version:  "v1",
			Resource: "targets",
		}).Namespace(NAMESPACE).List(context.Background(), metav1.ListOptions{})
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
	// Verify instance
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "solution.symphony",
			Version:  "v1",
			Resource: "instances",
		}).Namespace(NAMESPACE).List(context.Background(), metav1.ListOptions{})
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

func TestBasic_VerifyPodsExist(t *testing.T) {
	// Get kube client
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)

	i := 0
	for {
		i++

		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods(NAMESPACE).List(context.Background(), metav1.ListOptions{})
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