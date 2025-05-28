/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
)

func WaitFailpointServer(podlabel string) error {
	clientset, err := testhelpers.KubeClient()
	if err != nil {
		return err
	}
	err = testhelpers.WaitPodOnline(podlabel)
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
				err = testhelpers.ShellExec(fmt.Sprintf("kubectl exec %s -- curl localhost:22381", pod.Name))
				if err == nil {
					return nil
				} else {
					fmt.Println("failed to connect to failpoint server, waiting...")
				}
			} else {
				fmt.Println("pod not ready yet, waiting..." + pod.Status.Phase)
			}
		}
		time.Sleep(time.Second * 10)
	}
	return fmt.Errorf("timeout waiting for pod to be ready")
}

func InjectPodFailure() error {
	InjectCommand := os.Getenv(InjectFaultEnvKey)
	PodLabel := os.Getenv(PodEnvKey)
	if InjectCommand == "" || PodLabel == "" {
		fmt.Println("InjectCommand is ", InjectCommand, "and InjectPodLabel is ", PodLabel, ", skip error injection")
		return nil
	}

	WaitFailpointServer(PodLabel)
	err := shellcmd.Command(InjectCommand).Run()
	if err != nil {
		fmt.Println("Failed to inject pod failure: " + err.Error())
	}
	fmt.Println("Injected fault")
	return err
}

func DeletePodFailure() error {
	DeleteCommand := os.Getenv(DeleteFaultEnvKey)
	PodLabel := os.Getenv(PodEnvKey)
	if DeleteCommand == "" || PodLabel == "" {
		fmt.Println("DeleteCommand is ", DeleteCommand, "and PodLabel is ", PodLabel, ", skip error injection")
		return nil
	}

	WaitFailpointServer(PodLabel)
	err := shellcmd.Command(DeleteCommand).Run()
	if err != nil {
		fmt.Println("Failed to delete pod failure: " + err.Error())
	}
	fmt.Println("Deleted fault")
	return err
}
