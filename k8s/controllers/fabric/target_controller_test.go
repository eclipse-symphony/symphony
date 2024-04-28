/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package fabric

/*
import (
	. "gopls-workspace/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)
*/
// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
/* var _ = Describe("Target controller", Ordered, func() {
	var apiClient *MockApiClient
	var kubeClient client.Client
	var controller *TargetReconciler
	var target *symphonyv1.Target
	var reconcileError error
	var reconcileResult ctrl.Result

	BeforeEach(func() {
		By("setting up the controller")

		// We'll setup the controller exactly how it would have been setup if it was done by the manager
		// This means we'll need to mock out the api client and kube client
		var err error
		apiClient = &MockApiClient{}
		kubeClient = CreateFakeKubeClientForFabricGroup(
			BuildDefaultTarget(),
		)
		controller = &TargetReconciler{
			Client:                 kubeClient,
			Scheme:                 kubeClient.Scheme(),
			ReconciliationInterval: TestReconcileInterval,
			PollInterval:           TestPollInterval,
			DeleteTimeOut:          TestReconcileTimout,
			ApiClient:              apiClient,
		}

		controller.dr, err = controller.buildDeploymentReconciler()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Reconcile", func() {
		BeforeEach(func(ctx context.Context) {
			By("fetching resource")
			target = &symphonyv1.Target{}
			Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())
		})

		JustBeforeEach(func(ctx context.Context) {
			By("simulating a reconcile event")
			reconcileResult, reconcileError = controller.Reconcile(ctx, ctrl.Request{NamespacedName: DefaultTargetNamepsacedName})
		})

		When("the target is created", func() {
			JustBeforeEach(func(ctx context.Context) {
				By("fetching the target")
				Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())
			})

			Context("and the deployment completed successfully", func() {
				BeforeEach(func() {
					By("mocking the get summary call to return a successful deployment")
					hash := utils.HashObjects(utils.DeploymentResources{TargetCandidates: []symphonyv1.Target{*target}})
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(target, hash), nil)
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})

				It("should requeue after the reconciliation interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(controller.ReconciliationInterval))
				})
			})

			Context("and the deployment failed due to some error", func() {
				BeforeEach(func() {
					By("mocking the get summary call to return a not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking a failed deployment to the api")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error"))
				})

				It("should queue another reconciliation as soon as possible", func() {
					Expect(reconcileError).To(HaveOccurred())
				})

				It("should have a status of reconciling", func() {
					Expect(target.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("and the deployment completed with errors", func() {
				BeforeEach(func() {
					By("mocking the get summary call to return a successful deployment")
					hash := utils.HashObjects(utils.DeploymentResources{TargetCandidates: []symphonyv1.Target{*target}})
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockFailureSummaryResult(target, hash), nil)
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})

				It("should requeue after the reconciliation interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(controller.ReconciliationInterval))
				})

				It("should have a status of failed", func() {
					Expect(target.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
				})

				It("should have custom summary of errors", func() {
					Expect(target.Status.ProvisioningStatus.Error.Code).To(Equal("ErrorCode"))
					Expect(target.Status.ProvisioningStatus.Error.Details).To(ContainElement(apimodel.TargetError{
						Code:    "Update Failed",
						Message: "failed",
						Target:  "comp1",
					}))
				})
			})
		})

		When("the target is not found", func() {
			BeforeEach(func(ctx context.Context) {
				By("deleting the target")
				Expect(kubeClient.Delete(ctx, target)).To(Succeed())
			})

			It("should not return an error", func() {
				Expect(reconcileError).ToNot(HaveOccurred())
			})
		})

		When("the target is marked for deletion", func() {
			BeforeEach(func(ctx context.Context) {
				By("adding a finalizer to the target")
				target.SetFinalizers([]string{targetFinalizerName})

				By("updating the target")
				Expect(kubeClient.Update(ctx, target)).To(Succeed())
				Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())
				Expect(target.GetFinalizers()).To(ContainElement(targetFinalizerName))
			})

			BeforeEach(func(ctx context.Context) {
				By("deleting the target")
				Expect(kubeClient.Delete(ctx, target)).To(Succeed())
			})

			Context("and the deletion deployment is successful", func() {
				BeforeEach(func(ctx context.Context) {
					By("simulating a completed delete deployment from the api")
					hash := utils.HashObjects(utils.DeploymentResources{TargetCandidates: []symphonyv1.Target{*target}})
					summary := MockSucessSummaryResult(target, hash)
					summary.Summary.IsRemoval = true
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should no longer exist in the kubernetes api", func(ctx context.Context) {
					By("fetching the updated target")
					err := kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})
			})

			Context("and the deletion deployment is still in progress", func() {
				BeforeEach(func(ctx context.Context) {
					By("simulating a pending delete deployment from the api")
					hash := utils.HashObjects(utils.DeploymentResources{TargetCandidates: []symphonyv1.Target{*target}})
					summary := MockInProgressDeleteSummaryResult(target, hash)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				JustBeforeEach(func(ctx context.Context) {
					By("fetching the target")
					Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})

				It("should have a status of deleting", func() {
					Expect(target.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
				})

				It("should requeue after the poll interval", func() {
					Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(controller.PollInterval))
				})
			})

			Context("and the deletion deployment failed due to random error", func() {
				BeforeEach(func(ctx context.Context) {
					By("simulating a failed delete deployment from the api")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("some error"))
				})

				JustBeforeEach(func(ctx context.Context) {
					By("fetching the target")
					Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())
				})

				It("should have a status of deleting", func() {
					Expect(target.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
				})

				It("should requeue as soon as possible due to error", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
			})
		})
	})
})
*/
