//go:build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetInstanceName(solutionContainerName, objectName string) string {
	return objectName
}

func GetSolutionAndContainerName(name string) (string, string) {
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	parts = strings.Split(name, "-v-")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func GetInstanceTargetName(name string) string {
	return name
}

func GetInstanceRootResource(name string) string {
	return ""
}

func GetInstanceOwnerReferences(apiClient ApiClient, ctx context.Context, solutionContainer string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GetInstanceOwnerReferencesV1(apiClient ApiClient, ctx context.Context, objectName string, objectNamespace string, instanceState model.InstanceState, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GetSolutionContainerOwnerReferences(apiClient ApiClient, ctx context.Context, objectName string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GenerateSystemDataAnnotations(ctx context.Context, annotations map[string]string, solutionId string) map[string]string {
	return annotations
}

func ConvertReferenceToObjectNameHelper(name string) string {
	if strings.Contains(name, constants.ReferenceSeparator) {
		name = strings.ReplaceAll(name, constants.ReferenceSeparator, constants.ResourceSeperator)
	}
	return name
}

func GenerateOperationId() string {
	return ""
}
