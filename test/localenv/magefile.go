//go:build mage

/*
Use this tool to quickly get started developing in the symphony ecosystem. The
tool provides a set of common commands to make development easier for the team.
To get started using Minikube, run 'mage build minikube:start minikube:load deploy'.
*/
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/princjef/mageutil/shellcmd"
)

const (
	RELEASE_NAME           = "ecosystem"
	LOCAL_HOST_URL         = "http://localhost"
	OSS_CONTAINER_REGISTRY = "possprod.azurecr.io"
	NAMESPACE              = "default"
	DOCKER_TAG             = "latest"
	CHART_PATH             = "../../.azure/symphony-extension/helm/symphony"
)

var reWhiteSpace = regexp.MustCompile(`\n|\t| `)

type Minikube mg.Namespace
type Build mg.Namespace
type Pull mg.Namespace
type Cluster mg.Namespace
type Test mg.Namespace

/******************** Targets ********************/

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

// Deploys the symphony ecosystem to your local Minikube cluster.
func (Cluster) Deploy() error {
	return conditionalRun(
		func() error { //azure
			helmUpgrade := fmt.Sprintf("helm upgrade %s %s --install -n %s --create-namespace --wait -f ../../.azure/symphony-extension/helm/symphony/values.azure.yaml -f symphony-values.yaml", RELEASE_NAME, CHART_PATH, NAMESPACE)
			return shellcmd.Command(helmUpgrade).Run()
		},
		func() error { //oss
			helmUpgrade := fmt.Sprintf("helm upgrade %s %s --install -n %s --create-namespace --wait -f ../../.azure/symphony-extension/helm/symphony/values.yaml -f symphony-values.yaml --set symphonyImage.tag=%s --set paiImage.tag=%s", RELEASE_NAME, CHART_PATH, NAMESPACE, DOCKER_TAG, DOCKER_TAG)
			return shellcmd.Command(helmUpgrade).Run()
		})
}

// Up brings the minikube cluster up with symphony deployed
func Up() error {
	// Delete if a minikube cluster already exists
	mk := &Minikube{}
	_ = mk.Delete()

	c := &Cluster{}
	if err := c.Up(); err != nil {
		return err
	}

	data, err := ioutil.ReadFile("header.txt")
	if err == nil {
		fmt.Println(string(data))
	}

	fmt.Println("Press any key to shutdown")

	reader := bufio.NewReader(os.Stdin)
	_, _, _ = reader.ReadRune()

	fmt.Println("Cleaning up minikube cluster")

	if err := mk.Delete(); err != nil {
		return err
	}

	fmt.Println("done")

	return nil
}

// PullUp pulls the latest images and starts the local environment
func PullUp() error {
	mkTask := runBg(recreateMinikube)
	p := &Pull{}

	if err := p.All(); err != nil {
		return err
	}

	if err := runBgResult(mkTask); err != nil {
		return err
	}

	if err := Up(); err != nil {
		return err
	}

	return nil
}

// BuildUp builds the latest images and starts the local environment
func BuildUp() error {
	mkTask := runBg(recreateMinikube)
	b := &Build{}

	if err := b.All(); err != nil {
		return err
	}

	if err := runBgResult(mkTask); err != nil {
		return err
	}

	if err := Up(); err != nil {
		return err
	}

	return nil
}

