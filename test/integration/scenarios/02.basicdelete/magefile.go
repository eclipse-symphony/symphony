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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test config
const (
	TEST_NAME    = "basic delete"
	TEST_TIMEOUT = "10m"
)

var (
	NAMESPACES = []string{
		"default",
		"nondefault",
	}
)

var (
	// Manifests to deploy
	testManifests = []string{
		"manifest/target.yaml",
		"manifest/instance.yaml",
		"manifest/solution.yaml",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer Cleanup()
	err := SetupCluster()
	if err != nil {
		return err
	}
	for _, namespace := range NAMESPACES {
		os.Setenv("NAMESPACE", namespace)
		err = DeployManifests(namespace)
		if err != nil {
			return err
		}

		err = VerifyPodExists()
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 10)

		err = CleanUpSymphonyObjects(namespace)
		if err != nil {
			return err
		}

		err = VerifyPodNotExists()
		if err != nil {
			return err
		}
	}

	return nil
}

// Prepare the cluster
// Run this manually to prepare your local environment for testing/debugging
func SetupCluster() error {
	// Deploy symphony
	err := localenvCmd("cluster:deploy", "")
	if err != nil {
		return err
	}
	// Wait a few secs for symphony cert to be ready;
	// otherwise we will see error when creating symphony manifests in the cluster
	// <Error from server (InternalError): error when creating
	// "/mnt/vss/_work/1/s/test/integration/scenarios/basic/manifest/target.yaml":
	// Internal error occurred: failed calling webhook "mtarget.kb.io": failed to
	// call webhook: Post
	// "https://symphony-webhook-service.default.svc:443/mutate-symphony-microsoft-com-v1-target?timeout=10s":
	// x509: certificate signed by unknown authority>
	time.Sleep(time.Second * 10)
	return nil
}

func DeployManifests(namespace string) error {
	if namespace != "default" {
		// Create non-default namespace if not exist
		err := shellcmd.Command(fmt.Sprintf("kubectl get namespace %s", namespace)).Run()
		if err != nil {
			// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
			err = shellcmd.Command(fmt.Sprintf("kubectl create namespace %s", namespace)).Run()
			if err != nil {
				return err
			}
		}
	}
	// Deploy the manifests
	for _, manifest := range testManifests {
		fullPath, err := filepath.Abs(manifest)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		stringYaml := string(data)
		stringYaml = strings.ReplaceAll(stringYaml, "SCOPENAME", namespace+"scope")

		err = writeYamlStringsToFile(stringYaml, "./test.yaml")
		if err != nil {
			return err
		}
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
		if err != nil {
			return err
		}
		os.Remove("./test.yaml")
	}

	return nil
}

// Run tests
func VerifyPodExists() error {
	kubeClient, err := testhelpers.KubeClient()
	if err != nil {
		return err
	}
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	i := 0
	for {
		i++
		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods(namespace+"scope").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		// Verify that the pods we expect are running
		toFind := []string{"instance02"}

		notFound := make(map[string]bool)
		for _, s := range toFind {
			found := false
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, s) && pod.Status.Phase == "Running" {
					found = true
					break
				}
			}

			if !found {
				notFound[s] = true
			}
		}

		if len(notFound) == 0 {
			fmt.Println("All pods found!")
			break
		} else {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods: %v\n", notFound)
			}
		}
	}

	return nil
}

func CleanUpSymphonyObjects(namespace string) error {
	targetName := "target02-v1"
	err := shellcmd.Command(fmt.Sprintf("kubectl delete targets.fabric.symphony %s -n %s", targetName, namespace)).Run()
	if err != nil {
		return err
	}

	/*
		instanceName := "instance02-v1"
		solutionName := "solution02-v1"
		err = shellcmd.Command(fmt.Sprintf("kubectl delete instances.solution.symphony %s -n %s", instanceName, namespace)).Run()
		if err != nil {
			return err
		}

		err = shellcmd.Command(fmt.Sprintf("kubectl delete solutions.solution.symphony %s -n %s", solutionName, namespace)).Run()
		if err != nil {
			return err
		}*/
	return nil
}

func VerifyPodNotExists() error {
	kubeClient, err := testhelpers.KubeClient()
	if err != nil {
		return err
	}
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	i := 0
	for {
		i++
		// List all pods in the namespace
		pods, err := kubeClient.CoreV1().Pods(namespace+"scope").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		// Verify that the pods we expect are running
		toNotFind := []string{"instance02"}

		Found := make(map[string]bool)
		for _, s := range toNotFind {
			found := false
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, s) {
					found = true
					break
				}
			}

			if found {
				Found[s] = true
			}
		}

		if len(Found) == 0 {
			fmt.Println("All pods are cleaned up!")
			break
		} else {
			time.Sleep(time.Second * 5)

			if i%12 == 0 {
				fmt.Printf("Waiting for pods to disappear: %v\n", Found)
			}
		}
	}

	return nil
}

// Clean up
func Cleanup() {
	localenvCmd(fmt.Sprintf("dumpSymphonyLogsForTest '%s'", TEST_NAME), "")
	localenvCmd("destroy all", "")
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string, flavor string) error {
	return shellExec(fmt.Sprintf("cd ../../../localenv && mage %s %s", mageCmd, flavor))
}

func shellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func writeYamlStringsToFile(yamlString string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(yamlString))
	if err != nil {
		return err
	}

	return nil
}
