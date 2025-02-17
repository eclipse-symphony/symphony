//go:build !azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helper

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetInstanceTargetName(name string) string {
	return name
}

func GetInstanceRootResource(name string) string {
	return ""
}

func GetInstanceOwnerReferences(apiClient api_utils.ApiClient, ctx context.Context, objectName string, objectNamespace string, instanceState model.InstanceState, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}

func GetSolutionContainerOwnerReferences(apiClient api_utils.ApiClient, ctx context.Context, objectName string, objectNamespace string, user string, pwd string) ([]metav1.OwnerReference, error) {
	return nil, nil
}
