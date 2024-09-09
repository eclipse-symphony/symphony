/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	k8smodel "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/utils"

	"gopls-workspace/utils/diagnostic"
	utilsmodel "gopls-workspace/utils/model"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/go-logr/logr"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type (
	patchStatusOptions struct {
		deploymentQueued bool
		nonTerminalErr   error
		terminalErr      error
	}
	DeploymentReconciler struct {
		finalizerName          string
		kubeClient             client.Client
		apiClient              utils.ApiClient
		reconciliationInterval time.Duration
		pollInterval           time.Duration
		deleteTimeOut          time.Duration
		applyTimeOut           time.Duration // TODO: Use reconciliation policy to configure
		deleteSyncDelay        time.Duration // TODO: Use operator reconcile loop instead of this delay
		delayFunc              func(time.Duration)
		deploymentKeyResolver  func(Reconcilable) string
		deploymentErrorBuilder func(*apimodel.SummaryResult, error, *apimodel.ErrorType)
		deploymentBuilder      func(ctx context.Context, object Reconcilable) (*apimodel.DeploymentSpec, error)
	}
	DeploymentReconcilerOptions func(*DeploymentReconciler)
	ReconcilerSubject           string
)

const (
	defaultTimeout                = 15 * time.Minute
	defaultReconciliationInterval = 30 * time.Minute
	defaultPollInterval           = 10 * time.Second
)

var (
	_             Reconciler = &DeploymentReconciler{}
	termialErrors            = map[v1alpha2.State]struct{}{
		v1alpha2.TimedOut:               {},
		v1alpha2.TargetPropertyNotFound: {},
	}
)

