//go:build mage

/*
Use this tool to quickly get started developing in the symphony ecosystem. The
tool provides a set of common commands to make development easier for the team.
To get started using Minikube, run 'mage build minikube:start minikube:load deploy'.
*/
package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/princjef/mageutil/shellcmd"
)

const (
	RELEASE_NAME       = "ecosystem"
	LOCAL_HOST_URL     = "http://localhost"
	CONTAINER_REGISTRY = "symphonycr.azurecr.io"
	NAMESPACE          = "default"
	DOCKER_TAG         = "latest"
	CHART_PATH         = "../symphony-extension/helm/symphony"
)

var reWhiteSpace = regexp.MustCompile(`\n|\t| `)

type Minikube mg.Namespace

/******************** Targets ********************/

// Deploys the symphony ecosystem to your local Minikube cluster.
func Deploy() error {
	helmUpgrade := fmt.Sprintf("helm upgrade %s %s --install -n %s --create-namespace --wait -f symphony-values.yaml", RELEASE_NAME, CHART_PATH, NAMESPACE)
	return shellcmd.Command(helmUpgrade).Run()
}

// Uninstall all components
func Destroy(flags string) error {
	err := shellcmd.RunAll(
		shellcmd.Command(fmt.Sprintf("helm uninstall %s -n %s --wait", RELEASE_NAME, NAMESPACE)),
	)
	if err != nil {
		return err
	}

	// to indicate if we should wait for cleanup to finish
	shouldWait := true
	for _, flag := range strings.Split(reWhiteSpace.ReplaceAllString(strings.ToLower(flags), ``), ",") {
		if flag == "nowait" {
			shouldWait = false
		}
	}

	if shouldWait {
		// Wait for all pods to go away
		if err := waitForServiceCleanup(); err != nil {
			return err
		}
	}

	return nil
}

// Build builds all containers
func Build() error {
	err := buildAPI()
	if err != nil {
		return err
	}

	err = buildK8s()
	if err != nil {
		return err
	}

	return nil
}

func buildAPI() error {
	return shellcmd.Command("docker-compose -f ../api/docker-compose.yaml build").Run()
}

func buildK8s() error {
	return shellcmd.Command("docker-compose -f ../k8s/docker-compose.yaml build").Run()
}

/******************** Minikube ********************/

// Installs the Minikube binary on your machine.
func (Minikube) Install() error {
	whereIsMinikube, err := shellcmd.Command("whereis minikube").Output()
	if err != nil {
		return err
	}

	// Normalize 'whereis' command output to identify if Minikube is installed
	if reWhiteSpace.ReplaceAllString(string(whereIsMinikube), ``) != "minikube:" {
		return shellcmd.Command("minikube version").Run()
	}

	err = shellcmd.Command(`curl -o "minikube" -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64`).Run()
	if err != nil {
		return err
	}

	err = shellcmd.Command(`sudo install "minikube" /usr/local/bin/minikube`).Run()
	if err != nil {
		return err
	}

	err = shellcmd.Command(`rm minikube`).Run()
	if err != nil {
		return err
	}

	return nil
}

// Starts the Minikube cluster w/ select addons.
func (Minikube) Start() error {
	err := shellcmd.Command("minikube start").Run()
	if err != nil {
		return err
	}

	err = shellcmd.Command("minikube addons enable metrics-server").Run()
	if err != nil {
		return err
	}

	return nil
}

// Stops the Minikube cluster.
func (Minikube) Stop() error {
	return shellcmd.Command("minikube stop").Run()
}

// Loads symphony component docker images onto the Minikube VM.
func (Minikube) Load() error {
	return shellcmd.RunAll(load(
		fmt.Sprintf("symphony-api:%s", DOCKER_TAG),
		fmt.Sprintf("symphony-k8s:%s", DOCKER_TAG))...)
}

// Deletes the Minikube cluster from you dev box.
func (Minikube) Delete() error {
	return shellcmd.Command("minikube delete").Run()
}

// ClusterUp brings the cluster up with all images loaded
// but does not deploy.
func ClusterUp() error {
	// Install minikube
	mk := &Minikube{}
	err := mk.Install()
	if err != nil {
		return err
	}

	// Start minikube and load containers
	err = mk.Start()
	if err != nil {
		return err
	}

	err = Build()
	if err != nil {
		return err
	}

	err = mk.Load()
	if err != nil {
		return err
	}

	return nil
}

// Up brings the minikube cluster up with symphony deployed
func Up() error {
	err := ClusterUp()
	if err != nil {
		return err
	}

	// Deploy the helm chart and wait for all pods to become ready
	err = Deploy()
	if err != nil {
		return err
	}

	return nil
}

// UpClean deletes minikube if it exists and then runs Up
func UpClean() error {
	// Delete if a minikube cluster already exists
	mk := &Minikube{}
	_ = mk.Delete()

	// Run
	return Up()
}

