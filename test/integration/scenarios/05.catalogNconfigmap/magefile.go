//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Test config
const (
	TEST_NAME    = "Symphony CatalogVersion CRUD test scenario"
	TEST_TIMEOUT = "10m"
)

var (
	// catalogversions to deploy
	testCatalogVersions = []string{
		"manifests/instance-container.yaml",
		"manifests/solutionversion-container.yaml",
		"manifests/target-container.yaml",
		"manifests/asset-container.yaml",
		"manifests/config-container.yaml",
		"manifests/wrongconfig-container.yaml",
		"manifests/schema-container.yaml",

		"manifests/instance.yaml",
		"manifests/solutionversion.yaml",
		"manifests/target.yaml",
		"manifests/asset.yaml",
	}

	// catalogversions for namespace test
	testNamespace = []string{
		"namespace/config1-container.yaml",
		"namespace/config1.yaml",
		"namespace/config2-container.yaml",
		"namespace/config2.yaml",
		"namespace/config3-container.yaml",
		"namespace/config3.yaml",
		"namespace/campaign-container.yaml",
		"namespace/campaign.yaml",
	}

	testActivation = "namespace/activation.yaml"
	schemaCatalogVersion  = "manifests/schema.yaml"
	configCatalogVersion  = "manifests/config.yaml"
	wrongCatalogVersion   = "manifests/wrongconfig.yaml"

	testManifests = []string{
		"manifests/CatalogVersionforConfigMap1.yaml",
		"manifests/CatalogVersionforConfigMap2.yaml",
		"manifests/solutionversion3.yaml",
		"manifests/target3.yaml",
		"manifests/instanceForConfigMap.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

var (
	NAMESPACES = []string{
		"default",
		"nondefault",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer testhelpers.Cleanup(TEST_NAME)

	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}

	err = Verify()
	if err != nil {
		return err
	}

	return nil
}

// Run tests
func Verify() error {
	//CATALOG CRUD, needs to create a catalogversion yaml
	for _, namespace := range NAMESPACES {
		os.Setenv("NAMESPACE", namespace)
		err := testhelpers.EnsureNamespace(namespace)
		if err != nil {
			return err
		}

		err = deployNamespaceManifests(namespace)
		if err != nil {
			return err
		}

		// Deploy solutionversion, target and instance catalogversions
		err = createCatalogVersions(namespace)
		if err != nil {
			return err
		}
		// List catalogversions
		config, err := testhelpers.RestConfig()
		if err != nil {
			return err
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
		err, catalogversions := listCatalogVersions(namespace, dynamicClient)
		if err != nil {
			return err
		}
		if len(catalogversions.Items) < 4 {
			fmt.Printf("CatalogVersions not created. Expected 4, got %d\n", len(catalogversions.Items))
			return errors.New("CatalogVersions not created")
		}
		// read catalogversion
		err, catalogversion := readCatalogVersion("asset-v-version1", namespace, dynamicClient)
		if err != nil {
			return err
		}
		if catalogversion.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "東京" {
			return errors.New("CatalogVersion not created correctly.")
		}
		// Update catalogversion
		catalogversion.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] = "大阪"
		err, catalogversion = updateCatalogVersion("asset-v-version1", namespace, catalogversion, dynamicClient)
		if err != nil {
			return err
		}
		if catalogversion.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "大阪" {
			return errors.New("CatalogVersion not updated.")
		}
		// Delete catalogversion
		err = shellcmd.Command(fmt.Sprintf("kubectl delete catalogversion asset-v-version1 -n %s", namespace)).Run()
		if err != nil {
			return err
		}
		fmt.Printf("CatalogVersion integration test finished for namespace: %s\n", namespace)

		// Deploy manifests for configmap
		currentPath, err := os.Getwd()
		if err != nil {
			return err
		}
		for _, manifest := range testManifests {
			fullPath := filepath.Join(currentPath, manifest)
			err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", fullPath, namespace)).Run()
			if err != nil {
				return err
			}
		}

		err = shellcmd.Command("go clean -testcache").Run()
		if err != nil {
			return err
		}
		os.Setenv("SYMPHONY_FLAVOR", "oss")
		for _, verify := range testVerify {
			err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
			if err != nil {
				return err
			}
		}

	}
	fmt.Printf("CatalogVersion & configmap integration test finished successfully\n")
	return nil
}

// Deploy manifests for namespace
func deployNamespaceManifests(namespace string) error {
	// setup campaign
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, manifest := range testNamespace {
		absManifest := filepath.Join(currentPath, manifest)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absManifest, namespace)).Run()
		if err != nil {
			return err
		}
	}

	// setup activation
	absActivation := filepath.Join(currentPath, testActivation)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absActivation, namespace)).Run()
	if err != nil {
		return err
	}

	return nil
}

// Create catalogversions
func createCatalogVersions(namespace string) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, catalogversion := range testCatalogVersions {
		absCatalogVersion := filepath.Join(currentPath, catalogversion)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absCatalogVersion, namespace)).Run()
		if err != nil {
			return err
		}
	}
	// Deploy config catalogversion before schema catalogversion
	configPath := filepath.Join(currentPath, configCatalogVersion)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", configPath, namespace)).Run()
	if err == nil {
		return errors.New("CatalogVersion using schema should not be deployed before schema catalogversion being deployed.")
	}
	// Deploy schema catalogversion
	schemaPath := filepath.Join(currentPath, schemaCatalogVersion)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", schemaPath, namespace)).Run()
	if err != nil {
		return err
	}
	// Deploy config catalogversion
	configPath = filepath.Join(currentPath, configCatalogVersion)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", configPath, namespace)).Run()
	if err != nil {
		return err
	}
	//Deploy wrong catalogversion
	wrongPath := filepath.Join(currentPath, wrongCatalogVersion)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", wrongPath, namespace)).Run()
	if err == nil {
		return errors.New("Wrong catalogversion should not be deployed")
	}
	return nil
}

func readCatalogVersion(catalogversionName string, namespace string, dynamicClient dynamic.Interface) (error, *unstructured.Unstructured) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogversions"}
	catalogversion, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), catalogversionName, metav1.GetOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalogversion
}

func updateCatalogVersion(catalogversionName string, namespace string, object *unstructured.Unstructured, dynamicClient dynamic.Interface) (error, *unstructured.Unstructured) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogversions"}
	catalogversion, err := dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), object, metav1.UpdateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalogversion
}

func listCatalogVersions(namespace string, dynamicClient dynamic.Interface) (error, *unstructured.UnstructuredList) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogversions"}
	catalogversions, err := dynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalogversions
}
