/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package verify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Verify target has correct status
func TestBasic_TargetStatus(t *testing.T) {
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
		}).Namespace("default").List(context.Background(), metav1.ListOptions{})
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
func TestBasic_InstanceStatus(t *testing.T) {
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
		}).Namespace("default").List(context.Background(), metav1.ListOptions{})
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

// Verify instance and namespace after deletion
func TestBasic_InstanceDeletion(t *testing.T) {
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// List all namespaces
	namespacesBefore, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	fmt.Println("Get namespace before deletion: ", len(namespacesBefore.Items))

	// Run a mage command to delete instance
	execCmd := exec.Command("sh", "-c", "cd ../../../../localenv && mage remove instances.solution.symphony instance03")
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	cmdErr := execCmd.Run()
	require.NoError(t, cmdErr)

	sleepDuration, _ := time.ParseDuration("10s")
	time.Sleep(sleepDuration)

	// Check instance count after deletion
	resources, err := dyn.Resource(schema.GroupVersionResource{
		Group:    "solution.symphony",
		Version:  "v1",
		Resource: "instances",
	}).Namespace("default").List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resources.Items, 0, "there should be no instance resource")

	// List all namespaces after deletion
	namespacesAfter, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	fmt.Println("Get namespace after deletion: ", len(namespacesAfter.Items))
	diff := len(namespacesBefore.Items) - len(namespacesAfter.Items)
	require.Equal(t, diff, 1, "there should be one namespace difference")
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