// Deploys the symphony ecosystem to minikube and waits for all pods to be ready.
// This is intended for use with the automated integration tests.
// Dev workflows can use more optimized commands.
func SetupIntegrationTests() error {
	// Show the state of the cluster for CI scenarios
	// This should be shown even when an error occurs
	defer ClusterStatus()

	// Delete if a minikube cluster already exists
	mk := &Minikube{}
	_ = mk.Delete()

	// Build and load images without deploying
	// tests will run the deployment
	return ClusterUp()
}

// Show the state of the cluster for CI scenarios
func ClusterStatus() {
	fmt.Println("*******************[Cluster]**********************")
	shellcmd.Command("helm list --all").Run()
	shellcmd.Command("kubectl get pods -A -o wide").Run()
	shellcmd.Command("kubectl get deployments -A -o wide").Run()
	shellcmd.Command("kubectl get services -A -o wide").Run()
	shellcmd.Command("kubectl top pod -A").Run()
	shellcmd.Command("kubectl get events -A").Run()

	fmt.Println("Describing failed pods")
	dumpShellOutput(fmt.Sprintf("kubectl get pods --all-namespaces | grep -E 'CrashLoopBackOff|Error|ImagePullBackOff|InvalidImageName|Pending' | awk '{print $2}' | xargs -I {} kubectl describe pod {} -n %s", NAMESPACE))
	dumpShellOutput(fmt.Sprintf("kubectl get pods --all-namespaces | grep -E 'CrashLoopBackOff|Error|ImagePullBackOff|InvalidImageName|Pending' | awk '{print $2}' | xargs -I {} kubectl logs {} -n %s", NAMESPACE))
	fmt.Println("**************************************************")
}

// Launch the Minikube Kubernetes dashboard.
func (Minikube) Dashboard() error {
	return shellcmd.Command("minikube dashboard").Run()
}

/******************** Helpers ********************/
func browserOpen(url string) error {
	openBrowser := fmt.Sprintf("xdg-open %s", url)
	return shellcmd.Command(openBrowser).Run()
}

// runParallel parallelizes running the commands
// this will print out all errors and return only the last error
func runParallel(commands ...shellcmd.Command) error {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(commands))

	// latest error seen
	var finalErr error
	for _, cmd := range commands {
		go func(cmd shellcmd.Command) {
			defer waitGroup.Done()
			start := time.Now()

			fmt.Printf("[START] '%s'\n", cmd)

			if err := cmd.Run(); err != nil {
				finalErr = err
				fmt.Println(err)
			}

			fmt.Printf("[DONE] (%s) '%s'\n", time.Since(start), cmd)
		}(cmd)
	}

	waitGroup.Wait()
	return finalErr
}

func load(names ...string) []shellcmd.Command {
	loads := make([]shellcmd.Command, len(names))

	for i, name := range names {
		loads[i] = shellcmd.Command(fmt.Sprintf(
			"minikube image load %s/%s",
			CONTAINER_REGISTRY,
			name,
		))
	}

	return loads
}

func pull(names ...string) []shellcmd.Command {
	loads := make([]shellcmd.Command, len(names))

	for i, name := range names {
		loads[i] = shellcmd.Command(fmt.Sprintf(
			"docker pull %s/%s",
			CONTAINER_REGISTRY,
			name,
		))
	}

	return loads
}

// Run a command with | or other things that do not work in shellcmd
func dumpShellOutput(cmd string) error {
	fmt.Println("> ", cmd)

	b, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Println("failed to run command", err)
		return err
	} else {
		fmt.Println(string(b))
	}

	return nil
}

// Wait for cleanup to finish
func waitForServiceCleanup() error {
	var startTime = time.Now()

	fmt.Println("Waiting for all pods to go away...")

	loopCount := 0

	for {
		loopCount++
		if loopCount == 600 {
			return fmt.Errorf("Failed to clean up all the resources!")
		}

		o, err := shellcmd.Command.Output(`kubectl get pods -A --no-headers`)
		if err != nil {
			return err
		}

		pods := strings.Split(strings.TrimSpace(string(o)), "\n")
		notReady := make([]string, 0)

		for _, pod := range pods {
			if len(strings.TrimSpace(pod)) > 3 && !strings.Contains(pod, "kube-system") {
				parts := strings.Split(pod, " ")
				name := pod
				if len(parts) >= 2 {
					name = parts[1]
				}
				notReady = append(notReady, name)
			}
		}

		if len(notReady) > 0 {
			// Show pods that aren't ready
			if loopCount%30 == 0 {
				fmt.Printf("waiting for pod removal. Try: %d Not ready: %s\n", loopCount, strings.Join(notReady, ", "))
			}

			// Show complete status every 5 minutes to help debug
			if loopCount%300 == 0 {
				ClusterStatus()
			}

			time.Sleep(time.Second)
		} else {
			fmt.Println("All pods are cleaned up: ", time.Since(startTime).String())
			return nil
		}

		time.Sleep(time.Second)
	}
}
