/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reconcilers_test

import (
	"context"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	solutionv1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/reconcilers"

	. "gopls-workspace/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing timeOverDue conditions in deployment reconciler", func() {
	Context("Testing apply timeout scenario", func() {
		var reconciler *reconcilers.DeploymentReconciler
		var apiClient *MockApiClient
		var kubeClient client.Client
		var object *solutionv1.Instance
		var reconcileResult metrics.OperationStatus
		var reconcileError error

		BeforeEach(func() {
			By("setting up the reconciler")
			apiClient = &MockApiClient{}
			kubeClient = CreateFakeKubeClientForSolutionGroup(
				BuildDefaultInstance(),
			)
			var err error
			reconciler, err = reconcilers.NewDeploymentReconciler(append(
				DefaultTestReconcilerOptions(),
				reconcilers.WithApiClient(apiClient),
				reconcilers.WithClient(kubeClient))...,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		BeforeEach(func(ctx context.Context) {
			By("fetching the latest resources from kube api")
			object = &solutionv1.Instance{}
			err := kubeClient.Get(ctx, DefaultInstanceNamespacedName, object)
			Expect(err).NotTo(HaveOccurred())
		})

		When("operation start time is set and apply timeout is exceeded", func() {
			BeforeEach(func(ctx context.Context) {
				By("setting operation start time to past timestamp")
				if object.GetAnnotations() == nil {
					object.SetAnnotations(make(map[string]string))
				}
				// Set operation start time to 1 hour ago to ensure timeout
				pastTime := time.Now().Add(-1 * time.Hour)
				object.GetAnnotations()[targetOperationStartTimeKey] = pastTime.Format(time.RFC3339)

				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())

				By("setting up the api client to return pending status")
				summary := MockSucessSummaryResult(object, "test-hash")
				summary.State = model.SummaryStatePending
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
			})

			JustBeforeEach(func(ctx context.Context) {
				By("calling the polling result method")
				reconcileResult, _, reconcileError = reconciler.PollingResult(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
			})

			It("should not return error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})

			It("should return DeploymentTimedOut status", func() {
				Expect(reconcileResult).To(Equal(metrics.DeploymentTimedOut))
			})

			It("should update object status with terminal error", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(Equal("Failed"))
				Expect(object.Status.ProvisioningStatus.Error).NotTo(BeNil())
				Expect(object.Status.ProvisioningStatus.Error.Code).To(Equal("Symphony: [500]"))
				Expect(object.Status.ProvisioningStatus.Error.Message).To(ContainSubstring("failed to completely reconcile within the allocated time"))
			})
		})

		When("operation start time annotation is missing", func() {
			BeforeEach(func(ctx context.Context) {
				By("removing operation start time annotation")
				if object.GetAnnotations() == nil {
					object.SetAnnotations(make(map[string]string))
				}
				delete(object.GetAnnotations(), targetOperationStartTimeKey)

				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func(ctx context.Context) {
				By("calling the polling result method")
				reconcileResult, _, reconcileError = reconciler.PollingResult(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
			})

			It("should not return timeout error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})

			It("should return DeploymentPolling status", func() {
				Expect(reconcileResult).To(Equal(metrics.DeploymentPolling))
			})
		})

		When("operation start time is malformed", func() {
			BeforeEach(func(ctx context.Context) {
				By("setting malformed operation start time")
				if object.GetAnnotations() == nil {
					object.SetAnnotations(make(map[string]string))
				}
				object.GetAnnotations()[targetOperationStartTimeKey] = "invalid-time-format"

				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func(ctx context.Context) {
				By("calling the polling result method")
				reconcileResult, _, reconcileError = reconciler.PollingResult(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
			})

			It("should return parsing error", func() {
				Expect(reconcileError).To(HaveOccurred())
			})

			It("should return OperationStartTimeParseFailed status", func() {
				Expect(reconcileResult).To(Equal(metrics.OperationStartTimeParseFailed))
			})
		})
	})

	Context("Testing timeout with successful completion", func() {
		var reconciler *reconcilers.DeploymentReconciler
		var apiClient *MockApiClient
		var kubeClient client.Client
		var object *solutionv1.Instance

		BeforeEach(func() {
			By("setting up the reconciler")
			apiClient = &MockApiClient{}
			kubeClient = CreateFakeKubeClientForSolutionGroup(
				BuildDefaultInstance(),
			)
			var err error
			reconciler, err = reconcilers.NewDeploymentReconciler(append(
				DefaultTestReconcilerOptions(),
				reconcilers.WithApiClient(apiClient),
				reconcilers.WithClient(kubeClient))...,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		BeforeEach(func(ctx context.Context) {
			By("fetching the latest resources from kube api")
			object = &solutionv1.Instance{}
			err := kubeClient.Get(ctx, DefaultInstanceNamespacedName, object)
			Expect(err).NotTo(HaveOccurred())
		})

		When("operation completes successfully before timeout", func() {
			BeforeEach(func(ctx context.Context) {
				By("setting recent operation start time")
				if object.GetAnnotations() == nil {
					object.SetAnnotations(make(map[string]string))
				}
				recentTime := time.Now().Add(-1 * time.Minute)
				object.GetAnnotations()[targetOperationStartTimeKey] = recentTime.Format(time.RFC3339)

				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())

				By("setting up the api client to return successful status")
				summary := MockSucessSummaryResult(object, "test-hash")
				summary.State = model.SummaryStateDone
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
			})

			It("should complete successfully without timeout", func(ctx context.Context) {
				_, _, err := reconciler.PollingResult(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
				Expect(err).NotTo(HaveOccurred())
				Expect(object.Status.ProvisioningStatus.Status).To(Equal("Succeeded"))
			})
		})
	})
})
