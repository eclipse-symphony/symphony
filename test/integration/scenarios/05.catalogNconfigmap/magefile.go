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
	TEST_NAME    = "Symphony Catalog CRUD test scenario"
	TEST_TIMEOUT = "10m"
)

var (
	// catalogs to deploy
	testCatalogs = []string{
		"manifests/instance-container.yaml",
		"manifests/solution-container.yaml",
		"manifests/target-container.yaml",
		"manifests/asset-container.yaml",
		"manifests/config-container.yaml",
		"manifests/wrongconfig-container.yaml",
		"manifests/schema-container.yaml",

		"manifests/instance.yaml",
		"manifests/solution.yaml",
		"manifests/target.yaml",
		"manifests/asset.yaml",
	}

	// catalogs for namespace test
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
	schemaCatalog  = "manifests/schema.yaml"
	configCatalog  = "manifests/config.yaml"
	wrongCatalog   = "manifests/wrongconfig.yaml"

	testManifests = []string{
		"manifests/CatalogforConfigMap1.yaml",
		"manifests/CatalogforConfigMap2.yaml",
		"manifests/solution3.yaml",
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
	//CATALOG CRUD, needs to create a catalog yaml
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

		// Deploy solution, target and instance catalogs
		err = createCatalogs(namespace)
		if err != nil {
			return err
		}
		// List catalogs
		config, err := testhelpers.RestConfig()
		if err != nil {
			return err
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
		err, catalogs := listCatalogs(namespace, dynamicClient)
		if err != nil {
			return err
		}
		if len(catalogs.Items) < 4 {
			fmt.Printf("Catalogs not created. Expected 4, got %d\n", len(catalogs.Items))
			return errors.New("Catalogs not created")
		}
		// read catalog
		err, catalog := readCatalog("asset-v-version1", namespace, dynamicClient)
		if err != nil {
			return err
		}
		if catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "東京" {
			return errors.New("Catalog not created correctly.")
		}
		// Update catalog
		catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] = "大阪"
		err, catalog = updateCatalog("asset-v-version1", namespace, catalog, dynamicClient)
		if err != nil {
			return err
		}
		if catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "大阪" {
			return errors.New("Catalog not updated.")
		}
		// Delete catalog
		err = shellcmd.Command(fmt.Sprintf("kubectl delete catalog asset-v-version1 -n %s", namespace)).Run()
		if err != nil {
			return err
		}
		fmt.Printf("Catalog integration test finished for namespace: %s\n", namespace)

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
	fmt.Printf("Catalog & configmap integration test finished successfully\n")
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

// Create catalogs
func createCatalogs(namespace string) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, catalog := range testCatalogs {
		absCatalog := filepath.Join(currentPath, catalog)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absCatalog, namespace)).Run()
		if err != nil {
			return err
		}
	}
	// Deploy config catalog before schema catalog
	configPath := filepath.Join(currentPath, configCatalog)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", configPath, namespace)).Run()
	if err == nil {
		return errors.New("Catalog using schema should not be deployed before schema catalog being deployed.")
	}
	// Deploy schema catalog
	schemaPath := filepath.Join(currentPath, schemaCatalog)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", schemaPath, namespace)).Run()
	if err != nil {
		return err
	}
	// Deploy config catalog
	configPath = filepath.Join(currentPath, configCatalog)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", configPath, namespace)).Run()
	if err != nil {
		return err
	}
	//Deploy wrong catalog
	wrongPath := filepath.Join(currentPath, wrongCatalog)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", wrongPath, namespace)).Run()
	if err == nil {
		return errors.New("Wrong catalog should not be deployed")
	}
	return nil
}

func readCatalog(catalogName string, namespace string, dynamicClient dynamic.Interface) (error, *unstructured.Unstructured) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogs"}
	catalog, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), catalogName, metav1.GetOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalog
}

func updateCatalog(catalogName string, namespace string, object *unstructured.Unstructured, dynamicClient dynamic.Interface) (error, *unstructured.Unstructured) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogs"}
	catalog, err := dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), object, metav1.UpdateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalog
}

func listCatalogs(namespace string, dynamicClient dynamic.Interface) (error, *unstructured.UnstructuredList) {
	gvr := schema.GroupVersionResource{Group: "federation.symphony", Version: "v1", Resource: "catalogs"}
	catalogs, err := dynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err, nil
	}
	return nil, catalogs
}
