/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configutils

import (
	"context"
	"os"
	"strings"

	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	monitorv1 "gopls-workspace/apis/monitor/v1"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	data, err := os.ReadFile(namespaceFile)
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

func PopulateActivityAndDiagnosticsContextFromAnnotations(namespace string, objectId string, annotations map[string]string, operationName string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) context.Context {
	diagnosticResourceId, diagnosticResourceLocation := monitorv1.GetDiagnosticCloudResourceInfo(annotations, k8sClient, ctx, logger)
	retCtx := diagnostic.ConstructActivityAndDiagnosticContextFromAnnotations(namespace, objectId, diagnosticResourceId, diagnosticResourceLocation, annotations, operationName, k8sClient, ctx, logger)
	return retCtx
}
