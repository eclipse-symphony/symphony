/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reconcilers_test

import (
	"context"
	"errors"
	v1 "gopls-workspace/apis/fabric/v1"
	"gopls-workspace/constants"
	"gopls-workspace/reconcilers"

	. "gopls-workspace/testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Attempt Delete", func() {
	var reconciler *reconcilers.DeploymentReconciler
	var apiClient *MockApiClient
	var kubeClient client.Client
	var object *v1.Target
	var reconcileResult reconcile.Result
	var reconcileError error
	var reconcileResultPolling reconcile.Result
	var reconcileErrorPolling error
	var delayer *MockDelayer
	var jobID string

	BeforeEach(func() {
		By("building the clients")
		apiClient = &MockApiClient{}
		kubeClient = CreateFakeKubeClientForFabricGroup(
			BuildDefaultTarget(),
		)
		delayer = &MockDelayer{}

		By("building the reconciler")
		var err error
		reconciler, err = reconcilers.NewDeploymentReconciler(append(
			DefaultTestReconcilerOptions(),
			reconcilers.WithDelayFunc(delayer.Sleep),
			reconcilers.WithDeleteSyncDelay(TestDeleteSyncDelay),
			reconcilers.WithApiClient(apiClient),
			reconcilers.WithClient(kubeClient))...,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func(ctx context.Context) {
		By("fetching the object from the kubernetes api")
		object = &v1.Target{}
		Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, object)).To(Succeed())
	})

	BeforeEach(func(ctx context.Context) {
		By("adding a finalizer to the object")
		object.SetFinalizers([]string{TestFinalizer})
		Expect(kubeClient.Update(ctx, object)).To(Succeed())
		Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, object)).To(Succeed())
		Expect(object.GetFinalizers()).To(ContainElement(TestFinalizer))
	})

	BeforeEach(func() {
		By("deleting the object")
		Expect(kubeClient.Delete(context.Background(), object)).To(Succeed())
		Expect(kubeClient.Get(context.Background(), DefaultTargetNamepsacedName, object)).To(Succeed())
	})

	AfterEach(func() {
		By("checking that all mocks were called (or not called) with the expected arguments")
		apiClient.AssertExpectations(GinkgoT())
		delayer.AssertExpectations(GinkgoT())
	})

	JustBeforeEach(func(ctx context.Context) {
		By("calling the reconciler")
		_, reconcileResult, reconcileError = reconciler.AttemptUpdate(ctx, object, true, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Delete)
		annotations := object.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[constants.SummaryJobIdKey] = jobID
		object.SetAnnotations(annotations)
		_, reconcileResultPolling, reconcileErrorPolling = reconciler.PollingResult(ctx, object, true, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Delete)
	})

	When("the object has not been queued for deletion on the api but has been deployed", func() {
		BeforeEach(func(ctx context.Context) {
			jobID = uuid.New().String()
			By("returning a summary of a deployed but not deleted object from the api")
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
		})

		Context("and it successfully queues a delete job to the api", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning a successful delete job from the api")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			It("should call the api to get summary and queue a delete job", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of Deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})

			It("should return a result indicating that the reconciliation should be requeued within the polling interval", func() {
				Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
			})
		})

		Context("and it fails to queue a delete job to the api due to a non-terminal error", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning a non-terminal error from the api")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test-error"))
			})

			It("should call the api to get summary and queue a delete job", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of Deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})

			It("should kickoff a reconciliation as soon as possible because of an error", func() {
				Expect(reconcileError).To(HaveOccurred())
			})
		})

		Context("and it fails to queue a delete job to the api due to a terminal error", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning a terminal error from the api")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
			})

			It("should call the api as expected", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of failed", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
			})

			It("should have a finalizer", func() {
				Expect(object.GetFinalizers()).To(ContainElement(TestFinalizer))
			})

			It("should return a result indicating that the reconciliation should not be requeued", func() {
				Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
				Expect(reconcileError).NotTo(HaveOccurred())
			})
		})
	})

	When("the object has been queued for deletion on the api", func() {
		Context("and the delete job is still in progress", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning an in-progress delete summary from the api")
				summary := MockInProgressDeleteSummaryResult(object, "test-hash")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
			})

			It("should have called the api to get summary with the right arguments", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of Deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})

			It("should return a result indicating that the reconciliation should be requeued within the polling interval", func() {
				Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
			})

		})

		Context("and the delete job inprogress", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning a failed delete summary from the api")
				summary := MockInProgressDeleteSummaryResult(object, "test-hash")
				jobID = uuid.New().String()
				summary.Summary.JobID = jobID
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
			})

			It("should have called the api to get summary with the right arguments", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})

			It("should not have a finalizer", func() {
				Expect(object.GetFinalizers()).To(ContainElement(TestFinalizer))
			})

			It("should return a result indicating that the reconciliation", func() {
				Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
			})
		})

		Context("and the delete job has succeeded", func() {
			BeforeEach(func(ctx context.Context) {
				By("returning a successful delete summary from the api")
				summary := MockSucessSummaryResult(object, "test-hash")
				summary.Summary.IsRemoval = true
				jobID = uuid.New().String()
				summary.Summary.JobID = jobID
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
			})

			It("should have called the api to get summary with the right arguments", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should have a status of Succeeded", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Succeeded"))
			})

			It("should not have a finalizer", func() {
				Expect(object.GetFinalizers()).NotTo(ContainElement(TestFinalizer))
			})

			It("should return a result indicating that the reconciliation should not be requeued", func() {
				Expect(reconcileResultPolling.RequeueAfter).To(BeZero())
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
			})
		})
	})

	When("the delete job summary cannot be fetched from the api due to random error", func() {
		BeforeEach(func(ctx context.Context) {
			By("returning an error from the get summary api endpoint")
			apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test-error"))
		})

		It("should have called the api to get summary with the right arguments", func() {
			apiClient.AssertExpectations(GinkgoT())
		})

		It("should have a status of Deleting", func() {
			Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
		})

		It("should return a result indicating that the reconciliation polling should be requeued within the polling interval", func() {
			Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
			Expect(reconcileErrorPolling).To(HaveOccurred())
		})
	})

	When("the delete job summary cannot be fetched from the api due to not found", func() {
		BeforeEach(func(ctx context.Context) {
			By("returning an error from the get summary api endpoint")
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)
		})

		Context("so it successfully queues a delete job to the api", func() {
			BeforeEach(func(ctx context.Context) {
				By("mocking a successful call to the queue delete job api endpoint")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			It("should have called the api to get summary and queue a job", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should not return an error", func() {
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
			})

			It("should have a status of Deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})
		})

		Context("so it fails to queue a delete job to the api due to a non-terminal error", func() {
			BeforeEach(func(ctx context.Context) {
				By("mocking a non-terminal error from the queue delete job api endpoint")
				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test-error"))
			})

			It("should have called the api to get summary and queue a job", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should return an error", func() {
				Expect(reconcileError).To(HaveOccurred())
			})

			It("should have a status of Deleting", func() {
				Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
			})
		})
	})

	When("the api returns a summary in a pending state", func() {
		BeforeEach(func(ctx context.Context) {
			By("returning a pending summary from the api")
			summary := MockInProgressDeleteSummaryResult(object, "test-hash")
			summary.State = model.SummaryStatePending
			apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
		})

		It("should have called the api to get summary with the right arguments", func() {
			apiClient.AssertExpectations(GinkgoT())
		})

		It("should return a result indicating that the reconciliation should be requeued within the polling interval", func() {
			Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
			Expect(reconcileError).NotTo(HaveOccurred())
		})

	})
})
