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
	"strings"

	coacontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ConstructActivityAndDiagnosticContextFromAnnotations(namespace string, objectId string, diagnosticResourceId string, diagnosticResourceLocation string, annotations map[string]string, operationName string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) context.Context {
	retCtx := ConstructDiagnosticContextFromAnnotations(annotations, ctx, logger)
	retCtx = ConstructActivityContextFromAnnotations(namespace, objectId, diagnosticResourceId, diagnosticResourceLocation, annotations, operationName, k8sClient, retCtx, logger)
	return retCtx
}

func ConstructDiagnosticContextFromAnnotations(annotations map[string]string, ctx context.Context, logger logr.Logger) context.Context {
	correlationId := annotations[constants.AzureCorrelationIdKey]
	resourceId := annotations[constants.AzureResourceIdKey]
	retCtx := coacontexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)
	return retCtx
}

func ConstructActivityContextFromAnnotations(namespace string, objectId string, diagnosticResourceId string, diagnosticResourceLocation string, annotations map[string]string, operationName string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) context.Context {
	correlationId := annotations[constants.AzureCorrelationIdKey]
	resourceId := annotations[constants.AzureResourceIdKey]
	resourceCloudLocation := annotations[constants.AzureLocationKey]
	systemData := annotations[constants.AzureSystemDataKey]
	edgeLocation := annotations[constants.AzureEdgeLocationKey]

	resourceK8SId := objectId
	callerId := ""
	if systemData != "" {
		systemDataMap := make(map[string]string)
		if err := json.Unmarshal([]byte(systemData), &systemDataMap); err != nil {
			InfoWithCtx(logger, ctx, "Failed to unmarshal system data", "error", err)
		} else {
			callerId = systemDataMap[constants.AzureCreatedByKey]
		}
	}

	retCtx := coacontexts.PatchActivityLogContextToCurrentContext(coacontexts.NewActivityLogContext(diagnosticResourceId, diagnosticResourceLocation, resourceId, resourceCloudLocation, edgeLocation, operationName, correlationId, callerId, resourceK8SId), ctx)
	return retCtx
}

func decorateLogWithCtx(l logr.Logger, ctx context.Context, folding bool) logr.Logger {
	var diagCtx *coacontexts.DiagnosticLogContext
	var ok bool
	if ctx != nil {
		diagCtx, ok = ctx.(*coacontexts.DiagnosticLogContext)
		if !ok {
			diagCtx, ok = ctx.Value(coacontexts.DiagnosticLogContextKey).(*coacontexts.DiagnosticLogContext)
		}
		if ok && diagCtx != nil {
			if folding {
				l = l.WithValues(string(coacontexts.DiagnosticLogContextKey), diagCtx)
			} else {
				l = l.WithValues(
					coacontexts.OTEL_Diagnostics_CorrelationId, diagCtx.GetCorrelationId(),
					coacontexts.OTEL_Diagnostics_ResourceCloudId, strings.ToUpper(diagCtx.GetResourceId()),
				)
				traceCtxJson, err := json.Marshal(diagCtx.GetTraceContext())
				if err != nil {
					l = l.WithValues(coacontexts.OTEL_Diagnostics_TraceContext, diagCtx.GetTraceContext())
				} else {
					l = l.WithValues(coacontexts.OTEL_Diagnostics_TraceContext, string(traceCtxJson))
				}
			}
		}
	}
	return l
}

func InfoWithCtx(l logr.Logger, ctx context.Context, msg string, keysAndValues ...interface{}) {
	l = decorateLogWithCtx(l, ctx, true)
	l.WithCallDepth(1).Info(msg, keysAndValues...)
}

func ErrorWithCtx(l logr.Logger, ctx context.Context, err error, msg string, keysAndValues ...interface{}) {
	l = decorateLogWithCtx(l, ctx, true)
	l.WithCallDepth(1).Error(err, msg, keysAndValues...)
}
