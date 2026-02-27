/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package testing

import (
	"context"
	"errors"
	"fmt"
	ai_v1 "gopls-workspace/apis/ai/v1"
	fabric_v1 "gopls-workspace/apis/fabric/v1"
	federation_v1 "gopls-workspace/apis/federation/v1"
	k8smodel "gopls-workspace/apis/model/v1"
	solution_v1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/reconcilers"
	"gopls-workspace/utils"
	"strconv"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	gomegaTypes "github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type (
	MockApiClient struct {
		mock.Mock
	}

	MockDelayer struct {
		mock.Mock
	}

	TimeMatcher struct {
		wiggleroom time.Duration
		duration   time.Duration
	}
)

const (
	TestFinalizer         = "test-finalizer"
	TestPollInterval      = 1 * time.Second
	TestReconcileInterval = 10 * TestPollInterval
	TestReconcileTimout   = 20 * TestPollInterval
	TestDeleteSyncDelay   = 5 * time.Second
	TestNamespace         = "default"
	TestScope             = "symphony-test-scope"
)

var (
	_ utils.ApiClient           = &MockApiClient{}
	_ gomegaTypes.GomegaMatcher = &TimeMatcher{}
)

var (
	DefaultTargetNamepsacedName   = types.NamespacedName{Name: "testtarget", Namespace: TestNamespace}
	DefaultInstanceNamespacedName = types.NamespacedName{Name: "testinstance", Namespace: TestNamespace}
	DefaultSolutionNamespacedName = types.NamespacedName{Name: "solution-v-version1", Namespace: TestNamespace}
	SolutionReferenceName         = "solution:version1"

	TerminalError = v1alpha2.NewCOAError(errors.New(""), "timed out", v1alpha2.TimedOut)
	NotFoundError = v1alpha2.NewCOAError(errors.New(""), "not found", v1alpha2.NotFound)
)

func (m *MockDelayer) Sleep(duration time.Duration) {
	m.Called(duration)
}

func CreateFakeKubeClientForSolutionAndFabricGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{
			BuildDefaultInstance(),
			BuildDefaultSolution(),
			BuildDefaultTarget(),
		}
	}

	_ = solution_v1.AddToScheme(scheme)
	_ = fabric_v1.AddToScheme(scheme)
	clientObj := []client.Object{
		&solution_v1.Instance{},
		&fabric_v1.Target{},
		&solution_v1.Solution{},
	}
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		WithStatusSubresource(clientObj...).
		WithIndex(&solution_v1.Instance{}, "spec.solution", func(rawObj client.Object) []string {
			instance := rawObj.(*solution_v1.Instance)
			return []string{instance.Spec.Solution}
		}).
		Build()
}

func CreateFakeKubeClientForSolutionGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{
			BuildDefaultInstance(),
			BuildDefaultSolution(),
		}
	}

	_ = solution_v1.AddToScheme(scheme)
	clientObj := []client.Object{
		&solution_v1.Instance{},
		&solution_v1.Solution{},
	}
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		WithStatusSubresource(clientObj...).
		WithIndex(&solution_v1.Instance{}, "spec.solution", func(rawObj client.Object) []string {
			instance := rawObj.(*solution_v1.Instance)
			return []string{instance.Spec.Solution}
		}).
		Build()
}

func CreateFakeKubeClientForFabricGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{
			BuildDefaultTarget(),
		}
	}

	_ = fabric_v1.AddToScheme(scheme)
	clientObj := []client.Object{
		&fabric_v1.Target{},
	}
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		WithStatusSubresource(clientObj...).
		Build()
}

func CreateFakeKubeClientForFederationGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{}
	}

	_ = federation_v1.AddToScheme(scheme)
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		Build()
}

func CreateFakeKubeClientForAIGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{}
	}

	_ = ai_v1.AddToScheme(scheme)
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		Build()
}

func CreateFakeKubeClientForWorkflowGroup(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	if objects == nil {
		objects = []client.Object{}
	}

	_ = ai_v1.AddToScheme(scheme)
	return fake.NewClientBuilder().
		WithObjects(objects...).
		WithScheme(scheme).
		Build()
}

