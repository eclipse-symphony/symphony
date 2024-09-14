//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Test config
const (
	TEST_NAME = "Symphony Catalog CRUD test scenario"
)

var (
	// catalogs to deploy
	testCatalogs = []string{
		"catalogs/instance-container.yaml",
		"catalogs/solution-container.yaml",
		"catalogs/target-container.yaml",
		"catalogs/asset-container.yaml",
		"catalogs/config-container.yaml",
		"catalogs/wrongconfig-container.yaml",
		"catalogs/schema-container.yaml",

		"catalogs/instance.yaml",
		"catalogs/solution.yaml",
		"catalogs/target.yaml",
		"catalogs/asset.yaml",
	}

	// test in nondefault namespace only
	testNondefaultManifests = []string{
		"nondefault/config-container.yaml",
		"nondefault/config.yaml",
		"nondefault/campaign-container.yaml",
		"nondefault/campaign.yaml",
	}

	testActivation = "nondefault/activation.yaml"
	schemaCatalog  = "catalogs/schema.yaml"
	configCatalog  = "catalogs/config.yaml"
	wrongCatalog   = "catalogs/wrongconfig.yaml"
)

var (
	NAMESPACES = []string{
		"default",
		//"nondefault",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	// defer testhelpers.Cleanup(TEST_NAME)

	// err := testhelpers.SetupCluster()
	// if err != nil {
	// 	return err
	// }

	// err = Verify()
	// if err != nil {
	// 	return err
	// }

	err := VerifyWorkflow("nondefault")
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
		// Deploy solution, target and instance catalogs
		err := createCatalogs(namespace)
		if err != nil {
			return err
		}
		// List catalogs
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
		err, catalog := readCatalog("asset-v-v1", namespace, dynamicClient)
		if err != nil {
			return err
		}
		if catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "東京" {
			return errors.New("Catalog not created correctly.")
		}
		// Update catalog
		catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] = "大阪"
		err, catalog = updateCatalog("asset-v-v1", namespace, catalog, dynamicClient)
		if err != nil {
			return err
		}
		if catalog.Object["spec"].(map[string]interface{})["properties"].(map[string]interface{})["name"] != "大阪" {
			return errors.New("Catalog not updated.")
		}
		// Delete catalog
		err = shellcmd.Command(fmt.Sprintf("kubectl delete catalog asset-v-v1 -n %s", namespace)).Run()
		if err != nil {
			return err
		}
		fmt.Printf("Catalog integration test finished for namespace: %s\n", namespace)
	}
	fmt.Printf("Catalog integration test finished successfully\n")
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

// Verify namespace in nondefault namespace
func VerifyWorkflow(namespace string) error {
	// Ensure that namespace is defined
	err := testhelpers.EnsureNamespace(namespace)
	if err != nil {
		return err
	}

	os.Setenv("NAMESPACE", namespace)
	err = createActivation(namespace)
	if err != nil {
		return err
	}

	err = readActivation(namespace)
	if err != nil {
		return err
	}

	return nil
}

func createActivation(namespace string) error {
	// setup campaign
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, manifest := range testNondefaultManifests {
		absManifest := filepath.Join(currentPath, manifest)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absManifest, namespace)).Run()
		if err != nil {
			return err
		}
	}

	// wait for 5 seconds to make sure campaign is created
	time.Sleep(time.Second * 5)
	absActivation := filepath.Join(currentPath, testActivation)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absActivation, namespace)).Run()
	if err != nil {
		return err
	}

	return nil
}

func readActivation(namespace string) error {
	crd := &unstructured.Unstructured{}
	crd.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "workflow.symphony",
		Version: "v1",
		Kind:    "Activation",
	})

	cfg, err := testhelpers.RestConfig()
	if err != nil {
		return err
	}

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "workflow.symphony",
			Version:  "v1",
			Resource: "activations",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		if len(resources.Items) != 1 {
			return fmt.Errorf("there should be only one activation")
		}

		bytes, _ := json.Marshal(resources.Items[0].Object)
		var state model.ActivationState
		err = json.Unmarshal(bytes, &state)
		if err != nil {
			return err
		}
		status := state.Status.Status
		fmt.Printf("Current activation status: %s\n", status)
		if status == v1alpha2.Done {
			if len(state.Status.StageHistory) != 2 {
				return fmt.Errorf("there should be two stages")
			}
			if state.Status.StageHistory[0].Status != v1alpha2.Done {
				return fmt.Errorf("first stage status is %s", state.Status.StageHistory[0].Status)
			}
			if state.Status.StageHistory[1].Status != v1alpha2.Done {
				return fmt.Errorf("first stage status is %s", state.Status.StageHistory[0].Status)
			}
			if state.Status.StageHistory[1].Inputs["name"] != "sample" {
				return fmt.Errorf("Second stage input is incorrect")
			}
			break
		}

		sleepDuration, _ := time.ParseDuration("5s")
		time.Sleep(sleepDuration)
	}

	return nil
}
