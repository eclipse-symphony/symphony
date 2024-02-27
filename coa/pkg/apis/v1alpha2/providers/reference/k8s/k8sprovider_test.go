/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func TestInit(t *testing.T) {
	testRedis := os.Getenv("TEST_K8S")
	if testRedis == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReferenceProvider{}
	err := provider.Init(K8sReferenceProviderConfig{})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testRedis := os.Getenv("TEST_K8S")
	symphonySolution := os.Getenv("SYMPHONY_SOLUTION")
	if testRedis == "" || symphonySolution == "" {
		t.Skip("Skipping because TEST_K8S or SYMPHONY_SOLUTION enviornment variable is not set")
	}
	provider := K8sReferenceProvider{}
	err := provider.Init(K8sReferenceProviderConfig{})
	assert.Nil(t, err)
	_, err = provider.Get(symphonySolution, "default", "solution.symphony", "solutions", "v1", "")
	assert.NotNil(t, err)
}
func TestK8sReferenceProviderConfigFromMapMapNil(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapBadInClusterValue(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"inCluster": "bad",
	})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}
func TestK8sReferenceProviderConfigFromMap(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"configPath": "my-path",
		"inCluster":  "true",
	})
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-path", "true-path")
	os.Setenv("my-name", "true-name")
	config, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"name":       "$env:my-name",
		"configPath": "$env:my-path",
		"inCluster":  "true",
	})
	assert.Nil(t, err)
	assert.Equal(t, "true-path", config.ConfigPath)
	assert.Equal(t, "true-name", config.Name)
}

func TestK8sGet(t *testing.T) {
	TEST_MINIKUBE_ENABLED := os.Getenv("TEST_MINIKUBE_ENABLED")
	if TEST_MINIKUBE_ENABLED == "" {
		t.Skip("Skipping because TEST_MINIKUBE_ENABLED enviornment variable is not set")
	}
	provider := K8sReferenceProvider{}
	err := provider.InitWithMap(nil)
	assert.Nil(t, err)
	CreateTargetResource()
	target, err := provider.Get("ut-target", "default", "fabric.symphony", "targets", "v1", "")
	assert.Nil(t, err)
	assert.NotNil(t, target)
}

func TestK8sList(t *testing.T) {
	TEST_MINIKUBE_ENABLED := os.Getenv("TEST_MINIKUBE_ENABLED")
	if TEST_MINIKUBE_ENABLED == "" {
		t.Skip("Skipping because TEST_MINIKUBE_ENABLED enviornment variable is not set")
	}
	config, err := K8sReferenceProviderConfigFromMap(nil)
	assert.Nil(t, err)
	provider := K8sReferenceProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	CreateTargetResource()
	targets, err := provider.List("", "", "default", "fabric.symphony", "targets", "v1", "")
	tt := targets.([]interface{})
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(tt), 1)
}

func TestK8sClone(t *testing.T) {
	config, err := K8sReferenceProviderConfigFromMap(nil)
	assert.Nil(t, err)
	provider := K8sReferenceProvider{}
	err = provider.Init(config)
	assert.Nil(t, err)
	newProvider, err := provider.Clone(nil)
	assert.Nil(t, err)
	assert.NotNil(t, newProvider)
	k8sConfig, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"name": "ut-k8s",
	})
	assert.Nil(t, err)
	newProvider, err = provider.Clone(k8sConfig)
	assert.Nil(t, err)
	assert.NotNil(t, newProvider)
}

func CreateTargetResource() {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	target := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "fabric.symphony/v1",
			"kind":       "target",
			"metadata": map[string]interface{}{
				"name": "ut-target",
			},
			"spec": map[string]interface{}{
				"forceRedeploy": true,
				"topologies": []interface{}{
					map[string]interface{}{
						"bindings": []interface{}{
							map[string]interface{}{
								"role":     "instance",
								"provider": "providers.target.k8s",
								"config": map[string]interface{}{
									"inCluster": "true",
								},
							},
						},
					},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "fabric.symphony",
		Version:  "v1",
		Resource: "targets",
	}

	_, err = dynamicClient.Resource(gvr).Namespace("default").Create(context.Background(), target, metav1.CreateOptions{})
	if !strings.Contains(err.Error(), "is already taken") {
		panic(err)
	}
}
