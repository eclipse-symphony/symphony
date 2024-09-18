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
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func TestBasic_ReadNondefaultNamespaceActivation(t *testing.T) {
	fmt.Printf("Checking activation\n")
	namespace := os.Getenv("NAMESPACE")
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
			require.Equal(t, 2, len(state.Status.StageHistory))
			require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
			require.Equal(t, v1alpha2.Done, state.Status.StageHistory[1].Status)
			require.Equal(t, 3, len(state.Status.StageHistory[1].Inputs))
			require.Equal(t, 10., state.Status.StageHistory[1].Inputs["age"])
			require.Equal(t, "worker", state.Status.StageHistory[1].Inputs["job"])
			require.Equal(t, "sample", state.Status.StageHistory[1].Inputs["name"])
			break
		}

		sleepDuration, _ := time.ParseDuration("5s")
		time.Sleep(sleepDuration)
	}
}
