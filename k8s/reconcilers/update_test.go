package reconcilers_test

import (
	"context"
	"errors"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	solutionv1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/constants"
	"gopls-workspace/reconcilers"

	. "gopls-workspace/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating a reconciler", func() {
	When("no options are provided", func() {
		It("should return an error", func() {
			_, err := reconcilers.NewDeploymentReconciler()
			Expect(err).To(HaveOccurred())
		})
	})

	When("only finalizer name is provided", func() {
		It("should return an error", func() {
			_, err := reconcilers.NewDeploymentReconciler(
				reconcilers.WithFinalizerName("test-finalizer"),
			)
			Expect(err).To(HaveOccurred())
		})
	})

	When("only finalizer and kube client are provided", func() {
		It("should return an error", func() {
			_, err := reconcilers.NewDeploymentReconciler(
				reconcilers.WithFinalizerName("test-finalizer"),
				reconcilers.WithClient(CreateFakeKubeClientForFabricGroup()),
			)
			Expect(err).To(HaveOccurred())
		})
	})

	When("only finalizer, kube client and api client are provided", func() {
		It("should return an error", func() {
			_, err := reconcilers.NewDeploymentReconciler(
				reconcilers.WithFinalizerName("test-finalizer"),
				reconcilers.WithClient(CreateFakeKubeClientForFabricGroup()),
				reconcilers.WithApiClient(&MockApiClient{}),
			)
			Expect(err).To(HaveOccurred())
		})
	})

	When("all required options are provided", func() {
		It("should not return an error", func() {
			_, err := reconcilers.NewDeploymentReconciler(
				reconcilers.WithFinalizerName("test-finalizer"),
				reconcilers.WithClient(CreateFakeKubeClientForFabricGroup()),
				reconcilers.WithApiClient(&MockApiClient{}),
				reconcilers.WithDeploymentBuilder(CreateSimpleDeploymentBuilder()),
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("all required and default options are provided", func() {
		It("should not return an error", func() {
			reconciler, err := reconcilers.NewDeploymentReconciler(append(
				DefaultTestReconcilerOptions(),
				reconcilers.WithClient(CreateFakeKubeClientForFabricGroup()),
				reconcilers.WithApiClient(&MockApiClient{}),
				reconcilers.WithDeploymentErrorBuilder(func(*model.SummaryResult, error, *apimodel.ErrorType) {}),
				reconcilers.WithDeploymentKeyResolver(func(obj reconcilers.Reconcilable) string { return obj.GetName() }),
			)...)
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciler).NotTo(BeNil())
		})
	})
})

var _ = Describe("Calling 'AttemptUpdate' on object", func() {
	Context("all required and default options are provided and default objects exist in kube api", func() {
		var reconciler *reconcilers.DeploymentReconciler
		var apiClient *MockApiClient
		var kubeClient client.Client
		var object *solutionv1.Instance
		var reconcileResult reconcile.Result
		var reconcileError error
		var reconcileResultPolling reconcile.Result
		var reconcileErrorPolling error

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

		JustBeforeEach(func(ctx context.Context) {
			By("calling the reconciler")
			_, reconcileResult, reconcileError = reconciler.AttemptUpdate(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
			_, reconcileResultPolling, reconcileErrorPolling = reconciler.PollingResult(ctx, object, false, logr.Discard(), targetOperationStartTimeKey, constants.ActivityOperation_Write)
		})

		When("object is successfully deployed", func() {
			BeforeEach(func(ctx context.Context) {
				By("setting up the api client with a successful response")

				apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "test-hash"), nil)
			})

			It("should add a finalizer to the object", func() {
				Expect(object.GetFinalizers()).To(ContainElement("test-finalizer"))
			})
			It("should call api client with correct parameters", func() {
				apiClient.AssertExpectations(GinkgoT())
			})
			It("should have status Succeeded", func() {
				Expect(reconcileErrorPolling).NotTo(HaveOccurred())
				Expect(object.Status.ProvisioningStatus.Status).To(Equal("Succeeded"))
			})
			It("should requue after some time", func() {
				Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
			})
		})

		When("object has not been deployed", func() {
			Context("api returns not found when queried for summary", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with an undeployed response")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)

					By("setting up the api client with a successful deployment queued response")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				It("should call api client with correct parameters", func() {
					apiClient.AssertExpectations(GinkgoT())
					Expect(reconcileError).NotTo(HaveOccurred())
				})

				It("should have status Reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(Equal("Reconciling"))
				})

				It("should requue after some time", func() {
					Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})
			})

			Context("api returns pending when queried for summary", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with an pending response")
					summary := MockSucessSummaryResult(object, "test-hash")
					summary.State = model.SummaryStatePending
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				It("should call api client with correct parameters", func() {
					apiClient.AssertExpectations(GinkgoT())
					Expect(reconcileError).NotTo(HaveOccurred())
				})

				It("should requue after some time", func() {
					Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})
			})

			Context("api returns an error when queried for summary", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with an error response")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test error"))
				})

				It("should call api client with correct parameters", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("should requeue because of error", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("test error")) // TODO: this is not ideal. Error should be in a separate field
					Expect(reconcileErrorPolling).To(HaveOccurred())
				})

				It("should have status Reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
				})
			})

			Context("a terminal error occurs when trying to queue deployment job", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with a successful deployment queued response")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
				})

				It("should not queue further reconcile jobs", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should have a provisioning status of failed", func() {})
			})

			Context("a non-terminal error occurs when trying to queue deployment job", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with a successful deployment queued response")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(nil, NotFoundError)
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
				})
				It("should queue further reconcile jobs due to error", func() {
					Expect(reconcileError).To(HaveOccurred())
				})
				It("should have a provisioning status of reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("test error")) // TODO: this is not ideal. Error should be in a separate field
				})
			})
		})

		When("object is reconciling to poll for deployment status", func() {
			Context("api returns a summary indicating deployment is in progress", func() {
				BeforeEach(func(ctx context.Context) {
					By("setting up the api client with a successful deployment queued response")
					apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockInProgressSummaryResult(object, "test-hash"), nil)
				})

				It("should call api client with correct parameters", func() {
					apiClient.AssertExpectations(GinkgoT())
				})

				It("queueing should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("polling should not return an error", func() {
					Expect(reconcileErrorPolling).NotTo(HaveOccurred())
				})
				It("should have status Reconciling", func() {
					Expect(object.Status.ProvisioningStatus.Status).To(Equal("Reconciling"))
				})
				It("should requue after some time", func() {
					Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
				})
				It("should updpate the compoent status with progress", func() {
					Expect(object.Status.Properties["targets.default-target.comp1"]).To(ContainSubstring("updated"))
					Expect(object.Status.Properties["targets.default-target.comp2"]).To(ContainSubstring("pending"))
				})
			})

			When("api returns a summary for a different version of the object", func() {
				BeforeEach(func() {
					By("setting up the api client with a summary response for a different version of the object")
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(object, "another-hash"), nil)
				})
				Context("successfully queues a deployment job to api", func() {
					BeforeEach(func(ctx context.Context) {
						By("allowing a succesful queued deployment response")
						apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
					})

					It("should call api client with correct parameters", func() {
						apiClient.AssertExpectations(GinkgoT())
					})

					It("queueing should not return an error", func() {
						Expect(reconcileError).NotTo(HaveOccurred())
					})
					It("polling should not return an error", func() {
						Expect(reconcileErrorPolling).NotTo(HaveOccurred())
					})
					It("should have status Reconciling", func() {
						Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					})
					It("should requue after some time", func() {
						Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
					})
				})

				Context("fails to queue a deployment job due to a non-terminal error", func() {
					BeforeEach(func(ctx context.Context) {
						By("mocking a non-terminal error when queuing a deployment job")
						apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
					})

					It("should call api client with correct parameters", func() {
						apiClient.AssertExpectations(GinkgoT())
					})

					It("should requeue due to error", func() {
						Expect(reconcileError).To(HaveOccurred())
					})
					It("should have status Reconciling", func() {
						Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					})
				})
			})

			When("api returns a summary that's older than the reconcile interval", func() {
				BeforeEach(func() {
					By("mocking a summary that's older than the reconcile interval")
					summary := MockSucessSummaryResult(object, "test-hash")
					summary.Time = summary.Time.Add(-20 * TestReconcileInterval)
					apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(summary, nil)
				})

				Context("successfully queues a deployment job to api", func() {
					BeforeEach(func(ctx context.Context) {
						By("allowing a succesful queued deployment response")
						apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
					})

					It("should call api client with correct parameters", func() {
						apiClient.AssertExpectations(GinkgoT())
					})

					It("should not return an error", func() {
						Expect(reconcileError).NotTo(HaveOccurred())
					})
					It("should have status Reconciling", func() {
						Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					})
					It("should requue after some time", func() {
						Expect(reconcileResultPolling.RequeueAfter).To(BeWithin("1s").Of(TestPollInterval))
					})
					It("should requue after some time", func() {
						Expect(reconcileResult.RequeueAfter).To(BeWithin("1s").Of(TestReconcileInterval))
					})
				})

				Context("fails to queue a deployment job due to a non-terminal error", func() {
					BeforeEach(func(ctx context.Context) {
						By("mocking a non-terminal error when queuing a deployment job")
						apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
					})

					It("should call api client with correct parameters", func() {
						apiClient.AssertExpectations(GinkgoT())
					})

					It("should requeue due to error", func() {
						Expect(reconcileError).To(HaveOccurred())
					})
					It("should have status Reconciling", func() {
						Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Reconciling"))
					})
				})

				Context("fails to queue a deployment job due to a terminal error", func() {
					BeforeEach(func(ctx context.Context) {
						By("mocking a terminal error when queuing a deployment job")
						apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(TerminalError)
					})

					It("should call api client with correct parameters", func() {
						apiClient.AssertExpectations(GinkgoT())
					})

					It("should not requeue due to error", func() {
						Expect(reconcileError).NotTo(HaveOccurred())
					})

					It("polling should not requeue due to error", func() {
						Expect(reconcileErrorPolling).NotTo(HaveOccurred())
					})

					It("should have status Failed", func() {
						Expect(object.Status.ProvisioningStatus.Status).To(ContainSubstring("Failed"))
					})
				})
			})
		})
	})
})
