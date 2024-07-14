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
	k8smodel "gopls-workspace/apis/model/v1"
	"gopls-workspace/constants"
	"gopls-workspace/reconcilers"

	. "gopls-workspace/testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Reconcile Policies", func() {
	var reconciler *reconcilers.DeploymentReconciler
	var apiClient *MockApiClient
	var kubeClient client.Client
	var object *v1.Target
	var reconcileResult reconcile.Result
	var reconcileError error

	BeforeEach(func() {
		By("building the clients")
		apiClient = &MockApiClient{}
		kubeClient = CreateFakeKubeClientForFabricGroup(
			BuildDefaultTarget(),
		)

		By("building the reconciler")
		var err error
		reconciler, err = reconcilers.NewDeploymentReconciler(append(
			DefaultTestReconcilerOptions(),
			reconcilers.WithApiClient(apiClient),
			reconcilers.WithClient(kubeClient))...,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func(ctx context.Context) {
		By("fetching the latest object")
		object = &v1.Target{}
		err := kubeClient.Get(ctx, DefaultTargetNamepsacedName, object)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func(ctx context.Context) {
		By("asserting that mocks were called exactly as expected")
		apiClient.AssertExpectations(GinkgoT())
	})

	JustBeforeEach(func(ctx context.Context) {
		By("calling the reconciler")
		_, reconcileResult, reconcileError = reconciler.AttemptUpdate(ctx, object, logr.Discard(), targetOperationStartTimeKey, constants.ActivityCategory_Activity, constants.ActivityOperation_Write)
	})

	Context("object has invalid reconcile policy", func() {
		When("reconcile policy state is invalid", func() {
			BeforeEach(func(ctx context.Context) {
				By("updating the object with an invalid reconcile policy")
				object.Spec.ReconciliationPolicy = &k8smodel.ReconciliationPolicySpec{State: "invalid"}
				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			BeforeEach(func() {
				By("mocking the summary response with a successful deployment")
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
			})

			It("should make the expected api calls", func() {
				apiClient.AssertExpectations(GinkgoT())
			})

			It("should fall back to default reconciliation interval", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})

			It("should requue after some time", func() {
				Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
			})
		})

		When("reconcile policy state is valid but interval is invalid", func() {
			BeforeEach(func(ctx context.Context) {
				By("updating the object with an invalid reconcile policy")
				object.Spec.ReconciliationPolicy = &k8smodel.ReconciliationPolicySpec{State: "active", Interval: ToPointer("invalid")}
				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			BeforeEach(func() {
				By("mocking the summary response with a successful deployment")
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
			})

			It("should fall back to default reconciliation interval", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})

			It("should requue after some time", func() {
				Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
			})
		})
	})

	Context("object has no reconcile policy", func() {
		// use the default reconcile interval
		BeforeEach(func(ctx context.Context) {
			By("updating the object with no reconcile policy")
			object.Spec.ReconciliationPolicy = nil
			err := kubeClient.Update(ctx, object)
			Expect(err).NotTo(HaveOccurred())
		})

		BeforeEach(func() {
			By("mocking the summary response with a successful deployment")
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
		})

		It("should make the expected api calls", func() {
			apiClient.AssertExpectations(GinkgoT())
		})

		It("should fall back to default reconciliation interval", func() {
			Expect(reconcileError).NotTo(HaveOccurred())
		})

		It("should requue after some time", func() {
			Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
		})
	})

	Context("object has valid reconcile policy", func() {
		When("reconcile policy state is active", func() {
			BeforeEach(func(ctx context.Context) {
				By("updating the object with a valid reconcile policy")
				object.Spec.ReconciliationPolicy = &k8smodel.ReconciliationPolicySpec{State: "active", Interval: ToPointer("1m")}
				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("deployment is not queued to the api due to non-terminal error", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a non-terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should re-queue a reconcile job immediately due to non-terminal error", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
				It("should have a status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("deployment is not queued to the api due to terminal error", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should have a status of failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
				It("should queue a reconcile job after reconcile interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(time.Minute))
				})
			})

			Context("deployment job queued successfully", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a successful deployment")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should queue a reconcile job to poll for status", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})
			})

			Context("deployment to api is completed successful", func() {
				BeforeEach(func() {
					By("mocking the summary response with a successful deployment")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should return not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should queue a reconcile job after reconcile interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(time.Minute))
				})

				It("should have a status of Succeeded", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Succeeded"))
				})
			})
			Context("deployment to api is completed with failure", func() {
				BeforeEach(func() {
					By("mocking the summary response with a failed deployment")
					summary := MockSucessSummaryResult(object, "test-hash")
					summary.Summary.SuccessCount = 0
					summary.Summary.AllAssignedDeployed = false
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should return not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should queue a reconcile job after reconcile interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(time.Minute))
				})

				It("should have a status of Failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
			})
		})

		When("reconcile policy type is once", func() {
			BeforeEach(func(ctx context.Context) {
				By("updating the object with a valid reconcile policy")
				object.Spec.ReconciliationPolicy = &k8smodel.ReconciliationPolicySpec{State: "active", Interval: ToPointer("0s")}
				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("deployment is not queued to the api due to non-terminal error", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a non-terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should re-queue a reconcile job immediately", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
				It("should have a status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("deployment is not queued to the api due to terminal error", func() {

				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})

			Context("deployment job queued successfully", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a successful deployment")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should queue a reconcile job to poll for status", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})

				It("should have a status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("deployment to api is completed successful", func() {
				BeforeEach(func() {
					By("mocking the summary response with a successful deployment")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of succeeded", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Succeeded"))
				})

				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})
			Context("deployment to api is completed with failure", func() {
				BeforeEach(func() {
					By("mocking the summary response with a failed deployment")
					summary := MockSucessSummaryResult(object, "test-hash")
					summary.Summary.SuccessCount = 0
					summary.Summary.AllAssignedDeployed = false
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})
		})

		When("reconcile policy state is inactive", func() {
			BeforeEach(func(ctx context.Context) {
				By("updating the object with a valid reconcile policy: state inactive, interval 10m (interval will be ignored)")
				object.Spec.ReconciliationPolicy = &k8smodel.ReconciliationPolicySpec{State: "inactive", Interval: ToPointer("10m")}
				err := kubeClient.Update(ctx, object)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("deployment is not queued to the api due to non-terminal error", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a non-terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should re-queue a reconcile job immediately", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
				It("should have a status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("deployment is not queued to the api due to terminal error", func() {

				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a terminal error")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})

			Context("deployment job queued successfully", func() {
				BeforeEach(func() {
					By("mocking the summary response with not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking the queue deployment response with a successful deployment")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should queue a reconcile job to poll for status", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})

				It("should have a status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("deployment to api is completed successful", func() {
				BeforeEach(func() {
					By("mocking the summary response with a successful deployment")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of succeeded", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Succeeded"))
				})

				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})
			Context("deployment to api is completed with failure", func() {
				BeforeEach(func() {
					By("mocking the summary response with a failed deployment")
					summary := MockSucessSummaryResult(object, "test-hash")
					summary.Summary.SuccessCount = 0
					summary.Summary.AllAssignedDeployed = false
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should make the expected api calls", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should have a status of failed", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})
				It("should not queue a reconcile job because the reconcile session is done", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(reconcileResult.Requeue).To(BeFalse())
					Expect(reconcileResult.RequeueAfter).To(BeZero())
				})
			})
		})
	})

	Context("object has stil not finished reconciling and has timed out", func() {
		BeforeEach(func(ctx context.Context) {
			By("mocking a summary response with in progress deployment")
			apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockInProgressSummaryResult(object, "test-hash"), nil)
		})

		JustBeforeEach(func(ctx context.Context) {
			Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))

			By("updating the object operation start time to be in the past")
			object.GetAnnotations()[targetOperationStartTimeKey] = time.Now().Add(-time.Hour).Format(time.RFC3339)
			Expect(kubeClient.Update(ctx, object)).NotTo(HaveOccurred())
		})

		JustBeforeEach(func(ctx context.Context) {
			By("calling the reconciler")
			_, reconcileResult, reconcileError = reconciler.AttemptUpdate(ctx, object, logr.Discard(), targetOperationStartTimeKey, constants.ActivityCategory_Activity, constants.ActivityOperation_Write)
		})

		It("should not return an error", func() {
			Expect(reconcileError).NotTo(HaveOccurred())
		})

		It("should queue a reconcile job after reconcile interval", func() {
			Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
		})

		It("should have a status of failed", func() {
			Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
		})
	})
})
