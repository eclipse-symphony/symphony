/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

// testhelpers contains helpers for tests
package testhelpers

import (
	"context"
	"flag"
	"path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	configInst      *rest.Config
	configInstMutex sync.Mutex
)

// KubeClient returns the kubectl client
func KubeClient() (*kubernetes.Clientset, error) {
	config, err := RestConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// RestConfig returns the kube config
func RestConfig() (*rest.Config, error) {
	configInstMutex.Lock()
	defer configInstMutex.Unlock()

	if configInst == nil {
		// based on the example from https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}

		configInst = config
	}

	return configInst, nil
}

// Ensures that the namespace exists. If it does not exist, it creates it.
func EnsureNamespace(namespace string) error {
	kubeClient, err := KubeClient()
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