// Uninstall all components, e.g. mage destroy all
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
func (Build) All() error {
	defer logTime(time.Now(), "build:all")

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

// Build api container
func (Build) Api() error {
	return buildAPI()
}
func buildAPI() error {
	return conditionalRun(
		func() error {
			return shellcmd.Command("docker compose -f ../../api/docker-compose.azure.yaml build").Run() //azure
		},
		func() error {
			return shellcmd.Command("docker compose -f ../../api/docker-compose.yaml build").Run() //oss
		})
}

func Azure() error {
	return nil
}

// Build k8s container
func (Build) K8s() error {
	return buildK8s()
}
func buildK8s() error {
	return conditionalRun(
		func() error {
			return shellcmd.Command("docker compose -f ../../k8s/docker-compose.azure.yaml build").Run() //azure
		},
		func() error {
			return shellcmd.Command("docker compose -f ../../k8s/docker-compose.yaml build").Run() //oss
		})
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

// Brings the cluster up with all images loaded
func (Cluster) Load() error {
	if err := ensureMinikubeUp(); err != nil {
		return err
	}

	mk := &Minikube{}
	if err := mk.Load(); err != nil {
		return err
	}

	return nil
}

// Brings the cluster up, loads the image and deploys
func (Cluster) Up() error {
	defer logTime(time.Now(), "cluster:up")

	// Install minikube
	c := &Cluster{}
	if err := c.Load(); err != nil {
		return err
	}

	if err := c.Deploy(); err != nil {
		return err
	}

	return nil
}

// Stop the cluster
func (Cluster) Down() error {
	mk := &Minikube{}
	return mk.Stop()
}

// Deploys the symphony ecosystem to minikube and waits for all pods to be ready.
// This is intended for use with the automated integration tests.
// Dev workflows can use more optimized commands.
func (Test) Up() error {
	defer logTime(time.Now(), "test:up")

	// Show the state of the cluster for CI scenarios
	// This should be shown even when an error occurs
	c := &Cluster{}
	defer c.Status()

	// Delete if a minikube cluster already exists
	mk := &Minikube{}
	_ = mk.Delete()

	// Build and load images without deploying
	// tests will run the deployment
	return c.Up()
}

// Show the state of the cluster for CI scenarios
func (Cluster) Status() {
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

// Pulls all docker images for symphony
func (Pull) All() error {
	defer logTime(time.Now(), "pull:all")

	if err := ACRLogin(); err != nil {
		return err
	}

	// Pull directly from ACR
	return shellcmd.RunAll(pull(
		"symphony-k8s",
		"symphony-api",
	)...)
}

// Pull symphony-k8s
func (Pull) K8s() error {
	if err := ACRLogin(); err != nil {
		return err
	}

	// Pull directly from ACR
	return shellcmd.RunAll(pull(
		"symphony-k8s",
	)...)
}

// Pull symphony-api
func (Pull) Api() error {
	if err := ACRLogin(); err != nil {
		return err
	}

	// Pull directly from ACR
	return shellcmd.RunAll(pull(
		"symphony-api",
	)...)
}

// Log into the ACR, prompt if az creds are expired
func ACRLogin() error {
	for i := 0; i < 3; i++ {
		err := shellcmd.Command.Run("az acr login --name possprod") //oss
		if err != nil {
			err := shellcmd.Command.Run("az login --use-device-code")
			if err != nil {
				return err
			}

			if i == 3 {
				return err
			}
		} else {
			return nil
		}
	}

	return nil
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
			OSS_CONTAINER_REGISTRY,
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
			OSS_CONTAINER_REGISTRY,
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
	c := &Cluster{}

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
				c.Status()
			}

			time.Sleep(time.Second)
		} else {
			fmt.Println("All pods are cleaned up: ", time.Since(startTime).String())
			return nil
		}

		time.Sleep(time.Second)
	}
}

// Run a command in the background
func runBg(f func() error) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		if err := f(); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

// Wait for an error or the channel to close
func runBgResult(errChan <-chan error) error {
	if errChan != nil {
		err, ok := <-errChan
		if !ok {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete and recreate minikube
func recreateMinikube() error {
	defer logTime(time.Now(), "recreate minikube")

	mk := &Minikube{}
	_ = mk.Delete()

	return ensureMinikubeUp()
}

// Ensure minikube is running, otherwise install and start it
func ensureMinikubeUp() error {
	defer logTime(time.Now(), "start minikube")

	if !minikubeRunning() {
		mk := &Minikube{}
		if err := mk.Install(); err != nil {
			return err
		}

		if err := mk.Start(); err != nil {
			return err
		}
	}

	if err := ensureMinikubeContext(); err != nil {
		return err
	}

	return nil
}

// True if minikube is active and running
func minikubeRunning() bool {
	b, err := shellcmd.Command.Output(`minikube status -f="{{.Host}}"`)
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(b)) == "Running"
}

// Set the kubectl context to minikube
func ensureMinikubeContext() error {
	return shellcmd.Command(`kubectl config use-context minikube`).Run()
}

func logTime(start time.Time, name string) {
	fmt.Printf("[DONE] (%s) '%s'\n", time.Since(start), name)
}
