//go:build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helper

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetInstanceName(solutionContainerName, objectName string) string {
	return fmt.Sprintf("%s-v-%s", solutionContainerName, objectName)
}

func GetInstanceTargetName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return name
	}
	return parts[len(parts)-1]
}

func GetSolutionAndContainerName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) < 5 {
		return "", ""
	}
	container := fmt.Sprintf("%s-v-%s", parts[len(parts)-5], parts[len(parts)-3])
	return container, parts[len(parts)-1]
}

func GetInstanceRootResource(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-3]
}

func GenerateOperationId() string {
	return uuid.New().String()
}
func GetInstanceOwnerReferences(apiClient api_utils.ApiClient, ctx context.Context, solutionContainer string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	sc, err := apiClient.GetSolutionContainer(ctx, solutionContainer, objectNamespace, user, pwd)
	if err != nil {
		return nil, err
	}
	return []metav1.OwnerReference{
		{
			APIVersion: fmt.Sprintf("%s/%s", model.SolutionGroup, "v1"),
			Kind:       "SolutionContainer",
			Name:       sc.ObjectMeta.Name,
			UID:        sc.ObjectMeta.UID,
		},
	}, nil
}

func GetSolutionContainerOwnerReferences(apiClient api_utils.ApiClient, ctx context.Context, objectName string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	target, err := apiClient.GetTarget(ctx, objectName, objectNamespace, user, pwd)
	if err != nil {
		return nil, err
	}

	return []metav1.OwnerReference{
		{
			APIVersion: fmt.Sprintf("%s/%s", model.FabricGroup, "v1"),
			Kind:       "Target",
			Name:       target.ObjectMeta.Name,
			UID:        target.ObjectMeta.UID,
		},
	}, nil
}

func GenerateSystemDataAnnotations(annotations map[string]string) map[string]string {
	if isPrivateResourceProvider(annotations[constants.AzureResourceIdKey]) {
		annotations[constants.AzureSystemDataKey] = `{"clientLocation":"eastus2euap"}`
	}
	return annotations
}

func isPrivateResourceProvider(resourceId string) bool {
	pattern := `^/subscriptions/([0-9a-fA-F-]+)/resourcegroups/([^/]+)/providers/private.edge/.*`
	re := regexp.MustCompile(pattern)
	return re.MatchString(strings.ToLower(resourceId))
}
