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
	"sync"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var (
	yamlFiles = map[string]string{
		"target":            "scenario2/target",
		"solution":          "scenario2/solution",
		"instance":          "scenario2/instance",
		"solutioncontainer": "scenario2/solution-container",
	}
	AzureIdFormat = map[string]string{
		"target":   "/subscriptions/aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e/resourcegroups/test-rg/providers/Microsoft.Edge/targets/scenario2targetINDEX",
		"solution": "/subscriptions/aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e/resourcegroups/test-rg/providers/Microsoft.Edge/targets/scenario2targetINDEX/solutions/scenario2solutioncontainerINDEX/versions/version1",
		"instance": "",
	}
	objectAzureName = map[string]string{
		"target":            "scenario2targetINDEX",
		"solution":          "scenario2targetINDEX-v-scenario2solutioncontainerINDEX-v-version1",
		"instance":          "scenario2targetINDEX-v-scenario2solutioncontainerINDEX-v-instanceINDEX",
		"solutioncontainer": "scenario2targetINDEX-v-scenario2solutioncontainerINDEX",
	}
)

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
	// Manifest templates
	numCRs          int // for prepare
	namespace       string
	basePath        string
	mapKindResource map[string]string
	cleanOnly       bool
)

func TestScenario_Stress_AllNamespaces(t *testing.T) {
	cleanOnly = false
	mapKindResource = map[string]string{
		"Activation":        "activations",
		"Campaign":          "campaigns",
		"CampaignContainer": "campaigncontainers",
		"Catalog":           "catalogs",
		"CatalogContainer":  "catalogcontainers",
		"SolutionContainer": "solutioncontainers",
		"Solution":          "solutions",
		"Instance":          "instances",
		"Target":            "targets",
	}
	numCRs = 1
	basePath = ".."
	namespace = os.Getenv("NAMESPACE")
	if namespace != "default" {
		// Create non-default namespace if not exist
		err := shellcmd.Command(fmt.Sprintf("kubectl get namespace %s", namespace)).Run()
		if err != nil {
			// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
			err = shellcmd.Command(fmt.Sprintf("kubectl create namespace %s", namespace)).Run()
			require.NoError(t, err)
		}
	}
	Scenario_Stress(t, namespace)
}

func Scenario_Stress(t *testing.T, namespace string) {
	log.SetLevel(log.InfoLevel)
	start := time.Now()

	config, err := testhelpers.RestConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}
	config.QPS = 15
	config.Burst = 15
	log.Infof("K8s config qps: %f, burst: %d", config.QPS, config.Burst)

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(numCRs)

	stopCh := make(chan int) // add 1 for watchscenario

	if !cleanOnly {
		go watchScenario2(dynamicClient, numCRs, stopCh)
	}

	for i := 0; i < numCRs; i++ {
		go func(i int) {
			defer wg.Done()
			if !cleanOnly {
				createScenario2(dynamicClient, i)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("stress test started trigger all deployments time: %s\n", elapsed)

	if !cleanOnly {
		<-stopCh
	}

	elapsed = time.Since(start)
	fmt.Printf("stress test execution time: %s\n", elapsed)
	wg.Add(numCRs)
	for i := 0; i < numCRs; i++ {
		go func(i int) {
			defer wg.Done()
			deleteScenario2(dynamicClient, i)
		}(i)
	}
	wg.Wait()
}

func watchScenario2(dynamicClient dynamic.Interface, nums int, wgTo chan int) {

	watcher, err := dynamicClient.Resource(getGVR("solution.symphony/v1", "Instance")).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	eventCount := 0

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			obj := event.Object.(*unstructured.Unstructured)
			// ss, _ := json.Marshal(obj)
			// log.Infof("custom resource modified " + string(ss))
			status, _, err := unstructured.NestedString(obj.Object, "status", "status")
			if err != nil {
				log.Errorf("%v", err)
			}

			name, found, err := unstructured.NestedString(obj.Object, "metadata", "name")
			if err != nil {
				log.Errorf("%v", err)
			}

			if found && status == "Succeeded" {
				log.Infof("custom resource %s, succeeded", name)
				eventCount++
			}
		case watch.Deleted:
			// Handle delete event if needed
		case watch.Error:
			// Handle error event if needed
		}
		if eventCount == nums {
			break
		}
	}
	log.Infof("ended")
	wgTo <- 1
}

func createScenario2(dynamicClient dynamic.Interface, index int) {
	createBasicContinerAndNested(dynamicClient, "solution", index, nil)
	createBasic(dynamicClient, "target", index, nil)
	adjust := func(spec map[interface{}]interface{}, index int) {
		if testhelpers.IsTestInAzure() {
			spec["solution"] = strings.ToLower(strings.ReplaceAll(AzureIdFormat["solution"], "INDEX", fmt.Sprintf("%d", index)))
			spec["target"].(map[interface{}]interface{})["name"] = strings.ToLower(strings.ReplaceAll(AzureIdFormat["target"], "INDEX", fmt.Sprintf("%d", index)))
		} else {
			spec["solution"] = fmt.Sprintf("scenario2solutioncontainer%d-v-version1", index)
			spec["target"].(map[interface{}]interface{})["name"] = fmt.Sprintf("scenario2target%d", index)
		}
	}
	createBasic(dynamicClient, "instance", index, adjust)
}