func NewDeploymentReconciler(opts ...DeploymentReconcilerOptions) (*DeploymentReconciler, error) {
	r := &DeploymentReconciler{
		deploymentKeyResolver:  defaultDeploymentKeyResolver,
		deploymentErrorBuilder: defaultProvisioningErrorBuilder,
		delayFunc:              time.Sleep,
		applyTimeOut:           defaultTimeout,
		reconciliationInterval: defaultReconciliationInterval,
		pollInterval:           defaultPollInterval,
		deleteTimeOut:          defaultTimeout,
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.finalizerName == "" {
		return nil, fmt.Errorf("finalizer name cannot be empty")
	}
	if r.kubeClient == nil {
		return nil, fmt.Errorf("kube client cannot be nil")
	}
	if r.apiClient == nil {
		return nil, fmt.Errorf("api client cannot be nil")
	}
	if r.deploymentBuilder == nil {
		return nil, fmt.Errorf("deployment builder cannot be nil")
	}
	return r, nil
}

func (r *DeploymentReconciler) deriveReconcileInterval(log logr.Logger, ctx context.Context, target Reconcilable) (reconciliationInterval, timeout time.Duration) {
	rp := target.GetReconciliationPolicy()
	reconciliationInterval = r.reconciliationInterval
	timeout = r.applyTimeOut
	if rp != nil {
		// reconciliationPolicy is set, use the interval if it's active
		if rp.State.IsActive() {
			// periodic reconciliation, interval is set
			if rp.Interval != nil {
				interval, err := time.ParseDuration(*rp.Interval)
				if err != nil {
					diagnostic.ErrorWithCtx(log, ctx, err, fmt.Sprintf("failed to parse reconciliation interval %s, using default %s", *rp.Interval, reconciliationInterval))
					return
				}
				reconciliationInterval = interval
			}
		}
		if rp.State.IsInActive() {
			// only reconcile once
			reconciliationInterval = 0
		}

	}
	// no reconciliationPolicy configured or reconciliationPolicy.state is invalid, use default reconciliation interval: r.reconciliationInterval
	return
}

func (r *DeploymentReconciler) populateDiagnosticsAndActivitiesFromAnnotations(ctx context.Context, object Reconcilable, operationName string, k8sClient client.Reader, log logr.Logger) context.Context {
	diagnostic.InfoWithCtx(log, ctx, "Populating diagnostics and activities from annotations")
	if object == nil {
		return ctx
	}
	annotations := object.GetAnnotations()
	if annotations == nil {
		return ctx
	}
	resourceK8SId := object.GetNamespace() + "/" + object.GetName()
	return configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(object.GetNamespace(), resourceK8SId, annotations, operationName, k8sClient, ctx, log)
}

// attemptUpdate attempts to update the instance
func (r *DeploymentReconciler) AttemptUpdate(ctx context.Context, object Reconcilable, isRemoval bool, log logr.Logger, operationStartTimeKey string, operationName string) (metrics.OperationStatus, reconcile.Result, error) {
	// populate diagnostics and activities from annotations
	ctx = r.populateDiagnosticsAndActivitiesFromAnnotations(ctx, object, operationName, r.kubeClient, log)
	if !controllerutil.ContainsFinalizer(object, r.finalizerName) && !isRemoval {
		controllerutil.AddFinalizer(object, r.finalizerName)
		// updates the object in Kubernetes cluster
		if err := r.kubeClient.Update(ctx, object); err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to add finalizer to object")
			return metrics.StatusUpdateFailed, ctrl.Result{}, err
		}
	}
	// Get reconciliation interval
	reconciliationInterval, timeout := r.deriveReconcileInterval(log, ctx, object)

	if isRemoval {
		// Timeout will be deletion timestamp + delete timeout duration
		timeout := object.GetDeletionTimestamp().Time.Add(r.deleteTimeOut)

		if metav1.Now().Time.After(timeout) {
			return metrics.DeploymentTimedOut, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
		}
	} else {
		if object.GetAnnotations()[operationStartTimeKey] == "" || utilsmodel.IsTerminalState(object.GetStatus().ProvisioningStatus.Status) {
			r.patchOperationStartTime(object, operationStartTimeKey)
			if err := r.kubeClient.Update(ctx, object); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status with timeout error")
				return metrics.StatusUpdateFailed, ctrl.Result{}, err
			}
		}
		// If the object hasn't reached a terminal state and the time since the operation started is greater than the
		// apply timeout, we should update the status with a terminal error and return
		startTime, err := time.Parse(time.RFC3339, object.GetAnnotations()[operationStartTimeKey])
		if err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to parse operation start time")
			return metrics.StatusUpdateFailed, ctrl.Result{}, err
		}
		if time.Since(startTime) > timeout {
			diagnostic.InfoWithCtx(log, ctx, "Requeueing after timeout", "requeueAfter", reconciliationInterval)
			return metrics.DeploymentTimedOut, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
		}
	}

	if err := r.queueDeploymentJob(ctx, object, isRemoval, true, operationStartTimeKey); err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to queue deployment job")
		return r.handleDeploymentError(ctx, object, nil, isRemoval, reconciliationInterval, err, log)
	}

	diagnostic.InfoWithCtx(log, ctx, "Updating object status with deployment queued")
	if _, err := r.updateObjectStatus(ctx, object, nil, patchStatusOptions{deploymentQueued: true}, log); err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status with deployment queued")
		return metrics.StatusUpdateFailed, ctrl.Result{}, err
	}

	if isRemoval || reconciliationInterval == 0 {
		// remove, don't need to reconcile
		// If the reconcile policy is once (interval == 0 or state==inactive), we should not queue a new job and return
		diagnostic.InfoWithCtx(log, ctx, "Won't queue a new job because of reconcile policy or remove", "isRemoval", isRemoval)
		return metrics.DeploymentQueued, ctrl.Result{}, nil
	} else {
		diagnostic.InfoWithCtx(log, ctx, "Queueing deployment job because of reconciliation interval", "requeueAfter", reconciliationInterval)
		return metrics.DeploymentQueued, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
	}
}

