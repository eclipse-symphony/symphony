//go:build mage

package main

import (
	"fmt"
	"os"

	//mage:import
	_ "dev.azure.com/msazure/One/_git/symphony.git/packages/mage"
	"github.com/magefile/mage/mg"
	"github.com/princjef/mageutil/bintool"
	"github.com/princjef/mageutil/shellcmd"
)

const (
	EnvTestK8sVersion = "1.23"
	ImageRepository   = "symphony.azurecr.io/symphony-k8s"
)

var (
	controllerGen = bintool.Must(bintool.NewGo(
		"sigs.k8s.io/controller-tools/cmd/controller-gen",
		"v0.11.1",
	))

	envTest = bintool.Must(bintool.NewGo(
		"sigs.k8s.io/controller-runtime/tools/setup-envtest",
		"latest",
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
		shellcmd.Command("rm -rf config/crd/bases"),
		controllerGen.Command("rbac:roleName=manager-role crd webhook paths=./... output:crd:artifacts:config=config/crd/bases"),
	)
}

// Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
func Generate() error {
	mg.Deps(ensureControllerGen)
	return controllerGen.Command("object:headerFile=hack/boilerplate.go.txt paths=./...").Run()
}

// Run tests.
func OperatorTest() error {
	mg.Deps(ensureEnvTest)
	assets, err := envTest.Command(fmt.Sprintf("use %s -p path", EnvTestK8sVersion)).Output()
	if err != nil {
		return err
	}
	os.Setenv("KUBEBUILDER_ASSETS", string(assets))

	return shellcmd.Command("go test ./... -race -v -coverprofile cover.out").Run()
}

// Build manager binary.
func Build() error {
	return shellcmd.Command("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager main.go").Run()
}

// Run a controller from your host.
func Run() error {
	return shellcmd.Command("go run ./main.go").Run()
}

// Generate helm template.
func HelmTemplate() error {
	mg.Deps(ensureKustomize, Manifests)
	return kustomize.Command("build config/helm -o ../symphony-extension/helm/symphony/templates/symphony.yaml").Run()
}

// Install CRDs into the K8s cluster specified in ~/.kube/config.
func InstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return shellcmd.Command("kubectl apply -f config/crd").Run()
}

// Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
func UninstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return shellcmd.Command("kubectl delete --ignore-not-found -f config/crd/bases").Run()
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
