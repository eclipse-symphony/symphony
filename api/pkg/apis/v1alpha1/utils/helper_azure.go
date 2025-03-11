//go:build azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AzureSolutionVersionIdPattern = "^/subscriptions/([0-9a-fA-F-]+)/resourcegroups/([^/]+)/providers/([^/]+)/targets/([^/]+)/solutions/([^/]+)/versions/([^/]+)$"
	AzureTargetIdPattern          = "^/subscriptions/([0-9a-fA-F-]+)/resourcegroups/([^/]+)/providers/([^/]+)/targets/([^/]+)$"
)

func ConvertAzureSolutionVersionReferenceToObjectName(name string) (string, bool) {
	log.Infof("Azure: convert solution version reference to object name: %s", name)
	r := regexp.MustCompile(AzureSolutionVersionIdPattern)
	if !r.MatchString(name) {
		return "", false
	}
	return r.ReplaceAllString(name, fmt.Sprintf("$4%s$5%s$6", constants.ResourceSeperator, constants.ResourceSeperator)), true
}

func ConvertAzureTargetReferenceToObjectName(name string) (string, bool) {
	log.Infof("Azure: convert target reference to object name: %s", name)
	r := regexp.MustCompile(AzureTargetIdPattern)
	if !r.MatchString(name) {
		return "", false
	}
	return r.ReplaceAllString(name, "$4"), true
}

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

func GetInstanceOwnerReferences(apiClient ApiClient, ctx context.Context, solutionContainer string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
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

func GetSolutionContainerOwnerReferences(apiClient ApiClient, ctx context.Context, objectName string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
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

func GenerateSystemDataAnnotations(annotations map[string]string, resourceId string) map[string]string {
	log.Infof("Azure: check if annotation need to be added: %v", annotations)
	if isPrivateResourceProvider(resourceId) {
		annotations[constants.AzureSystemDataKey] = `{"clientLocation":"eastus2euap"}`
	}
	return annotations
}

func isPrivateResourceProvider(resourceId string) bool {
	pattern := `^/subscriptions/([0-9a-fA-F-]+)/resourcegroups/([^/]+)/providers/private.edge/.*`
	re := regexp.MustCompile(pattern)
	return re.MatchString(strings.ToLower(resourceId))
}

func ConvertReferenceToObjectNameHelper(name string) string {
	// deal with Azure pattern
	if strings.Contains(name, constants.ReferenceSeparator) {
		name = strings.ReplaceAll(name, constants.ReferenceSeparator, constants.ResourceSeperator)
		return name
	}
	if n, ok := ConvertAzureSolutionVersionReferenceToObjectName(name); ok {
		return n
	}
	if n, ok := ConvertAzureTargetReferenceToObjectName(name); ok {
		return n
	}
	return name
}