func deleteScenario2(dynamicClient dynamic.Interface, index int) {
	deleteBasic(dynamicClient, "instance", index)
	_, err := getBasic(dynamicClient, "instance", index)
	for err == nil || !errors.IsNotFound(err) {
		time.Sleep(2 * time.Second)
		_, err = getBasic(dynamicClient, "instance", index)
	}
	deleteBasic(dynamicClient, "target", index)
	deleteBasicContinerAndNested(dynamicClient, "solution", index)
}
func createBasic(dynamicClient dynamic.Interface, objectType string, index int, adjust func(map[interface{}]interface{}, int)) {
	var cr map[interface{}]interface{}
	crTemplate, err := os.ReadFile(fmt.Sprintf("%s/%s.yaml", basePath, yamlFiles[objectType]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Printf("Error unmarshalling custom resource template: %v", err)
		return
	}
	containerName := strings.Replace(yamlFiles[objectType], "/", "", -1)

	if testhelpers.IsTestInAzure() {
		cr["metadata"].(map[interface{}]interface{})["name"] = strings.ReplaceAll(objectAzureName[objectType], "INDEX", fmt.Sprintf("%d", index))
		if cr["metadata"].(map[interface{}]interface{})["annotations"] == nil {
			cr["metadata"].(map[interface{}]interface{})["annotations"] = map[interface{}]interface{}{}
		}
		cr["metadata"].(map[interface{}]interface{})["annotations"].(map[interface{}]interface{})["management.azure.com/resourceId"] = strings.ReplaceAll(AzureIdFormat[objectType], "INDEX", fmt.Sprintf("%d", index))
	} else {
		cr["metadata"].(map[interface{}]interface{})["name"] = fmt.Sprintf("%s%d", containerName, index)
	}

	if adjust != nil {
		adjust(cr["spec"].(map[interface{}]interface{}), index)
	}
	convertedCR := convertToUnstructured(cr)

	_, err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Create(context.TODO(), convertedCR, metav1.CreateOptions{})

	if err != nil {
		log.Errorf("Error creating custom resource, %s: %v", fmt.Sprintf("%s%d", containerName, index), err)
	} else {
		log.Debugf("Successfully created custom resource %s", fmt.Sprintf("%s%d", containerName, index))
	}
}

func getBasic(dynamicClient dynamic.Interface, objectType string, index int) (*unstructured.Unstructured, error) {
	var cr map[interface{}]interface{}
	crTemplate, err := os.ReadFile(fmt.Sprintf("%s/%s.yaml", basePath, yamlFiles[objectType]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Debugf("Error unmarshalling custom resource template: %v", err)
		return nil, err
	}
	objectName := strings.ReplaceAll(objectAzureName[objectType], "INDEX", fmt.Sprintf("%d", index))
	resource, err := dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Get(context.TODO(), objectName, metav1.GetOptions{})
	return resource, err
}

func deleteBasic(dynamicClient dynamic.Interface, objectType string, index int) {
	var cr map[interface{}]interface{}
	crTemplate, err := os.ReadFile(fmt.Sprintf("%s/%s.yaml", basePath, yamlFiles[objectType]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Debugf("Error unmarshalling custom resource template: %v", err)
		return
	}
	containerName := strings.Replace(yamlFiles[objectType], "/", "", -1)
	objectName := fmt.Sprintf("%s%d", containerName, index)
	if testhelpers.IsTestInAzure() {
		objectName = strings.ReplaceAll(objectAzureName[objectType], "INDEX", fmt.Sprintf("%d", index))
	}
	err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Delete(context.TODO(), objectName, metav1.DeleteOptions{})

	if err != nil {
		log.Warnf("Error deleting custom resource %s: %v", objectName, err)
	} else {
		log.Debugf("Successfully deleted custom resource %s", objectName)
	}
}

func createBasicContinerAndNested(dynamicClient dynamic.Interface, objectType string, index int, adjust func(map[interface{}]interface{}, int)) {
	var cr map[interface{}]interface{}
	crTemplate, err := os.ReadFile(fmt.Sprintf("%s/%s.yaml", basePath, yamlFiles[objectType+"container"]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Errorf("Error unmarshalling custom resource template: %v", err)
		return
	}
	containerName := strings.Replace(yamlFiles[objectType], "/", "", -1) + "container"
	if testhelpers.IsTestInAzure() {
		cr["metadata"].(map[interface{}]interface{})["name"] = strings.ReplaceAll(objectAzureName[objectType+"container"], "INDEX", fmt.Sprintf("%d", index))
	} else {
		cr["metadata"].(map[interface{}]interface{})["name"] = fmt.Sprintf("%s%d", containerName, index)
	}

	convertedCR := convertToUnstructured(cr)

	_, err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Create(context.TODO(), convertedCR, metav1.CreateOptions{})

	if err != nil {
		log.Errorf("Error creating custom resource, %s: %v", fmt.Sprintf("%s%d", containerName, index), err)
	} else {
		log.Debugf("Successfully created custom resource %s", fmt.Sprintf("%s%d", containerName, index))
	}

	crTemplate, err = os.ReadFile(fmt.Sprintf("%s/%s.yaml", basePath, yamlFiles[objectType]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Errorf("Error unmarshalling custom resource template: %v", err)
		return
	}

	if testhelpers.IsTestInAzure() {
		cr["metadata"].(map[interface{}]interface{})["name"] = strings.ReplaceAll(objectAzureName[objectType], "INDEX", fmt.Sprintf("%d", index))
		cr["spec"].(map[interface{}]interface{})["rootResource"] = strings.ReplaceAll(objectAzureName[objectType+"container"], "INDEX", fmt.Sprintf("%d", index))
		if cr["metadata"].(map[interface{}]interface{})["annotations"] == nil {
			cr["metadata"].(map[interface{}]interface{})["annotations"] = map[interface{}]interface{}{}
		}
		cr["metadata"].(map[interface{}]interface{})["annotations"].(map[interface{}]interface{})["management.azure.com/resourceId"] = strings.ReplaceAll(AzureIdFormat[objectType], "INDEX", fmt.Sprintf("%d", index))
	} else {
		cr["metadata"].(map[interface{}]interface{})["name"] = fmt.Sprintf("%s%d-v-version1", containerName, index)
		cr["spec"].(map[interface{}]interface{})["rootResource"] = fmt.Sprintf("%s%d", containerName, index)
	}
	if adjust != nil {
		adjust(cr["spec"].(map[interface{}]interface{}), index)
	}
	convertedCR = convertToUnstructured(cr)

	_, err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Create(context.TODO(), convertedCR, metav1.CreateOptions{})

	if err != nil {
		log.Warnf("Error creating custom resource, %s: %v", fmt.Sprintf("%s%d-v-version1", containerName, index), err)
	} else {
		log.Debugf("Successfully created custom resource %s", fmt.Sprintf("%s%d-v-version1", containerName, index))
	}
}

func deleteBasicContinerAndNested(dynamicClient dynamic.Interface, objectType string, index int) {
	var cr map[interface{}]interface{}
	crTemplate, err := os.ReadFile(fmt.Sprintf("%s/%s-container.yaml", basePath, yamlFiles[objectType]))
	if err != nil {
		log.Fatalf("Error reading custom resource template file: %v", err)
	}
	if err := yaml.Unmarshal(crTemplate, &cr); err != nil {
		log.Errorf("Error unmarshalling custom resource template: %v", err)
		return
	}
	objectName := fmt.Sprintf("%s%d-v-version1", strings.Replace(yamlFiles[objectType], "/", "", -1)+"container", index)
	if testhelpers.IsTestInAzure() {
		objectName = strings.ReplaceAll(objectAzureName[objectType], "INDEX", fmt.Sprintf("%d", index))
	}

	err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), strings.Replace(cr["kind"].(string), "Container", "", -1))).Namespace(namespace).Delete(context.TODO(), objectName, metav1.DeleteOptions{})
	if err != nil {
		log.Warnf("Error deleting custom resource %s,  %v", objectName, err)
	} else {
		log.Debugf("Successfully deleted custom resource %s", objectName)
	}
	containerName := strings.Replace(yamlFiles[objectType], "/", "", -1) + "container"
	if testhelpers.IsTestInAzure() {
		containerName = strings.ReplaceAll(objectAzureName[objectType+"container"], "INDEX", fmt.Sprintf("%d", index))
	}
	err = dynamicClient.Resource(getGVR(cr["apiVersion"].(string), cr["kind"].(string))).Namespace(namespace).Delete(context.TODO(), containerName, metav1.DeleteOptions{})

	if err != nil {
		log.Warnf("Error deleting custom resource %s, %d: %v", containerName, index, err)
	} else {
		log.Debugf("Successfully deleted custom resource %s, %d", containerName, index)
	}
}

func getGVR(apiVersion string, kind string) schema.GroupVersionResource {
	res1 := strings.Split(apiVersion, "/")
	return schema.GroupVersionResource{
		Group:    res1[0],
		Version:  res1[1],
		Resource: mapKindResource[kind],
	}
}

func convertToUnstructured(cr map[interface{}]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: convertToStringMap(cr),
	}
}

func convertToStringMap(in map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for key, value := range in {
		strKey := fmt.Sprintf("%v", key)
		switch value := value.(type) {
		case map[interface{}]interface{}:
			out[strKey] = convertToStringMap(value)
		case []interface{}:
			out[strKey] = convertToStringSlice(value)
		default:
			out[strKey] = value
		}
	}
	return out
}

func convertToStringSlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, value := range in {
		switch value := value.(type) {
		case map[interface{}]interface{}:
			out[i] = convertToStringMap(value)
		case []interface{}:
			out[i] = convertToStringSlice(value)
		default:
			out[i] = value
		}
	}
	return out
}