func (r *DeploymentReconciler) PollingResult(ctx context.Context, object Reconcilable, isRemoval bool, log logr.Logger, operationStartTimeKey string, operationName string) (metrics.OperationStatus, reconcile.Result, error) {
	// populate diagnostics and activities from annotations
	ctx = r.populateDiagnosticsAndActivitiesFromAnnotations(ctx, object, operationName, r.kubeClient, log)
	// Get reconciliation interval
	reconciliationInterval, timeout := r.deriveReconcileInterval(log, ctx, object)
	if isRemoval {
		// Timeout will be deletion timestamp + delete timeout duration
		timeout := object.GetDeletionTimestamp().Time.Add(r.deleteTimeOut)

		if metav1.Now().Time.After(timeout) {
			diagnostic.InfoWithCtx(log, ctx, "Operation timed out", "timeout", r.deleteTimeOut)
			r.updateObjectStatus(ctx, object, nil, patchStatusOptions{
				terminalErr: v1alpha2.NewCOAError(nil, "failed to completely delete the resource within the allocated time", v1alpha2.TimedOut),
			}, log)
			return metrics.DeploymentTimedOut, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
		}
	} else {
		if object.GetAnnotations()[operationStartTimeKey] == "" {
			return metrics.DeploymentPolling, ctrl.Result{RequeueAfter: r.pollInterval}, nil
		}
		// If the object hasn't reached a terminal state and the time since the operation started is greater than the
		// apply timeout, we should update the status with a terminal error and return
		startTime, err := time.Parse(time.RFC3339, object.GetAnnotations()[operationStartTimeKey])
		if err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to parse operation start time")
			return metrics.StatusUpdateFailed, ctrl.Result{}, err
		}
		if time.Since(startTime) > timeout {
			diagnostic.InfoWithCtx(log, ctx, "Failed to completely poll within the allocated time.", "timeout", timeout)
			if _, err := r.updateObjectStatus(ctx, object, nil, patchStatusOptions{
				terminalErr: v1alpha2.NewCOAError(nil, "failed to completely reconcile within the allocated time", v1alpha2.TimedOut),
			}, log); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status with timeout error")
				return metrics.StatusUpdateFailed, ctrl.Result{}, err
			}
			diagnostic.InfoWithCtx(log, ctx, "Repollinging after timeout", "requeueAfter", reconciliationInterval)
			return metrics.DeploymentTimedOut, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
		}
	}

	summary, err := r.getDeploymentSummary(ctx, object)
	if err != nil {
		// If the error is anything but 404, we should return the error so the reconciler can retry
		if !v1alpha2.IsNotFound(err) {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to get deployment summary")
			// updates the object status to reconciling
			if _, err := r.updateObjectStatus(ctx, object, summary, patchStatusOptions{
				nonTerminalErr: err,
			}, log); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status with non-terminal error")
				return metrics.StatusUpdateFailed, ctrl.Result{RequeueAfter: r.pollInterval}, err
			}
			diagnostic.InfoWithCtx(log, ctx, "Repolling after failed to get deployment summary")
			return metrics.GetDeploymentSummaryFailed, ctrl.Result{RequeueAfter: r.pollInterval}, err
		} else {
			// It's not found in api so we should mark as reconciling, queue a job and check back in POLL seconds
			diagnostic.InfoWithCtx(log, ctx, "Deployment summary not found, queueing job")
			return metrics.DeploymentQueued, ctrl.Result{RequeueAfter: r.pollInterval}, nil
		}
	}

	if !r.hasParity(ctx, object, summary, log) || (reconciliationInterval != 0 && time.Since(summary.Time) > reconciliationInterval) {
		diagnostic.InfoWithCtx(log, ctx, "Looks like queueing is pending will check after polling interval", "requeueAfter", r.pollInterval)
		return metrics.DeploymentQueued, ctrl.Result{RequeueAfter: r.pollInterval}, nil
	}

	switch summary.State {
	case apimodel.SummaryStatePending:
		// do nothing and check back in POLL seconds
		diagnostic.InfoWithCtx(log, ctx, "Deployment is pending, checking after poll interval", "requeueAfter", r.pollInterval)
		return metrics.StatusNoOp, ctrl.Result{RequeueAfter: r.pollInterval}, nil
	case apimodel.SummaryStateRunning:
		// if there is a parity mismatch between the object and the summary, the api is probably busy reconciling
		// a previous revision, so we'll only make sure the status is Non-terminal
		// But if they are the same, it's currently reconciling this generatation
		// we'll update the status and also the current progress. Either way, we'll check back in POLL seconds
		diagnostic.InfoWithCtx(log, ctx, "Updating object status when deployment is running")
		if _, err := r.updateObjectStatus(ctx, object, summary, patchStatusOptions{}, log); err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status")
			return metrics.StatusUpdateFailed, ctrl.Result{}, err
		}
		diagnostic.InfoWithCtx(log, ctx, "Deployment is running, checking after poll interval", "requeueAfter", r.pollInterval)
		return metrics.DeploymentStatusPolled, ctrl.Result{RequeueAfter: r.pollInterval}, nil
	case apimodel.SummaryStateDone:
		diagnostic.InfoWithCtx(log, ctx, "Updating object status when deployment is done")
		if _, err := r.updateObjectStatus(ctx, object, summary, patchStatusOptions{}, log); err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status")
			return metrics.StatusUpdateFailed, ctrl.Result{}, err
		}
		if object.GetStatus().ProvisioningStatus.Status == string(utilsmodel.ProvisioningStatusFailed) {
			diagnostic.InfoWithCtx(log, ctx, "provisioning status is failed")
			return metrics.DeploymentTimedOut, ctrl.Result{RequeueAfter: reconciliationInterval}, nil
		}
		// TODO: handle crash consistency that finalizer removal fails

		if isRemoval {
			diagnostic.InfoWithCtx(log, ctx, "Clean up deployment summary")
			err = r.deleteDeploymentSummary(ctx, object)
			if err != nil {
				diagnostic.InfoWithCtx(log, ctx, "failed to clean up deployment summary")
				return metrics.DeleteDeploymentSummaryFailed, ctrl.Result{}, err

			}
			diagnostic.InfoWithCtx(log, ctx, "Removing finalizer")
			if err := r.concludeDeletion(ctx, object); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "failed to remove finalizer")
				return metrics.StatusUpdateFailed, ctrl.Result{}, err
			}
			diagnostic.InfoWithCtx(log, ctx, "Deletion concluded")
			// remove, don't need to reconcile
			return metrics.DeploymentSucceeded, ctrl.Result{}, nil
		} else {
			// If the reconcile policy is once (interval == 0 or state==inactive), we should not queue a new job and return
			if reconciliationInterval == 0 {
				diagnostic.InfoWithCtx(log, ctx, "Won't polling a new job because of reconcile policy")
				return metrics.DeploymentSucceeded, ctrl.Result{}, nil
			} else {
				diagnostic.InfoWithCtx(log, ctx, "Polling a new job because of reconcile policy", "requeueAfter", reconciliationInterval-time.Since(summary.Time))
				return metrics.DeploymentSucceeded, ctrl.Result{RequeueAfter: reconciliationInterval - time.Since(summary.Time)}, nil
			}
		}
	default:
		diagnostic.ErrorWithCtx(log, ctx, errors.New("should not reach here"), "invalid deployment summary state")
		return metrics.StatusNoOp, ctrl.Result{}, fmt.Errorf("should not reach here")
	}
}

