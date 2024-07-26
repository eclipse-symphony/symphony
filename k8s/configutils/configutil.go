/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configutils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/constants"

	coacontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

var (
	namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	configName    = os.Getenv(constants.ConfigName)
)

func GetValidationPoilicies() (map[string][]configv1.ValidationPolicy, error) {
	myConfig, err := GetProjectConfig()
	if err != nil {
		return nil, err
	}

	return myConfig.ValidationPolicies, nil
}

func GetProjectConfig() (*configv1.ProjectConfig, error) {
	// home := homedir.HomeDir()
	// // use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // create the clientset
	// clientset, err := kubernetes.NewForConfig(config)

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespace, err := getNamespace()
	if err != nil {
		return nil, err
	}

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), configName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var myConfig configv1.ProjectConfig
	data := configMap.Data["controller_manager_config.yaml"]
	err = yaml.Unmarshal([]byte(data), &myConfig)
	if err != nil {
		return nil, err
	}

	return &myConfig, nil
}

func getNamespace() (string, error) {
	// read the namespace from the file
	data, err := ioutil.ReadFile(namespaceFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func CheckValidationPack(myName string, myValue, validationType string, pack []configv1.ValidationStruct) (string, error) {
	if validationType == "unique" {
		for _, p := range pack {
			if p.Field == myValue {
				if myName != p.Name {
					return myValue, nil
				}
			}
		}
	}
	return "", nil
}

func CheckOwnerReferenceAlreadySet(existingRefs []metav1.OwnerReference, ownerRefToCheck metav1.OwnerReference) bool {
	for _, r := range existingRefs {
		if areSameOwnerReferences(ownerRefToCheck, r) {
			return true
		}
	}
	return false
}

// Returns true if a and b point to the same object
func areSameOwnerReferences(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}

func ValidateObjectName(name string, rootResource string) *field.Error {
	if rootResource == "" {
		return field.Invalid(field.NewPath("spec").Child("rootResource"), rootResource, "rootResource must be a non-empty string")
	}

	if !strings.HasPrefix(name, rootResource) {
		return field.Invalid(field.NewPath("name"), name, "name must start with spec.rootResource")
	}

	prefix := rootResource + constants.ResourceSeperator
	remaining := strings.TrimPrefix(name, prefix)

	if remaining == name {
		return field.Invalid(field.NewPath("name"), name, "name should be in the format '<rootResource>-v-<version>'")
	}

	if strings.Contains(remaining, constants.ResourceSeperator) || strings.HasPrefix(remaining, "v-") {
		return field.Invalid(field.NewPath("name"), name, "name should be in the format <rootResource>-v-<version> where <version> does not contain '-v-' or starts with 'v-'")
	}

	return nil
}

func PopulateActivityAndDiagnosticsContextFromAnnotations(objectId string, annotations map[string]string, activityCategory string, operationName string, ctx context.Context, log logr.Logger) context.Context {
	correlationId := annotations[constants.AzureCorrelationIdKey]
	resourceId := annotations[constants.AzureResourceIdKey]
	location := annotations[constants.AzureLocationKey]
	systemData := annotations[constants.AzureSystemDataKey]

	// correlationId := uuid.New().String()
	// resourceId := objectId
	// location := "on-premise"
	// systemData := "{\"createdBy\":\"On-Premise\"}"

	resourceK8SId := objectId
	callerId := ""
	if systemData != "" {
		systemDataMap := make(map[string]string)
		if err := json.Unmarshal([]byte(systemData), &systemDataMap); err != nil {
			log.Info("Failed to unmarshal system data", "error", err)
		} else {
			// callerId = systemDataMap[constants.AzureCreatedByKey]
			callerId = "******"
		}
	}
	retCtx := coacontexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)
	retCtx = coacontexts.PatchActivityLogContextToCurrentContext(coacontexts.NewActivityLogContext(resourceId, location, operationName, activityCategory, correlationId, callerId, resourceK8SId), retCtx)
	return retCtx
}
