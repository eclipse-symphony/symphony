/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reconcilers

import (
	"context"
	"gopls-workspace/utils"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithFinalizerName(name string) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.finalizerName = name
	}
}

func WithClient(c client.Client) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.kubeClient = c
	}
}

func WithApiClient(c utils.ApiClient) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.apiClient = c
	}
}

func WithReconciliationInterval(d time.Duration) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.reconciliationInterval = d
	}
}

func WithPollInterval(d time.Duration) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.pollInterval = d
	}
}

func WithDeleteTimeOut(d time.Duration) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.deleteTimeOut = d
	}
}

func WithDeleteSyncDelay(d time.Duration) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.deleteSyncDelay = d
	}
}

func WithDelayFunc(f func(time.Duration)) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.delayFunc = f
	}
}

func WithDeploymentKeyResolver(f func(Reconcilable) string) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.deploymentKeyResolver = f
	}
}

func WithDeploymentErrorBuilder(f func(*model.SummaryResult, error, *model.ErrorType)) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.deploymentErrorBuilder = f
	}
}

func WithDeploymentBuilder(f func(ctx context.Context, object Reconcilable) (*model.DeploymentSpec, error)) DeploymentReconcilerOptions {
	return func(r *DeploymentReconciler) {
		r.deploymentBuilder = f
	}
}