func (r *DeploymentReconciler) handleDeploymentError(ctx context.Context, object Reconcilable, summary *model.SummaryResult, isRemoval bool, reconcileInterval time.Duration, err error, log logr.Logger) (metrics.OperationStatus, ctrl.Result, error) {
	patchOptions := patchStatusOptions{}
	if isTermnalError(err, termialErrors) {
		patchOptions.terminalErr = err
	} else {
		patchOptions.nonTerminalErr = err
	}

	// update the object status
	if _, err = r.updateObjectStatus(ctx, object, summary, patchOptions, log); err != nil {
		return metrics.StatusUpdateFailed, ctrl.Result{}, err
	}

	// If there was a terminal error, then we don't return an error so the reconciler can respect the reconcile policy
	// but if there was a non-terminal error, we should return the error so the reconciler can retry
	if patchOptions.terminalErr != nil {
		diagnostic.ErrorWithCtx(log, ctx, patchOptions.terminalErr, "Deployment job failed due to terminal error.")
		return metrics.DeploymentFailed, ctrl.Result{RequeueAfter: reconcileInterval}, nil
	}
	diagnostic.ErrorWithCtx(log, ctx, patchOptions.nonTerminalErr, "Deployment job failed due to non-terminal error.")
	return metrics.QueueDeploymentFailed, ctrl.Result{}, patchOptions.nonTerminalErr
}

func (r *DeploymentReconciler) concludeDeletion(ctx context.Context, object Reconcilable) error {
	controllerutil.RemoveFinalizer(object, r.finalizerName)
	if err := r.kubeClient.Update(ctx, object); err != nil {
		return err
	}
	return nil
}

