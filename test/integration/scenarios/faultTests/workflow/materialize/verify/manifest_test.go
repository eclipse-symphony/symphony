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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/eclipse-symphony/symphony/test/integration/scenarios/faultTests/utils"
	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	testCatalogVersions = []string{
		"./manifest/catalog-catalog-container.yaml",
		"./manifest/catalog-catalog-container-2.yaml",
		"./manifest/instance-catalogversion-container.yaml",
		"./manifest/solutionversion-catalogversion-container.yaml",
		"./manifest/target-catalogversion-container.yaml",

		"./manifest/catalogversion-catalog.yaml",
		"./manifest/catalogversion-catalog-2.yaml",
		"./manifest/instance-catalogversion.yaml",
		"./manifest/solutionversion-catalogversion.yaml",
		"./manifest/target-catalogversion.yaml",
	}

	testCampaignVersion = []string{
		"./manifest/campaign.yaml",
		"./manifest/campaignversion.yaml",
	}

	testActivations = []string{
		"./manifest/activation.yaml",
	}
)

func TestMaterializeWorkflow(t *testing.T) {
	namespace := "nondefault"
	err := utils.InjectPodFailure()
	require.NoError(t, err)
	DeployManifests(t, namespace)
	CheckCatalogVersions(t, namespace)
	CheckActivationStatus(t, namespace)
	CheckTargetStatus(t, namespace)
	CheckInstanceStatus(t, namespace)
	VerifyPodsExist(t, namespace)
}

func DeployManifests(t *testing.T, namespace string) error {
	repoPath := "../"
	if namespace != "default" {
		// Create non-default namespace if not exist
		err := shellcmd.Command(fmt.Sprintf("kubectl get namespace %s", namespace)).Run()
		if err != nil {
			// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
			err = shellcmd.Command(fmt.Sprintf("kubectl create namespace %s", namespace)).Run()
			if err != nil {
				return err
			}
		}
	}
	// Deploy the catalogversions
	for _, catalogversion := range testCatalogVersions {
		absCatalogVersion := filepath.Join(repoPath, catalogversion)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absCatalogVersion, namespace)).Run()
		if err != nil {
			return err
		}
	}

	for _, campaignversion := range testCampaignVersion {
		absCampaignVersion := filepath.Join(repoPath, campaignversion)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absCampaignVersion, namespace)).Run()
		if err != nil {
			return err
		}
	}

	CheckCampaignVersion(t, namespace)

	for _, activation := range testActivations {
		absActivation := filepath.Join(repoPath, activation)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absActivation, namespace)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Verify catalogversion created
func CheckCatalogVersions(t *testing.T, namespace string) {
	fmt.Printf("Checking CatalogVersions\n")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "federation.symphony",
		Version: "v1",
		Kind:    "CatalogVersion",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogversions",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		catalogversions := []string{}
		for _, item := range resources.Items {
			catalogversions = append(catalogversions, item.GetName())
		}
		fmt.Printf("CatalogVersions: %v\n", catalogversions)
		if len(resources.Items) == 7 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify campaignversion created
func CheckCampaignVersion(t *testing.T, namespace string) {
	fmt.Printf("Checking CampaignVersion\n")
	if namespace == "" {
		namespace = "default"
	}
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "workflow.symphony",
		Version: "v1",
		Kind:    "CampaignVersion",
	})

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "workflow.symphony",
			Version:  "v1",
			Resource: "campaignversions",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		if len(resources.Items) == 1 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

func CheckActivationStatus(t *testing.T, namespace string) {
	fmt.Printf("Checking Activation\n")
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
			// Skip checking the stageHistory since we don't have stage dedup
			// require.Equal(t, 3, len(state.Status.StageHistory))
			// require.Equal(t, "wait", state.Status.StageHistory[0].Stage)
			// require.Equal(t, "list", state.Status.StageHistory[0].NextStage)
			// require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
			// require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[0].StatusMessage)
			// require.Equal(t, "catalogversions", state.Status.StageHistory[0].Inputs["objectType"])
			// require.Equal(t, []interface{}{"sitecatalog:version1", "sitecatalog2:version1", "siteapp:version1", "sitek8starget:version1", "siteinstance:version1"}, state.Status.StageHistory[0].Inputs["names"].([]interface{}))
			// require.Equal(t, "catalogversions", state.Status.StageHistory[0].Outputs["objectType"])
			// require.Equal(t, "list", state.Status.StageHistory[1].Stage)
			// require.Equal(t, "deploy", state.Status.StageHistory[1].NextStage)
			// require.Equal(t, v1alpha2.Done, state.Status.StageHistory[1].Status)
			// require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[1].StatusMessage)
			// require.Equal(t, "catalogversions", state.Status.StageHistory[1].Inputs["objectType"])
			// require.Equal(t, true, state.Status.StageHistory[1].Inputs["namesOnly"])
			// require.Equal(t, []interface{}{"siteapp-v-version1", "sitecatalog-v-version1", "sitecatalog2-v-version1", "siteinstance-v-version1", "sitek8starget-v-version1"}, state.Status.StageHistory[1].Outputs["items"].([]interface{}))
			// require.Equal(t, "catalogversions", state.Status.StageHistory[1].Outputs["objectType"])
			// require.Equal(t, "deploy", state.Status.StageHistory[2].Stage)
			// require.Equal(t, "", state.Status.StageHistory[2].NextStage)
			// require.Equal(t, v1alpha2.Done, state.Status.StageHistory[2].Status)
			// require.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[2].StatusMessage)
			// require.Equal(t, []interface{}{"siteapp-v-version1", "sitecatalog-v-version1", "sitecatalog2-v-version1", "siteinstance-v-version1", "sitek8starget-v-version1"}, state.Status.StageHistory[2].Inputs["names"].([]interface{}))
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify target has correct status
func CheckTargetStatus(t *testing.T, namespace string) {
	fmt.Printf("Checking Target\n")
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

// Verify instance has correct status
func CheckInstanceStatus(t *testing.T, namespace string) {
	fmt.Printf("Checking Instances\n")
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

// Verify that the pods we expect are running in the namespace
// Lists pods from the cluster and then verifies that the
// expected strings are found in the list.
func VerifyPodsExist(t *testing.T, namespace string) {
	fmt.Printf("Checking Pod Status\n")
	kubeClient, err := testhelpers.KubeClient()
	require.NoError(t, err)

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
