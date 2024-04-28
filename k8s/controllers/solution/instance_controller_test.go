/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

/*
import (
	"context"
	"errors"
	fabricv1 "gopls-workspace/apis/fabric/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"

	. "gopls-workspace/testing"
	"gopls-workspace/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Instance controller", Ordered, func() {
	var apiClient *MockApiClient
	var kubeClient client.Client
	var controller *InstanceReconciler
	var instance *solutionv1.Instance
	var target *fabricv1.Target
	var solution *solutionv1.Solution
	var reconcileError error
	var reconcileResult ctrl.Result

	BeforeEach(func() {
		By("setting up the controller")

		// We'll setup the controller exactly how it would have been setup if it was done by the manager
		// This means we'll need to mock out the api client and kube client
		var err error
		apiClient = &MockApiClient{}
		kubeClient = CreateFakeKubeClientForSolutionAndFabricGroup(
			BuildDefaultInstance(),
			BuildDefaultTarget(),
			BuildDefaultSolution(),
		)
		controller = &InstanceReconciler{
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

	BeforeEach(func(ctx context.Context) {
		By("fetching resources")
		instance = &solutionv1.Instance{}
		Expect(kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)).To(Succeed())

		target = &fabricv1.Target{}
		Expect(kubeClient.Get(ctx, DefaultTargetNamepsacedName, target)).To(Succeed())

		solution = &solutionv1.Solution{}
		Expect(kubeClient.Get(ctx, DefaultSolutionNamespacedName, solution)).To(Succeed())

		By("mocking the get solution call")
		solution := &model.SolutionState{}
		solution.ObjectMeta.Name = "test-solution"
		apiClient.On("GetSolution", mock.Anything, mock.Anything, mock.Anything).Return(solution, nil)
	})

	Describe("Reconcile", func() {
		JustBeforeEach(func(ctx context.Context) {
			By("simulating a reconcile event")
			reconcileResult, reconcileError = controller.Reconcile(ctx, ctrl.Request{NamespacedName: DefaultInstanceNamespacedName})
		})
		When("the instance is created", func() {

			JustBeforeEach(func(ctx context.Context) {
				By("fetching the instance")
				Expect(kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)).To(Succeed())
			})

			Context("and all necessary resources are present in the cluster", func() {
				Context("and the deployment completed successfully", func() {
					BeforeEach(func() {
						By("mocking the get summary call to return a successful deployment")
						hash := utils.HashObjects(utils.DeploymentResources{Instance: *instance, Solution: *solution, TargetCandidates: []fabricv1.Target{*target}})
						apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(instance, hash), nil)
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

					It("should queue anotther reconciliation as soon as possible", func() {
						Expect(reconcileError).To(HaveOccurred())
					})
					It("should have a status of reconciling", func() {
						Expect(instance.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					})
				})
			})

			Context("and the solution is not present in the cluster", func() {
				BeforeEach(func(ctx context.Context) {
					By("deleting the solution")
					Expect(kubeClient.Delete(ctx, solution)).To(Succeed())

					By("mocking a successful deployment to the api")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				BeforeEach(func() {
					By("mocking the get summary call to return a not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)
				})

				It("should have a status of Reconciling", func() {
					Expect(instance.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})

				It("should requeue without error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
			})

			Context("and the target is not present in the cluster", func() {
				BeforeEach(func(ctx context.Context) {
					By("deleting the target")
					Expect(kubeClient.Delete(ctx, target)).To(Succeed())
				})

				BeforeEach(func() {
					By("mocking the get summary call to return a not found error")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("mocking a successful deployment to the api")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				It("should have a status of Reconciling", func() {
					Expect(instance.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})

				It("should requeue without error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
			})
		})

		When("the instance is not found", func() {
			BeforeEach(func(ctx context.Context) {
				By("deleting the instance")
				Expect(kubeClient.Delete(ctx, instance)).To(Succeed())
			})

			It("should not return an error", func() {
				Expect(reconcileError).ToNot(HaveOccurred())
			})
		})

		When("the instance is marked for deletion", func() {
			BeforeEach(func(ctx context.Context) {
				By("adding a finalizer to the instance")
				instance.SetFinalizers([]string{instanceFinalizerName})

				By("updating the instance")
				Expect(kubeClient.Update(ctx, instance)).To(Succeed())
				Expect(kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)).To(Succeed())
				Expect(instance.GetFinalizers()).To(ContainElement(instanceFinalizerName))
			})

			BeforeEach(func(ctx context.Context) {
				By("deleting the instance")
				Expect(kubeClient.Delete(ctx, instance)).To(Succeed())
			})

			Context("and the deletion deployment is successful", func() {
				BeforeEach(func(ctx context.Context) {
					By("simulating a completed delete deployment from the api")
					hash := utils.HashObjects(utils.DeploymentResources{Instance: *instance, Solution: *solution, TargetCandidates: []fabricv1.Target{*target}})
					summary := MockSucessSummaryResult(instance, hash)
					summary.Summary.IsRemoval = true
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should no longer exist in the kubernetes api", func(ctx context.Context) {
					By("fetching the updated instance")
					err := kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})
			})

			Context("and the deletion deployment is still in progress", func() {
				BeforeEach(func(ctx context.Context) {
					By("simulating a pending delete deployment from the api")
					hash := utils.HashObjects(utils.DeploymentResources{Instance: *instance, Solution: *solution, TargetCandidates: []fabricv1.Target{*target}})
					summary := MockInProgressDeleteSummaryResult(instance, hash)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				JustBeforeEach(func(ctx context.Context) {
					By("fetching the instance")
					Expect(kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)).To(Succeed())
				})

				It("should not return an error", func() {
					Expect(reconcileError).ToNot(HaveOccurred())
				})

				It("should have a status of deleting", func() {
					Expect(instance.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
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
					By("fetching the instance")
					Expect(kubeClient.Get(ctx, DefaultInstanceNamespacedName, instance)).To(Succeed())
				})

				It("should have a status of deleting", func() {
					Expect(instance.Status.ProvisioningStatus.Status).To(ContainSubstring("Deleting"))
				})

				It("should requeue as soon as possible due to error", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Solution Events", func() {
		When("the solution referenced by the instance is changed", func() {
			var requests []ctrl.Request
			BeforeEach(func(ctx context.Context) {
				By("simulating a call to the handleSolution function")
				requests = controller.handleSolution(solution)
			})

			It("should return a request for the instance", func() {
				Expect(requests).To(ContainElement(ctrl.Request{NamespacedName: DefaultInstanceNamespacedName}))
			})
		})
	})

	Describe("Target Events", func() {
		When("the target referenced by the instance is changed", func() {
			var requests []ctrl.Request
			BeforeEach(func(ctx context.Context) {
				By("simulating a call to the handleTarget function")
				requests = controller.handleTarget(target)
			})

			It("should return a request for the instance", func() {
				Expect(requests).To(ContainElement(ctrl.Request{NamespacedName: DefaultInstanceNamespacedName}))
			})
		})
	})
})
*/