func (r *DeploymentReconciler) hasParity(ctx context.Context, object Reconcilable, summary *model.SummaryResult, log logr.Logger) bool {
	if object == nil || summary == nil { // we don't expect any of these to be nil
		return false
	}
	generationMatch := r.generationMatch(object, summary)
	operationTypeMatch := r.operationTypeMatch(object, summary)
	deploymentHashMatch := r.deploymentHashMatch(ctx, object, summary)
	diagnostic.InfoWithCtx(log, ctx, "Checking for parity", "generationMatch", generationMatch, "operationTypeMatch", operationTypeMatch, "deploymentHashMatch", deploymentHashMatch)
	return generationMatch && operationTypeMatch && deploymentHashMatch
}

func (r *DeploymentReconciler) generationMatch(object Reconcilable, summary *model.SummaryResult) bool {
	if object == nil || summary == nil { // we don't expect any of these to be nil
		return false
	}
	return summary.Generation == strconv.FormatInt(object.GetGeneration(), 10)
}

func (r *DeploymentReconciler) operationTypeMatch(object Reconcilable, summary *model.SummaryResult) bool {
	if object == nil || summary == nil { // we don't expect any of these to be nil
		return false
	}
	if summary.Summary.IsRemoval {
		return object.GetDeletionTimestamp() != nil
	}
	return object.GetDeletionTimestamp() == nil
}

func (r *DeploymentReconciler) deploymentHashMatch(ctx context.Context, object Reconcilable, summary *model.SummaryResult) bool {
	if object == nil || summary == nil { // we don't expect any of these to be nil
		return false
	}
	deployment, err := r.deploymentBuilder(ctx, object)
	if err != nil {
		return false
	}
	return summary.DeploymentHash == deployment.Hash
}

func (r *DeploymentReconciler) queueDeploymentJob(ctx context.Context, object Reconcilable, isRemoval bool, updateCorrelationId bool, operationStartTimeKey string) error {
	// If previous status was terminal and there is no parity between the summary and current object, then update correlation id.
	// This will ensure that there is a new correlation id between deployments including deployments that periodically occur.
	if updateCorrelationId && utilsmodel.IsTerminalState(object.GetStatus().ProvisioningStatus.Status) {
		r.updateCorrelationIdMetaData(ctx, object, operationStartTimeKey)
	}

	// Build the deployment object to send to the api
	deployment, err := r.deploymentBuilder(ctx, object)
	if err != nil {
		return err
	}

	// Send the deployment object to the api to queue a job
	err = r.apiClient.QueueDeploymentJob(ctx, object.GetNamespace(), isRemoval, *deployment, "", "")
	if err != nil {
		return err
	}
	return nil
}

func (r *DeploymentReconciler) getDeploymentSummary(ctx context.Context, object Reconcilable) (*model.SummaryResult, error) {
	return r.apiClient.GetSummary(ctx, r.deploymentKeyResolver(object), object.GetNamespace(), "", "")
}

func (r *DeploymentReconciler) deleteDeploymentSummary(ctx context.Context, object Reconcilable) error {
	return r.apiClient.DeleteSummary(ctx, r.deploymentKeyResolver(object), object.GetNamespace(), "", "")
}

func (r *DeploymentReconciler) updateCorrelationIdMetaData(ctx context.Context, object Reconcilable, operationStartTimeKey string) error {
	correlationId := uuid.New()
	r.patchOperationStartTime(object, operationStartTimeKey)
	object.GetAnnotations()[constants.AzureCorrelationIdKey] = correlationId.String()
	if err := r.kubeClient.Update(ctx, object); err != nil {
		return err
	}

	return nil
}

func (r *DeploymentReconciler) patchOperationStartTime(object Reconcilable, operationStartTimeKey string) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[operationStartTimeKey] = time.Now().Format(time.RFC3339)
	object.SetAnnotations(annotations)
}

func (r *DeploymentReconciler) ensureOperationState(annotations map[string]string, objectStatus *k8smodel.DeployableStatus, provisioningState string) {
	objectStatus.ProvisioningStatus.Status = provisioningState
	objectStatus.ProvisioningStatus.OperationID = annotations[constants.AzureOperationIdKey]
}

