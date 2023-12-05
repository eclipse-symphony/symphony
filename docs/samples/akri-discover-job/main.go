/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func collectEnvVariable(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return ""
	}
	return val
}

func main() {
	// We expected several borker properties
	// AKRI_CONFIG_NAME: Akri configuration name
	// AKRI_INSTNACE_NAME: Akri instance name
	// AKRI_INSTANCE_NAMESPACE: Akri instance namespace
	// AKRI_DEVICE_TYPE: Akri device type
	akriConfigName, present := os.LookupEnv("AKRI_CONFIG_NAME")
	if !present {
		panic("environment variable AKRI_CONFIG_NAME is not set")
	}
	akriNamespace, present := os.LookupEnv("AKRI_INSTANCE_NAMESPACE")
	if !present {
		panic("environment variable AKRI_INSTANCE_NAMESPACE is not set")
	}
	akriDeviceType, present := os.LookupEnv("AKRI_DEVICE_TYPE")
	if !present {
		panic("environment variable AKRI_DEVICE_TYPE is not set")
	}

	// Collect Akri properties
	akriProperties := make(map[string]interface{})
	hashTarget := ""
	if s := collectEnvVariable("DEBUG_ECHO_DESCRIPTION"); s != "" {
		akriProperties["DEBUG_ECHO_DESCRIPTION"] = s
		hashTarget = hashTarget + ":" + s
	}
	if s := collectEnvVariable("ONVIF_DEVICE_SERVICE_URL"); s != "" {
		akriProperties["ONVIF_DEVICE_SERVICE_URL"] = s
		hashTarget = hashTarget + ":" + s
	}
	if s := collectEnvVariable("ONVIF_DEVICE_IP_ADDRESS"); s != "" {
		akriProperties["ONVIF_DEVICE_IP_ADDRESS"] = s
		//TODO: THis is temporary test code
		akriProperties["RTSP_URL"] = "rtsp://" + s + ":554"
		hashTarget = hashTarget + ":" + s
	}
	if s := collectEnvVariable("ONVIF_DEVICE_MAC_ADDRESS"); s != "" {
		akriProperties["ONVIF_DEVICE_MAC_ADDRESS"] = s
		hashTarget = hashTarget + ":" + s
	}
	if s := collectEnvVariable("OPCUA_DISCOVERY_URL"); s != "" {
		akriProperties["OPCUA_DISCOVERY_URL"] = s
		hashTarget = hashTarget + ":" + s
	}
	if s := collectEnvVariable("UDEV_DEVNODE"); s != "" {
		akriProperties["UDEV_DEVNODE"] = s
		hashTarget = hashTarget + ":" + s
	}

	inCluster := collectEnvVariable("IN_CLUSTER") == "true"

	var kConfig *rest.Config
	var err error
	var client dynamic.Interface
	if inCluster {
		kConfig, err = rest.InClusterConfig()
	} else {
		if home := homedir.HomeDir(); home != "" {
			configPath := filepath.Join(home, ".kube", "config")
			kConfig, err = clientcmd.BuildConfigFromFlags("", configPath)
		} else {
			panic("failed to locate home folder")
		}
	}

	if err != nil {
		panic("failed to create Kubernetes config: " + err.Error())
	}
	client, err = dynamic.NewForConfig(kConfig)
	if err != nil {
		panic("failed to create Kubernetes client: " + err.Error())
	}
	id := akriConfigName + "-" + fmt.Sprintf("%x", sha256.Sum256([]byte(hashTarget)))[:8]

	resource := schema.GroupVersionResource{
		Group:    "devices.fabric.symphony",
		Version:  "v1",
		Resource: "devices",
	}
	existing, err := client.Resource(resource).Namespace(akriNamespace).Get(context.Background(), id, metav1.GetOptions{})

	if !(err == nil && existing.Object["metadata"].(map[string]interface{})["name"].(string) == id) {
		device := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"akri-config": akriConfigName,
					},
					"name":      id,
					"namespace": akriNamespace,
				},
				"spec": map[string]interface{}{
					"ref": map[string]interface{}{
						"id":         id,
						"registry":   "Akri",
						"properties": akriProperties,
					},
					"type": akriDeviceType,
				},
			},
		}
		device.SetGroupVersionKind(schema.GroupVersionKind{ //Watch AKri instances
			Kind:    "Device",
			Group:   "devices.fabric.symphony",
			Version: "v1",
		})

		_, err = client.Resource(resource).Namespace(akriNamespace).Create(context.Background(), device, metav1.CreateOptions{})
		if err != nil {
			panic("failed to create Device: " + err.Error())
		}
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		os.Exit(0)
	}()
	for {
		time.Sleep(10 * time.Second) // or runtime.Gosched() or similar per @misterbee
	}
}
