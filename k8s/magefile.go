//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"fmt"
	"os"

	//mage:import
	base "github.com/eclipse-symphony/symphony/packages/mage"
	"github.com/magefile/mage/mg"
	"github.com/princjef/mageutil/bintool"
	"github.com/princjef/mageutil/shellcmd"
)

const (
	EnvTestK8sVersion = "1.23"
)

var (
	controllerGen = bintool.Must(bintool.NewGo(
		"sigs.k8s.io/controller-tools/cmd/controller-gen",
		"v0.11.1",
	))

	envTest = bintool.Must(bintool.NewGo(
		"sigs.k8s.io/controller-runtime/tools/setup-envtest",
		"v0.0.0-20240320141353-395cfc7486e6",
	))

	kustomize = bintool.Must(bintool.New(
		"kustomize",
		"v4.5.7",
		"https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F{{.Version}}/kustomize_{{.Version}}_linux_amd64.tar.gz",
	))
)

// Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition manifest.
func Manifests() error {
	mg.Deps(ensureControllerGen)
	return shellcmd.RunAll(
		shellcmd.Command("rm -rf config/oss/crd/bases"),
		controllerGen.Command("rbac:roleName=manager-role crd webhook paths=./apis/ai/v1 paths=./apis/fabric/v1 paths=./apis/solution/v1 paths=./apis/workflow/v1 paths=./apis/federation/v1 output:crd:artifacts:config=config/oss/crd/bases output:webhook:artifacts:config=config/oss/webhook output:rbac:artifacts:config=config/oss/rbac"),
	)

}

// Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
func Generate() error {
	mg.Deps(ensureControllerGen)
	return shellcmd.RunAll(
		controllerGen.Command("object:headerFile=hack/boilerplate.go.txt paths=./..."),
		controllerGen.Command("object:headerFile=hack/boilerplate.go.txt paths=../api/pkg/apis/v1alpha1/model"),
	)
}

// Run suites and unit tests in k8s.
func OperatorTest() error {
	mg.Deps(ensureEnvTest)
	assets, err := envTest.Command(fmt.Sprintf("use %s -p path", EnvTestK8sVersion)).Output()
	if err != nil {
		return err
	}
	os.Setenv("KUBEBUILDER_ASSETS", string(assets))

	return base.RunUnitTestAndSuiteTest()
}

// Run unit tests in k8s.
func OperatorUnitTest() error {
	mg.Deps(ensureEnvTest)

	assets, err := envTest.Command(fmt.Sprintf("use %s -p path", EnvTestK8sVersion)).Output()
	if err != nil {
		return err
	}

	os.Setenv("KUBEBUILDER_ASSETS", string(assets))

	return base.UnitTest()
}

// Build manager binary.
func Build() error {
	return shellcmd.Command("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager").Run()
}

// Run a controller from your host.
func Run() error {
	return shellcmd.Command("go run ./main.go").Run()
}

// Kustomize startup symphony yaml for helm chart.
func HelmTemplate() error {
	mg.Deps(ensureKustomize, Manifests)
	return kustomize.Command("build config/oss/helm -o ../packages/helm/symphony/templates/symphony-core/symphonyk8s.yaml").Run()
}

// Install CRDs into the K8s cluster specified in ~/.kube/config.
func InstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return shellcmd.Command("kubectl apply -f config/oss/crd").Run()
}

// Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
func UninstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return shellcmd.Command("kubectl delete --ignore-not-found -f config/oss/crd/bases").Run()
}

func ensureControllerGen() error {
	return controllerGen.Ensure()
}

func ensureEnvTest() error {
	return envTest.Ensure()
}

func ensureKustomize() error {
	return kustomize.Ensure()
}