func MockSucessSummaryResultWithJobID(obj reconcilers.Reconcilable, hash string, jobID string) *model.SummaryResult {
	return &model.SummaryResult{
		Summary: model.SummarySpec{
			TargetCount:         1,
			SuccessCount:        1,
			AllAssignedDeployed: true,
			JobID:               jobID,
		},
		Time:           time.Now(),
		State:          model.SummaryStateDone,
		Generation:     strconv.Itoa(int(obj.GetGeneration())),
		DeploymentHash: hash,
	}
}

func MockSucessSummaryResult(obj reconcilers.Reconcilable, hash string) *model.SummaryResult {
	return &model.SummaryResult{
		Summary: model.SummarySpec{
			TargetCount:         1,
			SuccessCount:        1,
			AllAssignedDeployed: true,
		},
		Time:           time.Now(),
		State:          model.SummaryStateDone,
		Generation:     strconv.Itoa(int(obj.GetGeneration())),
		DeploymentHash: hash,
	}
}

func MockFailureSummaryResult(obj reconcilers.Reconcilable, hash string) *model.SummaryResult {
	return &model.SummaryResult{
		Summary: model.SummarySpec{
			TargetCount:         1,
			SuccessCount:        0,
			AllAssignedDeployed: false,
			TargetResults: map[string]model.TargetResultSpec{
				"default-target": {
					Status: "ErrorCode",
					ComponentResults: map[string]model.ComponentResultSpec{
						"comp1": {Status: v1alpha2.UpdateFailed, Message: "failed"},
						"comp2": {Status: v1alpha2.Accepted, Message: "untoched"},
					},
				},
			},
		},
		Time:           time.Now(),
		State:          model.SummaryStateDone,
		Generation:     strconv.Itoa(int(obj.GetGeneration())),
		DeploymentHash: hash,
	}
}

func MockInProgressSummaryResult(obj reconcilers.Reconcilable, hash string) *model.SummaryResult {
	return &model.SummaryResult{
		Summary: model.SummarySpec{
			TargetCount:         1,
			SuccessCount:        0,
			AllAssignedDeployed: false,
			TargetResults: map[string]model.TargetResultSpec{
				"default-target": {
					Status: "pending",
					ComponentResults: map[string]model.ComponentResultSpec{
						"comp1": {Status: v1alpha2.Updated, Message: "updated"},
						"comp2": {Status: v1alpha2.Accepted, Message: "pending"},
					},
				},
			},
		},
		Time:           time.Now(),
		State:          model.SummaryStateRunning,
		Generation:     strconv.Itoa(int(obj.GetGeneration())),
		DeploymentHash: hash,
	}
}

func MockInProgressDeleteSummaryResult(obj reconcilers.Reconcilable, hash string) *model.SummaryResult {
	return &model.SummaryResult{
		Summary: model.SummarySpec{
			TargetCount:         1,
			SuccessCount:        0,
			AllAssignedDeployed: false,
			TargetResults: map[string]model.TargetResultSpec{
				"default-target": {
					Status: "pending",
					ComponentResults: map[string]model.ComponentResultSpec{
						"comp1": {Status: v1alpha2.Updated, Message: "deleted"},
						"comp2": {Status: v1alpha2.Accepted, Message: "pending"},
					},
				},
			},
			IsRemoval: true,
		},
		Time:           time.Now(),
		State:          model.SummaryStateRunning,
		Generation:     strconv.Itoa(int(obj.GetGeneration())),
		DeploymentHash: hash,
	}
}

// GetSummary implements ApiClient.
func (c *MockApiClient) GetSummary(ctx context.Context, id string, name string, namespace string, user string, password string) (*model.SummaryResult, error) {
	args := c.Called(ctx, id, namespace)
	summary := args.Get(0)
	if summary == nil {
		return nil, args.Error(1)
	}
	return summary.(*model.SummaryResult), args.Error(1)
}

