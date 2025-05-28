/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package dynamicclient

import (
	"context"
	"fmt"
	"gopls-workspace/utils/diagnostic"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var dynamicClient dynamic.Interface
var dynamicclientlog = logf.Log.WithName("dynamicclient")

func SetClient(config *rest.Config) error {
	var err error
	// to do: read this from config
	config.QPS = 300
	config.Burst = 300
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func Get(ctx context.Context, resourceType validation.ResourceType, name string, namespace string) (*unstructured.Unstructured, error) {
	resource, err := switchResourceType(resourceType)
	if err != nil {
		diagnostic.ErrorWithCtx(dynamicclientlog, ctx, err, fmt.Sprintf("Unsupported resourceType %s ", resourceType))
		return nil, err
	}
	obj, err := resource.Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		diagnostic.ErrorWithCtx(dynamicclientlog, ctx, err, fmt.Sprintf("Failed to get %s %s in namespace %s with error %s", resourceType, name, namespace, err.Error()))
		return nil, err
	}
	return obj, nil
}

func ListWithLabels(ctx context.Context, resourceType validation.ResourceType, namespace string, labels map[string]string, count int64) (*unstructured.UnstructuredList, error) {
	resource, err := switchResourceType(resourceType)
	if err != nil {
		diagnostic.ErrorWithCtx(dynamicclientlog, ctx, err, fmt.Sprintf("Unsupported resourceType %s ", resourceType))
		return nil, err
	}
	listOption := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels)),
	}
	if count > 0 {
		listOption.Limit = count
	}
	list, err := resource.Namespace(namespace).List(context.Background(), listOption)
	if err != nil {
		diagnostic.ErrorWithCtx(dynamicclientlog, ctx, err, fmt.Sprintf("Failed to list %s in namespace %s with error %s", resourceType, namespace, err.Error()))
		return nil, err
	}
	return list, nil
}

func GetObjectWithUniqueName(ctx context.Context, resourceType validation.ResourceType, displayName string, namespace string) (*unstructured.Unstructured, error) {
	objectList, err := ListWithLabels(ctx, resourceType, namespace, map[string]string{api_constants.DisplayName: utils.ConvertStringToValidLabel(displayName)}, 1)
	if err != nil {
		// return true if List call failed
		return nil, err
	}
	if len(objectList.Items) > 0 {
		return &objectList.Items[0], nil
	}
	return nil, v1alpha2.NewCOAError(nil, string(resourceType)+" not found", v1alpha2.NotFound)
}

func switchResourceType(resourceType validation.ResourceType) (dynamic.NamespaceableResourceInterface, error) {
	group, version, resource, _ := validation.GetResourceMetadata(resourceType)
	return dynamicClient.Resource(schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}), nil
}
