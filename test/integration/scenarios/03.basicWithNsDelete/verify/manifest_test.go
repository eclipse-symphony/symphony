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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// Verify that circular reference is detected
func TestBasic_DetectCircularReference(t *testing.T) {
	// Verify instance status failed due to circular reference
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
		if status == "Failed" {
			message := getErrorMessage(resources.Items[0])
			require.Equal(t, "Deployment failed. failed to evaluate deployment spec: Bad Config: Detect circular dependency, object: config1-v-v1, field: image", message)
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}

	// Update to correct config
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "fabric.symphony",
		Version: "v1",
		Kind:    "Target",
	})

	namespace := "default"
	// read catalog
	catalog, err := readCatalog("config1-v-v1", namespace, dyn)
	require.NoError(t, err)

	// Update catalog
	catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["image"] = "prom/prometheus"
	_, err = updateCatalog(namespace, catalog, dyn)
	require.NoError(t, err)

	// Deploy the updated solution manifest
	manifest := "../manifest/oss/solution-new.yaml"
	fullPath, err := filepath.Abs(manifest)
	require.NoError(t, err)

	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
	require.NoError(t, err)
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
		if status == "Succeeded" {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

func TestBasic_VerifyPod(t *testing.T) {
	// Get kube client
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)
	namespace := "k8s-scope"

	i := 0
	for {
		i++
		pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		podToFind := "instance03"
		found := false
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, podToFind) && pod.Status.Phase == "Running" {
				found = true

				for _, container := range pod.Spec.Containers {
					requests := container.Resources.Requests
					cpuRequest := requests["cpu"]
					memRequest := requests["memory"]
					fmt.Printf("Container: %s, CPU request: %s, memory request: %s\n", container.Name, cpuRequest.String(), memRequest.String())
					require.Equal(t, "100m", cpuRequest.String(), "CPU should be 100 milliCPU.")
					require.Equal(t, "100Mi", memRequest.String(), "Memory should be 100 Mebibytes")

					for _, port := range container.Ports {
						fmt.Printf("Container: %s, Port: %d\n", container.Name, port.ContainerPort)
						require.Equal(t, int32(9090), port.ContainerPort, "instance03", "container port should be 9090")
						break
					}
				}
				break
			}
		}

		if !found {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods: %v\n", podToFind)
			}
		} else {
			fmt.Println("Pod is found.")
			break
		}
	}
}

func TestBasic_VerifyPodUpdatedInNamespace(t *testing.T) {
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)
	namespace := "k8s-scope"

	// Get old pod name
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)

	podToFind := "instance03"
	var podNameBefore string
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, podToFind) && pod.Status.Phase == "Running" {
			podNameBefore = pod.Name
			break
		}
	}

	// Deploy the updated solution manifest
	manifest := "../manifest/oss/solution-update.yaml"
	fullPath, err := filepath.Abs(manifest)
	require.NoError(t, err)

	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
	require.NoError(t, err)

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)
	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	// Verify instance status
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

	// Verify pod status
	i := 0
	for {
		i++
		pods, err := kubeClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		podToFind := "instance03"
		found := false
		for _, pod := range pods.Items {
			fmt.Printf("Pod name: %s\n", pod.Name)
			if strings.Contains(pod.Name, podToFind) && pod.Status.Phase == "Running" && pod.Name != podNameBefore {
				found = true

				for _, container := range pod.Spec.Containers {
					requests := container.Resources.Requests
					cpuRequest := requests["cpu"]
					memRequest := requests["memory"]
					fmt.Printf("Container: %s, CPU request: %s, memory request: %s\n", container.Name, cpuRequest.String(), memRequest.String())
					require.Equal(t, "500m", cpuRequest.String(), "CPU should be 500 milliCPU.")
					require.Equal(t, "500Mi", memRequest.String(), "Memory should be 500 Mebibytes")

					for _, port := range container.Ports {
						fmt.Printf("Container: %s, Port: %d\n", container.Name, port.ContainerPort)
						require.Equal(t, int32(9900), port.ContainerPort, "instance03", "container port should be 9900")
					}
				}
				break
			}
		}

		if !found {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods: %v\n", podToFind)
			}
		} else {
			fmt.Println("Pod is found.")
			break
		}
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