func (r *DeploymentReconciler) updateObjectStatus(ctx context.Context, object Reconcilable, summaryResult *model.SummaryResult, opts patchStatusOptions, log logr.Logger) (provisioningState string, err error) {
	status := r.determineProvisioningStatus(ctx, object, summaryResult, opts, log)
	originalStatus := object.GetStatus()
	nextStatus := originalStatus.DeepCopy()
	diagnostic.InfoWithCtx(log, ctx, "Updating object status", "status", status, "patchStatusOptions", opts)

	r.patchBasicStatusProps(ctx, object, summaryResult, status, nextStatus, opts, log)
	r.patchComponentStatusReport(ctx, object, summaryResult, nextStatus, log)
	r.updateProvisioningStatus(ctx, object, summaryResult, status, nextStatus, opts, log)

	if reflect.DeepEqual(&originalStatus, nextStatus) {
		return string(status), nil
	}
	nextStatus.LastModified = metav1.Now()
	object.SetStatus(*nextStatus)

	err = r.kubeClient.Status().Update(context.Background(), object)
	if err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to update object status")
	}
	return string(status), err

}

func (r *DeploymentReconciler) determineProvisioningStatus(ctx context.Context, object Reconcilable, summaryResult *model.SummaryResult, opts patchStatusOptions, log logr.Logger) utilsmodel.ProvisioningStatus {
	if opts.terminalErr != nil {
		// add more details of the terminal error to the status
		return utilsmodel.ProvisioningStatusFailed
	}

	if opts.nonTerminalErr != nil || summaryResult == nil || !r.hasParity(ctx, object, summaryResult, log) || opts.deploymentQueued {
		return utilsmodel.GetNonTerminalStatus(object)
	}

	summary := summaryResult.Summary
	switch summaryResult.State {
	case model.SummaryStateDone:
		// Honor OSS changes: https://github.com/eclipse-symphony/symphony/pull/148
		// Use AllAssignedDeployed instead of targetCount/successCount to verify deployment.
		status := utilsmodel.ProvisioningStatusSucceeded
		if !summary.AllAssignedDeployed {
			status = utilsmodel.ProvisioningStatusFailed
		}
		return status
	default:
		return utilsmodel.GetNonTerminalStatus(object)
	}
}

func (r *DeploymentReconciler) patchBasicStatusProps(ctx context.Context, object Reconcilable, summaryResult *model.SummaryResult, status utilsmodel.ProvisioningStatus, objectStatus *k8smodel.DeployableStatus, opts patchStatusOptions, log logr.Logger) {
	if objectStatus.Properties == nil {
		objectStatus.Properties = make(map[string]string)
	}
	defer func() { // keeping for backward compatibility. Ideally we should remove this and use the provisioning status and provisioning status output
		objectStatus.Properties["status"] = string(status)
		if opts.nonTerminalErr != nil {
			objectStatus.Properties["status-details"] = fmt.Sprintf("%s: due to %s", status, opts.nonTerminalErr.Error())
		}
	}()

	if opts.terminalErr != nil {
		objectStatus.Properties["deployed"] = "failed"
		objectStatus.Properties["targets"] = "failed"
		objectStatus.Properties["status-details"] = opts.terminalErr.Error()
		return
	}

	if summaryResult == nil || !r.hasParity(ctx, object, summaryResult, log) {
		objectStatus.Properties["deployed"] = "pending"
		objectStatus.Properties["targets"] = "pending"
		objectStatus.Properties["status-details"] = ""
		return
	}

	summary := summaryResult.Summary
	targetCount := strconv.Itoa(summary.TargetCount)
	successCount := strconv.Itoa(summary.SuccessCount)

	objectStatus.Properties["deployed"] = successCount
	objectStatus.Properties["targets"] = targetCount
	objectStatus.Properties["status-details"] = summary.SummaryMessage
}

func (r *DeploymentReconciler) patchComponentStatusReport(ctx context.Context, object Reconcilable, summaryResult *model.SummaryResult, objectStatus *k8smodel.DeployableStatus, log logr.Logger) {
	if objectStatus.Properties == nil {
		return
	}
	// If a component is ever deployed, it will always show in Status.Properties
	// If a component is not deleted, it will first be reset to Untouched and
	// then changed to corresponding status later
	for k, v := range objectStatus.Properties {
		// Check status prefix (e.g. Deleted -) since status ends with a "-"
		if utils.IsComponentKey(k) && !strings.HasPrefix(v, v1alpha2.Deleted.String()) {
			objectStatus.Properties[k] = v1alpha2.Untouched.String()
		}
	}
	if summaryResult == nil || !r.hasParity(ctx, object, summaryResult, log) {
		return
	}
	summary := summaryResult.Summary
	// Change to corresponding status
	// TargetResults should be empty if there a successful deletion
	for k, v := range summary.TargetResults {
		objectStatus.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
		for kc, c := range v.ComponentResults {
			if c.Message == "" {
				// Honor OSS changes: https://github.com/eclipse-symphony/symphony/pull/225
				// If c.Message is empty, only show c.Status.
				objectStatus.Properties["targets."+k+"."+kc] = c.Status.String()
			} else {
				objectStatus.Properties["targets."+k+"."+kc] = fmt.Sprintf("%s - %s", c.Status, c.Message)
			}
		}
	}
}

