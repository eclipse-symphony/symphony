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

// Verify catalogversion created
func TestBasic_CatalogVersions(t *testing.T) {
	fmt.Printf("Checking CatalogVersions\n")
	namespace := "default"

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
		if len(resources.Items) == 4 {
			break
		}

		sleepDuration, _ := time.ParseDuration("30s")
		time.Sleep(sleepDuration)
	}
}

// Verify eval status
func Test_CatalogVersionsEvals(t *testing.T) {
	fmt.Printf("Checking evals\n")

	namespace := "default"
	cfg, err := testhelpers.RestConfig()
	require.NoError(t, err)

	dyn, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)
	var evaluateevalcatalogversion *unstructured.Unstructured
	retryWithTimeout(func() (any, error) {
		evaluateevalcatalogversion, err = dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogversionevalexpressions",
		}).Namespace(namespace).Get(context.Background(), "evaluateevalcatalogversion01", metav1.GetOptions{})
		require.NoError(t, err)
		status, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "status")
		require.NoError(t, err)
		require.Contains(t, []string{"Succeeded", "Failed"}, status)
		return evaluateevalcatalogversion, nil
	}, time.Minute*1)
	resourceRef, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "spec", "resourceRef", "name")
	require.NoError(t, err)
	require.Equal(t, "evalcatalogversion-v-version1", resourceRef)

	status, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "evaluationStatus")
	require.NoError(t, err)
	require.Equal(t, "Failed", status)

	address, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "address")
	require.NoError(t, err)
	require.Equal(t, "1st Avenue", address)

	city, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "city")
	require.NoError(t, err)
	require.Equal(t, "Sydney", city)

	zipcode, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "zipcode")
	require.NoError(t, err)
	require.Contains(t, zipcode, "Not Found")

	county, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "county")
	require.NoError(t, err)
	require.Contains(t, county, "Not Found")

	country, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "country")
	require.NoError(t, err)
	require.Contains(t, country, "Bad Config")

	fromCountry, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "from", "country")
	require.NoError(t, err)
	require.Equal(t, "Australia", fromCountry)

	// check evaluateevalcatalogversion02
	retryWithTimeout(func() (any, error) {
		evaluateevalcatalogversion, err = dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogversionevalexpressions",
		}).Namespace(namespace).Get(context.Background(), "evaluateevalcatalogversion02", metav1.GetOptions{})
		require.NoError(t, err)
		status, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "status")
		require.NoError(t, err)
		require.Contains(t, []string{"Succeeded", "Failed"}, status)
		return evaluateevalcatalogversion, nil
	}, time.Minute*1)
	code, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "error", "code")
	require.NoError(t, err)
	require.Equal(t, "ParentRefNotFound", code)

	// check evaluateevalcatalogversion03
	retryWithTimeout(func() (any, error) {
		evaluateevalcatalogversion, err = dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogversionevalexpressions",
		}).Namespace(namespace).Get(context.Background(), "evaluateevalcatalogversion03", metav1.GetOptions{})
		require.NoError(t, err)
		status, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "status")
		require.NoError(t, err)
		require.Contains(t, []string{"Succeeded", "Failed"}, status)
		return evaluateevalcatalogversion, nil
	}, time.Minute*1)
	status, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "evaluationStatus")
	require.NoError(t, err)
	require.Equal(t, "Succeeded", status)

	city, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "city")
	require.NoError(t, err)
	require.Equal(t, "Sydney", city)

	// check evaluateevalcatalogversion04
	retryWithTimeout(func() (any, error) {
		evaluateevalcatalogversion, err = dyn.Resource(schema.GroupVersionResource{
			Group:    "federation.symphony",
			Version:  "v1",
			Resource: "catalogversionevalexpressions",
		}).Namespace(namespace).Get(context.Background(), "evaluateevalcatalogversion04", metav1.GetOptions{})
		require.NoError(t, err)
		status, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "status")
		require.NoError(t, err)
		require.Contains(t, []string{"Succeeded", "Failed"}, status)
		return evaluateevalcatalogversion, nil
	}, time.Minute*1)
	status, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "evaluationStatus")
	require.NoError(t, err)
	require.Equal(t, "Failed", status)

	address, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "address")
	require.NoError(t, err)
	require.Equal(t, "1st Avenue", address)

	city, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "city")
	require.NoError(t, err)
	require.Equal(t, "Sydney", city)

	zipcode, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "zipcode")
	require.NoError(t, err)
	require.Contains(t, zipcode, "Not Found")

	county, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "county")
	require.NoError(t, err)
	require.Contains(t, county, "Not Found")

	country, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "country")
	require.NoError(t, err)
	require.Equal(t, "Australia", country)

	fromCountry, _, err = unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "from", "country")
	require.NoError(t, err)
	require.Equal(t, "Australia", fromCountry)

	fromState, _, err := unstructured.NestedString(evaluateevalcatalogversion.Object, "status", "actionStatus", "output", "from", "state")
	require.NoError(t, err)
	require.Equal(t, "Virginia", fromState)
}

func retryWithTimeout(fn func() (any, error), timeout time.Duration) (any, error) {
	start := time.Now()
	for {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout while waiting for function to succeed: %w", err)
		}
		time.Sleep(time.Second * 5)
	}
}
