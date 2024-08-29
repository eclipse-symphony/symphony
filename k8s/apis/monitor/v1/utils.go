/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"gopls-workspace/constants"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDiagnosticResourceId(namespace string, edgeLocation string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) string {
	// List all diagnostics resources under the namespace
	// If there is a resource with the same namespace and edgeLocation, use that as the diagnostic resource
	// If there is no such resource, return empty string
	diagnostic, err := GetDiagnosticCustomResource(namespace, edgeLocation, k8sClient, ctx, logger)
	if err != nil {
		logger.Info("Failed to get diagnostic resource", "error", err)
		return ""
	}
	if diagnostic != nil {
		return diagnostic.Annotations[constants.AzureResourceIdKey]
	}
	return ""
}

func GetDiagnosticCustomResource(namespace string, edgeLocation string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (*Diagnostic, error) {
	// List all diagnostics resources under the namespace
	// If there is a resource with the same namespace and edgeLocation, use that as the diagnostic resource
	// If there is no such resource, return nil
	diagnostics := &DiagnosticList{}
	err := k8sClient.List(ctx, diagnostics, client.InNamespace(namespace))
	if err != nil {
		logger.Info("Failed to list diagnostics resources", "error", err)
		return nil, err
	}
	for _, diagnostic := range diagnostics.Items {
		if diagnostic.Annotations[constants.AzureEdgeLocationKey] == edgeLocation {
			return &diagnostic, nil
		}
	}
	logger.Info("No diagnostics resource found for the namespace and edge location", "namespace", namespace, "edgeLocation", edgeLocation)
	return nil, nil
}
