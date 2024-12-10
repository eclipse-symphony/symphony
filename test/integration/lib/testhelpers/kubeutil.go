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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
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

func EnablePortForward(podlabel string, port string, stopChan chan struct{}) error {
	config, err := RestConfig()
	if err != nil {
		return err
	}

	clientset, err := KubeClient()
	if err != nil {
		return err
	}
	pods := clientset.CoreV1().Pods("default")
	podList, err := pods.List(context.Background(), metav1.ListOptions{
		LabelSelector: podlabel,
	})
	if err != nil {
		return err
	}
	pod := podList.Items[0]

	// Create a port-forward request
	url := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace("default").
		Name(pod.Name).
		SubResource("portforward").
		URL()
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	// Set up port-forwarding
	ports := []string{fmt.Sprintf("%s:%s", port, port)}
	readyChan := make(chan struct{})
	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, os.Stdout, os.Stderr)
	errCh := make(chan error)
	go func() {
		errCh <- forwarder.ForwardPorts()
		if err != nil {
			fmt.Printf("Error in port-forwarding: %v\n", err)
		}
	}()

	// Wait for the port-forwarding to be ready
	select {
	case <-readyChan:
		fmt.Println("Port-forwarding is ready")
		return nil
	case <-time.After(time.Second * 10):
		return fmt.Errorf("timeout waiting for port-forwarding to be ready")
	case err = <-errCh:
		return fmt.Errorf("forwarding ports: %v", err)
	}
}

func WaitPodOnline(podlabel string) error {
	clientset, err := KubeClient()
	if err != nil {
		return err
	}
	pods := clientset.CoreV1().Pods("default")
	for i := 0; i < 10; i++ {
		podList, err := pods.List(context.Background(), metav1.ListOptions{
			LabelSelector: podlabel,
		})
		if err != nil {
			return err
		}
		if len(podList.Items) > 0 {
			pod := podList.Items[0]
			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
			fmt.Println("pod not ready yet, waiting..." + pod.Status.Phase)
		}
		time.Sleep(time.Second * 10)
	}
	return fmt.Errorf("timeout waiting for pod to be ready")
}
