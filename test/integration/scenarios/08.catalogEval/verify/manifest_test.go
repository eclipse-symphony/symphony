/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Verify catalog created
func TestBasic_Catalogs(t *testing.T) {
	fmt.Printf("Checking Catalogs\n")
	namespace := "default"

	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	for {
		resources, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogs",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		catalogs := []string{}
		for _, item := range resources.Items {
			catalogs = append(catalogs, item.GetName())
		}
		fmt.Printf("Catalogs: %v\n", catalogs)
		if len(resources.Items) == 3 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify eval status
func Test_CatalogsEvals(t *testing.T) {
	fmt.Printf("Checking evals\n")

	namespace := "default"
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)

	var resources *unstructured.UnstructuredList
	for {
		resources, err = dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogevalexpressions",
		}).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		require.NoError(t, err)

		catalogEvals := []string{}
		for _, item := range resources.Items {
			catalogEvals = append(catalogEvals, item.GetName())
		}
		fmt.Printf("CatalogEvalExpression: %v\n", catalogEvals)
		if len(resources.Items) == 3 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}

	//check status
	for _, item := range resources.Items {
		if item.GetName() == "evaluateevalcatalog01" {
			// check the spec.resourceRef.Name
			resourceRef, _, err := unstructured.NestedString(item.Object, "spec", "resourceRef", "name")
			require.NoError(t, err)
			require.Equal(t, "evalcatalog-v-v1", resourceRef)

			status, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "status")
			require.NoError(t, err)
			require.Equal(t, "Succeeded", status)

			status, _, err = unstructured.NestedString(item.Object, "status", "actionStatus", "output", "evaluationStatus")
			require.NoError(t, err)
			require.Equal(t, "Failed", status)

			address, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "address")
			require.NoError(t, err)
			require.Equal(t, "1st Avenue", address)

			city, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "city")
			require.NoError(t, err)
			require.Equal(t, "Sydney", city)

			zipcode, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "zipcode")
			require.NoError(t, err)
			require.Contains(t, zipcode, "Not Found")

			county, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "county")
			require.NoError(t, err)
			require.Contains(t, county, "Not Found")

			country, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "country")
			require.NoError(t, err)
			require.Contains(t, country, "Bad Config")

			fromCountry, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "from", "country")
			require.NoError(t, err)
			require.Equal(t, "Australia", fromCountry)
		}
		if item.GetName() == "evaluateevalcatalog02" {
			status, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "status")
			require.NoError(t, err)
			require.Equal(t, "Failed", status)

			code, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "error", "code")
			require.NoError(t, err)
			require.Equal(t, "ParentRefNotFound", code)
		}
		if item.GetName() == "evaluateevalcatalog03" {
			status, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "status")
			require.NoError(t, err)
			require.Equal(t, "Succeeded", status)

			status, _, err = unstructured.NestedString(item.Object, "status", "actionStatus", "output", "evaluationStatus")
			require.NoError(t, err)
			require.Equal(t, "Succeeded", status)

			city, _, err := unstructured.NestedString(item.Object, "status", "actionStatus", "output", "city")
			require.NoError(t, err)
			require.Equal(t, "Sydney", city)
		}
	}

}
