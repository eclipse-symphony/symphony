/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	. "gopls-workspace/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	api "gopls-workspace/apis/solution/v1"
	controllers "gopls-workspace/controllers/solution"

	ctrl "sigs.k8s.io/controller-runtime"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestAPIs(t *testing.T) {
	//RegisterFailHandler(Fail)

	//RunGinkgoSpecs(t, "Controller Suite")
}

func TestUnmarshalSolution(t *testing.T) {
	solutionYaml := `apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: sample-staged-solution
spec:  
  components:
  - name: staged-component
    properties:
      foo: "bar"
      bar:
        baz: "qux"
`
	solution := &api.Solution{}
	err := yaml.Unmarshal([]byte(solutionYaml), solution)
	assert.NoError(t, err)

	expectedProperties := map[string]interface{}{
		"foo": "bar",
		"bar": map[string]interface{}{
			"baz": "qux",
		},
	}
	actualProperties := map[string]interface{}{}
	err = json.Unmarshal(solution.Spec.Components[0].Properties.Raw, &actualProperties)
	assert.NoError(t, err)

	assert.Equal(t, expectedProperties, actualProperties)
}

// This test is here for legacy reasons. It spins up a test environment
// with a kubernetes api server and etcd.
// It'll continue to live here for now but all integrations tests should be
// created in the /test/integration directory.
//
// Don't write any tests that use the testEnv or k8sClient in this package
// unless there's a strong reason to.
//
// The tests in the target_controller_test.go and instance_controller_test.go
// are now unit tests that use a mocked k8sClient. All behavior that requires
// a manager and full operator pattern should be tested in the integration tests
var _ = Describe("Legacy testing with envtest", Ordered, func() {
	var (
		cancel    context.CancelFunc
		cfg       *rest.Config
		k8sClient client.Client
		testEnv   *envtest.Environment
		ctx       context.Context
		apiClient *MockApiClient
	)

	BeforeAll(func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseFlagOptions(&zap.Options{Development: true, TimeEncoder: zapcore.ISO8601TimeEncoder})))
		ctx, cancel = context.WithCancel(context.TODO())
		apiClient = &MockApiClient{}

		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "oss", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}

		var err error
		// cfg is defined in this file globally.
		cfg, err = testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		//+kubebuilder:scaffold:scheme
		err = api.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		// The default client created by calling client.New behaves slightly different
		// from the client created by the manager. The manager's client has preserves
		// the type of the object passed to it. The default client does not. So when you
		// make a get call with the default client, the object returned doesn't Have the
		// GroupVersionKind set. This would cause some certain assersions to fail as some
		// objects are prepared and queried outside the reconciler for test assertions
		// for this reaseon, we use the manager's client for all tests.
		//
		// k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			// needs to disable metrics otherwise all controller suite tests will try to bind to the same port (8080)
			MetricsBindAddress: "0",
		})
		Expect(err).ToNot(HaveOccurred())
		k8sClient = k8sManager.GetClient()
		Expect(k8sClient).NotTo(BeNil())

		apiClient.On("GetSummary", mock.Anything, mock.Anything, mock.Anything).Return(MockSucessSummaryResult(BuildDefaultTarget(), ""), nil)
		apiClient.On("QueueDeploymentJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err = (&controllers.InstanceReconciler{
			Client:                 k8sManager.GetClient(),
			Scheme:                 k8sManager.GetScheme(),
			ReconciliationInterval: 2 * time.Second,
			DeleteTimeOut:          6 * time.Second,
			PollInterval:           1 * time.Second,
			ApiClient:              apiClient,
		}).SetupWithManager(k8sManager)
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			err = k8sManager.Start(ctx)
			Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		}()

	})

	AfterAll(func() {
		cancel()
		By("tearing down the test environment")
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	// This doesn't actually test reconcililiation. It only tests that the resources
	// can be created in the cluster. The reconciler tests are in the target_controller_test.go and
	// instance_controller_test.go files
	It("should be able to create valid resources", func() {
		By("creating a valid instance")
		Expect(k8sClient.Create(ctx, BuildDefaultInstance())).To(Succeed())
	})
})