// DeleteSummary implements ApiClient.
func (c *MockApiClient) DeleteSummary(ctx context.Context, id string, namespace string, user string, password string) error {
	return nil
}

// QueueDeploymentJob implements utils.ApiClient.
func (c *MockApiClient) QueueDeploymentJob(ctx context.Context, namespace string, isDelete bool, deployment model.DeploymentSpec, user string, password string) error {
	args := c.Called(ctx, namespace, isDelete, deployment)
	return args.Error(0)
}

// CancelDeploymentJob implements utils.ApiClient.
func (c *MockApiClient) CancelDeploymentJob(ctx context.Context, id string, jobId string, namespace string, user string, password string) error {
	args := c.Called(ctx, namespace, id, jobId, namespace)
	return args.Error(0)
}

// QueueJob implements ApiClient.
// Deprecated and not used.
func (c *MockApiClient) QueueJob(ctx context.Context, id string, scope string, isDelete bool, isTarget bool, user string, password string) error {
	panic("implement me")
}

func CreateSimpleDeploymentBuilder() func(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
	return func(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
		return &model.DeploymentSpec{
			Hash: "test-hash",
		}, nil
	}
}

func createDeploymentBuilder(dr utils.DeploymentResources) func(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
	return func(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
		deployment, err := utils.CreateSymphonyDeployment(
			ctx,
			dr.Instance,
			dr.Solution,
			dr.TargetCandidates,
			TestNamespace,
		)
		return &deployment, err
	}
}

func BuildDefaultInstance() *solution_v1.Instance {
	return &solution_v1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultInstanceNamespacedName.Name,
			Namespace: DefaultInstanceNamespacedName.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Instance",
			APIVersion: solution_v1.GroupVersion.String(),
		},
		Spec: k8smodel.InstanceSpec{
			Target: model.TargetSelector{
				Name: DefaultTargetNamepsacedName.Name,
			},
			Solution: SolutionReferenceName,
		},
	}
}

func BuildDefaultTarget() *fabric_v1.Target {
	return &fabric_v1.Target{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultTargetNamepsacedName.Name,
			Namespace: DefaultTargetNamepsacedName.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Target",
			APIVersion: fabric_v1.GroupVersion.String(),
		},
		Spec: k8smodel.TargetSpec{
			Scope: TestScope,
		},
	}
}

func BuildDefaultSolution() *solution_v1.Solution {
	return &solution_v1.Solution{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultSolutionNamespacedName.Name,
			Namespace: DefaultSolutionNamespacedName.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Solution",
			APIVersion: solution_v1.GroupVersion.String(),
		},
		Spec: k8smodel.SolutionSpec{},
	}
}

func DefaultTestReconcilerOptions() []reconcilers.DeploymentReconcilerOptions {
	return []reconcilers.DeploymentReconcilerOptions{
		reconcilers.WithDeleteTimeOut(TestReconcileTimout),
		reconcilers.WithPollInterval(TestPollInterval),
		reconcilers.WithReconciliationInterval(TestReconcileInterval),
		reconcilers.WithFinalizerName(TestFinalizer),
		reconcilers.WithDeploymentBuilder(CreateSimpleDeploymentBuilder()),
	}
}

func BeWithin(durationString string) *TimeMatcher {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		panic(err)
	}
	return &TimeMatcher{
		wiggleroom: duration,
	}
}

func (m *TimeMatcher) Of(expected time.Duration) *TimeMatcher {
	m.duration = expected
	return m
}

func (m *TimeMatcher) Match(actual interface{}) (success bool, err error) {
	duration, ok := actual.(time.Duration)
	if !ok {
		return false, nil
	}
	return duration >= m.duration-m.wiggleroom && duration <= m.duration+m.wiggleroom, nil
}

func (m *TimeMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto be within\n\t%s\nof\n\t%s", actual, m.wiggleroom, m.duration)
}

func (m *TimeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be within\n\t%#v\nof\n\t%#v", actual, m.wiggleroom, m.duration)
}

func ToPointer[T any](v T) *T { return &v }
