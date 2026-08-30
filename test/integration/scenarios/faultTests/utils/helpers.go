/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
)

// failpointBaseURL dials a fresh port-forward to the current running pod for
// the label and returns the base URL of its failpoint server plus a cleanup
// func. The port-forward created at scenario start does not survive the pod
// restarting after a panic fault, so every probe/injection dials a new one
// against the live pod.
func failpointBaseURL(podLabel string) (string, func(), error) {
	config, err := testhelpers.RestConfig()
	if err != nil {
		return "", nil, err
	}
	clientset, err := testhelpers.KubeClient()
	if err != nil {
		return "", nil, err
	}
	podList, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{
		LabelSelector: podLabel,
	})
	if err != nil {
		return "", nil, err
	}
	podName := ""
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodRunning {
			podName = pod.Name
			break
		}
	}
	if podName == "" {
		return "", nil, fmt.Errorf("no running pod for label %s", podLabel)
	}

	url := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace("default").
		Name(podName).
		SubResource("portforward").
		URL()
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return "", nil, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})
	// Local port 0: let the OS pick a free port
	forwarder, err := portforward.New(dialer, []string{"0:" + LocalPortForward}, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return "", nil, err
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- forwarder.ForwardPorts()
	}()
	select {
	case <-readyChan:
	case err := <-errCh:
		return "", nil, err
	case <-time.After(15 * time.Second):
		close(stopChan)
		return "", nil, fmt.Errorf("timeout waiting for port-forwarding to be ready")
	}
	ports, err := forwarder.GetPorts()
	if err != nil {
		close(stopChan)
		return "", nil, err
	}
	return fmt.Sprintf("http://127.0.0.1:%d", ports[0].Local), func() { close(stopChan) }, nil
}

// requestFailpoint issues a request to the pod's failpoint server, retrying
// with a fresh port-forward each time while the pod recovers from a previous
// fault (e.g. a panic restart).
func requestFailpoint(podLabel string, method string, path string, body string) error {
	var lastErr error
	for i := 0; i < 18; i++ {
		base, cleanup, err := failpointBaseURL(podLabel)
		if err == nil {
			var resp *http.Response
			req, err := http.NewRequest(method, strings.TrimSuffix(base, "/")+"/"+path, strings.NewReader(body))
			if err == nil {
				client := &http.Client{Timeout: 10 * time.Second}
				resp, err = client.Do(req)
			}
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode < 300 {
					cleanup()
					return nil
				}
				err = fmt.Errorf("failpoint server returned %s", resp.Status)
			}
			cleanup()
		}
		lastErr = err
		fmt.Println("failed to connect to failpoint server, waiting...")
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("timeout waiting for failpoint server of %s: %v", podLabel, lastErr)
}

func InjectPodFailure() error {
	PodLabel := os.Getenv(PodEnvKey)
	Fault := os.Getenv(FaultNameEnvKey)
	FaultType := os.Getenv(FaultTypeEnvKey)
	if Fault == "" || PodLabel == "" {
		fmt.Println("Fault is ", Fault, "and InjectPodLabel is ", PodLabel, ", skip error injection")
		return nil
	}

	if err := requestFailpoint(PodLabel, http.MethodPut, Fault, FaultType); err != nil {
		fmt.Println("Failed to inject pod failure: " + err.Error())
		return err
	}
	fmt.Println("Injected fault")
	return nil
}
