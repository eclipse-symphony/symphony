/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"gopls-workspace/utils/diagnostic"
	"sync"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var globalDiagnosticRes *Diagnostic
var globalDiagnosticResLock *sync.RWMutex = &sync.RWMutex{}

func filterDiagnosticResourceByAnnotationWhenListInCluster(diagnostic *Diagnostic, annotationFilterFunc func(diagResourceAnnotations map[string]string) bool) bool {
	if diagnostic == nil {
		return false
	}
	if diagnostic.Annotations == nil {
		return false
	}
	return annotationFilterFunc(diagnostic.Annotations)
}

func getGlobalDiagnosticResourceInCluster(annotationFilterFunc func(diagResourceAnnotations map[string]string) bool, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (*Diagnostic, error) {
	// List all diagnostics resources under the namespace
	// If there is a resource matchs the filter, use that as the diagnostic resource
	// If there is no such resource, return nil
	diagnostics := &DiagnosticList{}
	err := k8sClient.List(ctx, diagnostics)
	if err != nil {
		diagnostic.InfoWithCtx(logger, ctx, "Failed to list diagnostics resources", "error", err)
		return nil, err
	}
	for _, d := range diagnostics.Items {
		if filterDiagnosticResourceByAnnotationWhenListInCluster(&d, annotationFilterFunc) {
			diagnostic.InfoWithCtx(logger, ctx, "Found global diagnostics resource", "name", d.Name, "namespace", d.Namespace)
			return &d, nil
		}
	}
	diagnostic.InfoWithCtx(logger, ctx, "No global diagnostics resource found")
	return nil, nil
}

func GetDiagnosticCustomResourceFromCache(sourceResourceAnnotations map[string]string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (*Diagnostic, error) {
	// firstly get from local cache, if not found, then get from cluster
	diagRes := ReadDiagnosticResourceFromCache()
	if diagRes == nil {
		diagResFromCluster, err := GetGlobalDiagnosticResourceInCluster(sourceResourceAnnotations, k8sClient, ctx, logger)
		if err != nil {
			diagnostic.ErrorWithCtx(logger, ctx, err, "Failed to get global diagnostic resource from cluster")
			return nil, err
		} else {
			// update the cache
			SetDiagnosticResourceCache(diagResFromCluster)
			return diagResFromCluster, nil
		}
	} else {
		diagnostic.InfoWithCtx(logger, ctx, "Found global diagnostics resource in cache", "name", diagRes.Name, "namespace", diagRes.Namespace)
		return diagRes, nil
	}
}

func ReadDiagnosticResourceFromCache() *Diagnostic {
	globalDiagnosticResLock.RLock()
	defer globalDiagnosticResLock.RUnlock()
	return globalDiagnosticRes
}

func SetDiagnosticResourceCache(diagnosticRes *Diagnostic) {
	globalDiagnosticResLock.Lock()
	defer globalDiagnosticResLock.Unlock()
	globalDiagnosticRes = diagnosticRes
}

func ClearDiagnosticResourceCache() {
	SetDiagnosticResourceCache(nil)
}