func TestBasic_VerifySameInstanceRecreationInNamespace(t *testing.T) {
	// Manifests to deploy
	var testManifests = []string{
		"../manifest/oss/solution2.yaml",
		"../manifest/oss/target2.yaml",
		"../manifest/oss/instance-recreate.yaml",
	}

	// Deploy the manifests
	for _, manifest := range testManifests {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)

		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
		require.NoError(t, err)
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)
	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	// Verify new instance status
	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "solution.symphony",
			Version:  "v1",
			Resource: "instances",
		}).Namespace("default").List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		require.Len(t, resources.Items, 1, "there should be only one instance")

		status := getStatus(resources.Items[0])
		deployableStatus, err := getDeployableStatus(resources.Items[0])
		require.NoError(t, err)
		targetCount := deployableStatus.Targets
		target03Status := deployableStatus.GetTargetStatus("target03")
		helmTargetStatus := deployableStatus.GetTargetStatus("helm-target")

		fmt.Printf("Current instance status: %s\n", status)
		fmt.Printf("Current instance deployment count: %d\n", targetCount)
		fmt.Printf("Current instance deployment instance3: %s\n", target03Status)
		fmt.Printf("Current instance deployment helm: %s\n", helmTargetStatus)

		require.NotEqual(t, "Failed", status, "instance should not be in failed state")
		require.NotContains(t, target03Status, "OK", "instance should not show target03 status")
		if status == "Succeeded" && targetCount == 1 && target03Status == "" && strings.Contains(helmTargetStatus, "OK") {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

func TestBasic_VerifyTargetSolutionScope(t *testing.T) {
	// Manifests to deploy
	var testDefaultManifests = []string{
		"../manifest/oss/solution-configmap.yaml",
		"../manifest/oss/target-configmap-default.yaml",
		"../manifest/oss/instance-configmap-default.yaml",
	}

	// Deploy the manifests in default namespace
	for _, manifest := range testDefaultManifests {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)

		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
		require.NoError(t, err)
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)
	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// Verify configmap in default scope
	for {
		namespace := "default"
		configMapName := "configmap"
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
		if err == nil {
			require.Equal(t, "test-tag", configMap.Data["tags"], "configmap data should match the input")
			break
		}

		sleepDuration, _ := time.ParseDuration("1s")
		time.Sleep(sleepDuration)
	}

	// test update target solutionScope, expect error
	targetFile := "../manifest/oss/target-configmap-error.yaml"
	fullPath, err := filepath.Abs(targetFile)
	require.NoError(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", fullPath).CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(output), "Target has one or more associated instances. Cannot change SolutionScope of the target.")

	// test update instance scope, expect error
	instanceFile := "../manifest/oss/instance-configmap-error.yaml"
	fullPath, err = filepath.Abs(instanceFile)
	require.NoError(t, err)
	output, err = exec.Command("kubectl", "apply", "-f", fullPath).CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(output), "The instance is already created. Cannot change Scope of the instance.")

	// delete instance and associated deployments
	err = shellcmd.Command(fmt.Sprintf("kubectl delete instance.solution.symphony %s", "instance-configmap")).Run()
	require.NoError(t, err)

	// test update target solutionScope with no associated instance, expect no error
	fullPath, _ = filepath.Abs(targetFile)
	output, err = exec.Command("kubectl", "apply", "-f", fullPath).CombinedOutput()
	require.NoError(t, err)

	// deploy new instance and target with solutionScope
	var testSolutionScopeManifests = []string{
		"../manifest/oss/target-configmap-update.yaml",
		"../manifest/oss/instance-configmap-update.yaml",
	}

	// Deploy the manifests in default namespace
	for _, manifest := range testSolutionScopeManifests {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)

		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
		require.NoError(t, err)
	}

	// Verify configmp in target solutionScope
	for {
		namespace := "target-scope"
		configMapName := "configmap"
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
		if err == nil {
			require.Equal(t, "test-tag", configMap.Data["tags"], "configmap data should match the input")
			break
		}

		sleepDuration, _ := time.ParseDuration("1s")
		time.Sleep(sleepDuration)
	}
}

func TestBasic_VerifySolutionScopePrecedence(t *testing.T) {
	// Create instance with scope and target with solution scope
	var testSolutionScopeManifests = []string{
		"../manifest/oss/target-configmap-update.yaml",
		"../manifest/oss/instance-configmap-with-scope.yaml",
	}

	// Deploy the manifests in default namespace
	for _, manifest := range testSolutionScopeManifests {
		fullPath, err := filepath.Abs(manifest)
		require.NoError(t, err)

		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n default", fullPath)).Run()
		require.NoError(t, err)
	}

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)
	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// Verify configmap in nondefault instance scope
	for {
		namespace := "nondefault"
		configMapName := "configmap"
		configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
		if err == nil {
			require.Equal(t, "test-tag", configMap.Data["tags"], "configmap data should match the input")
			break
		}

		sleepDuration, _ := time.ParseDuration("1s")
		time.Sleep(sleepDuration)
	}
}

// Helper for read catalog
func readCatalog(catalogName string, namespace string, dynamicClient dynamic.Interface) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogs"}
	catalog, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), catalogName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

// Helper for update catalog
func updateCatalog(namespace string, object *unstructured.Unstructured, dynamicClient dynamic.Interface) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogs"}
	catalog, err := dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), object, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return catalog, nil
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

func getDeployableStatus(resource unstructured.Unstructured) (model.DeployableStatusV2, error) {
	status, ok := resource.Object["status"].(map[string]interface{})
	if ok {
		statusJson, err := json.Marshal(status)
		if err != nil {
			return model.DeployableStatusV2{}, err
		}
		var deployableStatus model.DeployableStatusV2
		err = json.Unmarshal(statusJson, &deployableStatus)
		if err != nil {
			return model.DeployableStatusV2{}, err
		}
		return deployableStatus, nil
	}
	return model.DeployableStatusV2{}, nil
}

func getProperty(resource unstructured.Unstructured, propertyName string) string {
	status, ok := resource.Object["status"].(map[string]interface{})
	if ok {
		props, ok := status["properties"].(map[string]interface{})
		if ok {
			property, ok := props[propertyName].(string)
			if ok {
				return property
			}
		}
	}

	return ""
}

// Helper for finding the status
func getErrorMessage(resource unstructured.Unstructured) string {
	status, ok := resource.Object["status"].(map[string]interface{})
	if ok {
		props, ok := status["provisioningStatus"].(map[string]interface{})
		if ok {
			statusString, ok := props["error"].(map[string]interface{})
			if ok {
				message, ok := statusString["message"].(string)
				if ok {
					return message
				}
			}
		}
	}

	return ""
}
