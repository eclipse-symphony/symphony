/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package diagnostic

import (
	"context"
	"encoding/json"
	"gopls-workspace/constants"

	coacontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ConstructActivityAndDiagnosticContextFromAnnotations(namespace string, objectId string, diagnosticResourceId string, annotations map[string]string, operationName string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) context.Context {
	retCtx := ConstructDiagnosticContextFromAnnotations(annotations, ctx, logger)
	retCtx = ConstructActivityContextFromAnnotations(namespace, objectId, diagnosticResourceId, annotations, operationName, k8sClient, retCtx, logger)
	return retCtx
}

func ConstructDiagnosticContextFromAnnotations(annotations map[string]string, ctx context.Context, logger logr.Logger) context.Context {
	correlationId := annotations[constants.AzureCorrelationIdKey]
	resourceId := annotations[constants.AzureResourceIdKey]
	retCtx := coacontexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)
	return retCtx
}

func ConstructActivityContextFromAnnotations(namespace string, objectId string, diagnosticResourceId string, annotations map[string]string, operationName string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) context.Context {
	correlationId := annotations[constants.AzureCorrelationIdKey]
	resourceId := annotations[constants.AzureResourceIdKey]
	location := annotations[constants.AzureLocationKey]
	systemData := annotations[constants.AzureSystemDataKey]
	edgeLocation := annotations[constants.AzureEdgeLocationKey]

	resourceK8SId := objectId
	callerId := ""
	if systemData != "" {
		systemDataMap := make(map[string]string)
		if err := json.Unmarshal([]byte(systemData), &systemDataMap); err != nil {
			logger.Info("Failed to unmarshal system data", "error", err)
		} else {
			callerId = systemDataMap[constants.AzureCreatedByKey]
		}
	}

	retCtx := coacontexts.PatchActivityLogContextToCurrentContext(coacontexts.NewActivityLogContext(diagnosticResourceId, resourceId, location, edgeLocation, operationName, correlationId, callerId, resourceK8SId), ctx)
	return retCtx
}
