/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	api "gopls-workspace/apis/solution/v1"
	internalTesting "gopls-workspace/testing"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	internalTesting.RunGinkgoSpecs(t, "Controller Suite")
}

func TestUnmarshalSolution(t *testing.T) {
	t.Skip("Skipping tests for now as they are no longer relevant")
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

var _ = BeforeSuite(func() {
	Skip("Skipping tests for now as they are no longer relevant")
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = api.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
