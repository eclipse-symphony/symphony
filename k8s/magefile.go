//go:build mage

package main

import (
	"fmt"
	"os"

	//mage:import
	_ "github.com/azure/symphony/packages/mage"
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
		"latest",
	))

	kustomize = bintool.Must(bintool.New(
		"kustomize",
		"v4.5.7",
		"https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F{{.Version}}/kustomize_{{.Version}}_linux_amd64.tar.gz",
	))
)

func conditionalRun(azureFunc func() error, ossFunc func() error) error {
	if len(os.Args) > 2 && os.Args[len(os.Args)-1] == "azure" {
		return azureFunc()
	}
	return ossFunc()
}
func conditionalString(azureStr string, ossStr string) string {
	if len(os.Args) > 2 && os.Args[len(os.Args)-1] == "azure" {
		return azureStr
	}
	return ossStr
}

// Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition manifest.
func Manifests() error {
	mg.Deps(ensureControllerGen)
	return conditionalRun(
		func() error {
			return shellcmd.RunAll(
				shellcmd.Command("rm -rf config/azure/crd/bases"),
				controllerGen.Command("rbac:roleName=manager-role crd webhook paths=./apis/symphony.microsoft.com/v1 output:crd:artifacts:config=config/azure/crd/bases output:webhook:artifacts:config=config/azure/webhook"),
			)
		},
		func() error {
			return shellcmd.RunAll(
				shellcmd.Command("rm -rf config/oss/crd/bases"),
				controllerGen.Command("rbac:roleName=manager-role crd webhook paths=./apis/ai/v1 paths=./apis/fabric/v1 paths=./apis/solution/v1 paths=./apis/workflow/v1 paths=./apis/federation/v1 output:crd:artifacts:config=config/oss/crd/bases output:webhook:artifacts:config=config/oss/webhook"),
			)
		})
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

func Azure() error {
	//this is a hack to get around the fact that mage doesn't support passing args to targets
	return nil
}

// Build manager binary.
func Build() error {
	return conditionalRun(
		func() error {
			return shellcmd.Command("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager -tags=azure").Run()
		},
		func() error {
			return shellcmd.Command("CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager").Run()
		})
}

// Run a controller from your host.
func Run() error {
	return shellcmd.Command("go run ./main.go").Run()
}

func HelmTemplate() error {
	return conditionalRun(
		func() error {
			return kustomize.Command("build config/azure/helm -o ../symphony-extension/helm/symphony/templates/symphony.yaml").Run()
		},
		func() error {
			return kustomize.Command("build config/oss/helm -o ../symphony-extension/helm/symphony/templates/symphony.yaml").Run()
		})
}

// Install CRDs into the K8s cluster specified in ~/.kube/config.
func InstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return conditionalRun(
		func() error {
			return shellcmd.Command("kubectl apply -f config/azure/crd").Run()
		},
		func() error {
			return shellcmd.Command("kubectl apply -f config/oss/crd").Run()
		})
}

// Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
func UninstallCRDs() error {
	mg.Deps(ensureKustomize, Manifests)
	return conditionalRun(
		func() error {
			return shellcmd.Command("kubectl delete --ignore-not-found -f config/azure/crd/bases").Run()
		},
		func() error {
			return shellcmd.Command("kubectl delete --ignore-not-found -f config/oss/crd/bases").Run()
		})
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