func (r *DeploymentReconciler) updateProvisioningStatus(ctx context.Context, object Reconcilable, summaryResult *model.SummaryResult, provisioningStatus utilsmodel.ProvisioningStatus, objectStatus *k8smodel.DeployableStatus, opts patchStatusOptions, log logr.Logger) {
	// THIS IS A HACK. to align with legacy expectations, we need to concatenate
	// the status with the non-terminal error message. This is not ideal and should
	// be removed in the future
	var statusText string = string(provisioningStatus)
	if opts.nonTerminalErr != nil {
		statusText = fmt.Sprintf("%s: due to %s", provisioningStatus, opts.nonTerminalErr.Error())
	}
	r.ensureOperationState(object.GetAnnotations(), objectStatus, statusText)

	// Start with a clean Error object and update all the fields
	objectStatus.ProvisioningStatus.Error = apimodel.ErrorType{}
	// Output field is updated if status is Succeeded
	objectStatus.ProvisioningStatus.Output = make(map[string]string)

	if provisioningStatus == utilsmodel.ProvisioningStatusFailed {
		errorObj := &objectStatus.ProvisioningStatus.Error

		// Fill error details into error object
		err := opts.nonTerminalErr
		if opts.terminalErr != nil {
			err = opts.terminalErr

		}

		r.deploymentErrorBuilder(summaryResult, err, errorObj)
		return
	}

	if summaryResult == nil || !r.hasParity(ctx, object, summaryResult, log) {
		return
	}
	summary := summaryResult.Summary

	outputMap := objectStatus.ProvisioningStatus.Output
	// Fill component details into output field
	for k, v := range summary.TargetResults {
		for ck, cv := range v.ComponentResults {
			outputMap[fmt.Sprintf("%s.%s", k, ck)] = cv.Status.String()
		}
	}
	if len(outputMap) == 0 {
		objectStatus.ProvisioningStatus.Output = nil
	}
}

func defaultDeploymentKeyResolver(object Reconcilable) string {
	return object.GetName()
}

func defaultProvisioningErrorBuilder(summaryResult *model.SummaryResult, err error, errorObj *apimodel.ErrorType) {
	// Fill error details into error object
	errorObj.Code = "Symphony: [500]"

	if summaryResult != nil {
		summary := summaryResult.Summary

		if summary.IsRemoval {
			errorObj.Message = fmt.Sprintf("Uninstall failed. %s", summary.SummaryMessage)
		} else {
			errorObj.Message = fmt.Sprintf("Deployment failed. %s", summary.SummaryMessage)
		}

		errorObj.Target = "Symphony"
		errorObj.Details = make([]apimodel.TargetError, 0)
		for k, v := range summary.TargetResults {
			targetObject := apimodel.TargetError{
				Code:    v.Status,
				Message: v.Message,
				Target:  k,
				Details: make([]apimodel.ComponentError, 0),
			}
			for ck, cv := range v.ComponentResults {
				targetObject.Details = append(targetObject.Details, apimodel.ComponentError{
					Code:    cv.Status.String(),
					Message: cv.Message,
					Target:  ck,
				})
			}
			errorObj.Details = append(errorObj.Details, targetObject)
		}
	}

	if err != nil {
		errorObj.Message = fmt.Sprintf("%s, %s", err.Error(), errorObj.Message)
	}
}

// checks if the error is terminal
func isTermnalError(err error, terminalErrors map[v1alpha2.State]struct{}) bool {
	if err == nil {
		return false
	}

	var coaErr v1alpha2.COAError
	if errors.As(err, &coaErr) {
		_, ok := terminalErrors[coaErr.State]
		return ok
	}

	return false
}
