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

func GetInstanceTargetName(name string) string {
	return name
}

func GetInstanceRootResource(name string) string {
	return ""
}

func GetInstanceOwnerReferences(apiClient ApiClient, ctx context.Context, objectName string, objectNamespace string, instanceState model.InstanceState, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GetSolutionContainerOwnerReferences(apiClient ApiClient, ctx context.Context, objectName string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GenerateSystemDataAnnotationsForInstanceHistory(annotations map[string]string, solutionId string) map[string]string {
	return annotations
}

func ConvertReferenceToObjectNameHelper(name string) string {
	if strings.Contains(name, constants.ReferenceSeparator) {
		name = strings.ReplaceAll(name, constants.ReferenceSeparator, constants.ResourceSeperator)
	}
	return name
}
