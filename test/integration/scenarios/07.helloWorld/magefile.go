//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	NAMESPACE = "sample-k8s-scope"
	TEST_NAME = "Symphony Hello World sample test scenario"
	TEST_TIMEOUT = "10m"
)

var (
	// Manifests to deploy
	testManifests = []string{
		"../../../../docs/samples/k8s/hello-world/solution-container.yaml",
		"../../../../docs/samples/k8s/hello-world//solution.yaml",
		"../../../../docs/samples/k8s/hello-world//target.yaml",
		"../../../../docs/samples/k8s/hello-world//instance.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
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

	// Deploy solution, target and instance
	err = DeployManifests()
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
	err := shellcmd.Command("go clean -testcache").Run()
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

	return nil
}

// Deploy solution, target and instance
func DeployManifests() error {
	// Get kube client
	err := ensureNamespace(NAMESPACE)
	if err != nil {
		return err
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}

	// Deploy the manifests
	for _, manifest := range testManifests {
		manifestPath := filepath.Join(currentPath, manifest)
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", manifestPath, NAMESPACE)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensures that the namespace exists. If it does not exist, it creates it.
func ensureNamespace(namespace string) error {
	kubeClient, err := testhelpers.KubeClient()
	if err != nil {
		return err
	}

	_, err = kubeClient.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if kerrors.IsNotFound(err) {
		_, err = kubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}