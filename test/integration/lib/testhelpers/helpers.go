package testhelpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
)

func SetupClusterWithTunnel() (context.CancelFunc, int, error) {
	err := SetupCluster()
	if err != nil {
		return nil, -1, err
	}

	// Create tunnel
	fmt.Println("Creating minikube tunnel....")
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "minikube", "tunnel")
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to create minikube tunnel.")
		return cancel, -1, err
	}
	fmt.Printf("Minikube tunnel started with PID: %d, starting another thread to wait\n", cmd.Process.Pid)
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("minikube tunnel stopped: %s\n", err)
		}
	}()
	return cancel, cmd.Process.Pid, nil
}

func DumpClusterState(ctx context.Context) {
	shell.Exec(ctx, "kubectl get all -A -o wide")
	shell.Exec(ctx, "kubectl get events -A --sort-by=.metadata.creationTimestamp")
	shell.Exec(ctx, "kubectl get targets.fabric.symphony -A -o yaml")
	shell.Exec(ctx, "kubectl get solutionversions.solution.symphony -A -o yaml")
	shell.Exec(ctx, "kubectl get instances.solution.symphony -A -o yaml")
	shell.Exec(ctx, "helm list -A -o yaml")
}

func CleanupManifests(ctx context.Context) error {
	return shell.ExecAll(
		ctx,
		"kubectl delete instances.solution.symphony --all -A",
		"kubectl delete targets.fabric.symphony --all -A",
		"kubectl delete solutionversions.solution.symphony --all -A",
	)
}

// Run a command with | or other things that do not work in shellcmd
func ShellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string, flavor string) error {
	return ShellExec(fmt.Sprintf("cd ../../../localenv && mage %s %s", mageCmd, flavor))
}

func SetClusterWithSetting(settings string) error {
	err := localenvCmd("cluster:deployWithSettings", settings)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 10)
	return nil
}

// Prepare the cluster
// Run this manually to prepare your local environment for testing/debugging
func SetupCluster() error {
	// Deploy symphony
	err := localenvCmd("cluster:deploy", "")
	if err != nil {
		return err
	}

	// Wait until the admission webhooks are actually serving. cert-manager's
	// cainjector updates the webhook caBundle asynchronously; a fixed sleep is
	// not enough on a loaded runner (observed as "InternalError: failed calling
	// webhook" on the first CR apply after deploy).
	return waitForWebhooksReady()
}

// waitForWebhooksReady probes the admission webhooks with a server dry-run
// apply of a minimal Target. Output containing "failed calling webhook" or
// "InternalError" means the webhook/caBundle is not ready yet; anything else
// (including a validation rejection, which proves the webhook answered) is
// treated as ready. The dry-run persists nothing.
func waitForWebhooksReady() error {
	probeManifest := `apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: webhook-readiness-probe
  namespace: default
spec:
  displayName: probe
`
	probeFile := filepath.Join(os.TempDir(), "symphony-webhook-probe.yaml")
	if err := os.WriteFile(probeFile, []byte(probeManifest), 0644); err != nil {
		return err
	}
	defer os.Remove(probeFile)

	ctx := context.Background()
	for i := 0; i < 24; i++ {
		out, _ := shell.Output(ctx, fmt.Sprintf("kubectl apply --dry-run=server -f %s 2>&1", probeFile))
		if !strings.Contains(string(out), "failed calling webhook") && !strings.Contains(string(out), "InternalError") {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for admission webhooks to become ready")
}

// Clean up
func Cleanup(testName string) {
	localenvCmd(fmt.Sprintf("dumpSymphonyLogsForTest '%s'", testName), "")
	localenvCmd("destroy all,nowait", "")
}

func CleanupWithTunnel(cancel context.CancelFunc, tunnelPid int, testName string) {
	Cleanup(testName)
	fmt.Println("Cancelling minikube tunnel....")
	cancel()

	fmt.Println("Waiting 5 seconds for tunnel to stop....")
	time.Sleep(time.Second * 5)
	if tunnelPid != -1 {
		// check if the tunnel is still running
		if isProcessRunning(tunnelPid) {
			// kill the tunnel
			fmt.Println("Tunnel is still running, killing it....")
			ShellExec(fmt.Sprintf("kill -9 %d", tunnelPid))
		}
	}
}

// Check if a process is running by its PID
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to the process
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func IsTestInAzure() bool {
	// Check if the environment variable is set to "true" (case-insensitive)
	return strings.EqualFold(os.Getenv("AZURE_TEST"), "true")
}

func WriteYamlStringsToFile(yamlString string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(yamlString))
	if err != nil {
		return err
	}

	return nil
}

func ReplacePlaceHolderInManifestWithString(manifest string, targetName string, solutionversionContainerName string, solutionversionName string, instanceName string, historyName string) (string, error) {
	fullPath, err := filepath.Abs(manifest)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	stringYaml := string(data)
	if IsTestInAzure() {
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONREFNAME",
			"/subscriptions/aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e/resourcegroups/test-rg/providers/microsoft.edge/targets/TARGETNAME/solutionversions/SOLUTIONCONTAINERNAME/versions/SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "TARGETREFNAME", "/subscriptions/aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e/resourcegroups/test-rg/providers/microsoft.edge/targets/TARGETNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME-v-INSTANCENAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME-v-SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEHISTORYFULLNAME", "TARGETNAME-v-SOLUTIONCONTAINERNAME-v-INSTANCENAME-v-HISTORYNAME")
	} else {
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONREFNAME", "SOLUTIONCONTAINERNAME:SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "TARGETREFNAME", "TARGETNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEFULLNAME", "INSTANCENAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONFULLNAME", "SOLUTIONCONTAINERNAME-v-SOLUTIONNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERFULLNAME", "SOLUTIONCONTAINERNAME")
		stringYaml = strings.ReplaceAll(stringYaml, "INSTANCEHISTORYFULLNAME", "INSTANCENAME-v-HISTORYNAME")
	}
	stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONCONTAINERNAME", solutionversionContainerName)
	stringYaml = strings.ReplaceAll(stringYaml, "INSTANCENAME", instanceName)
	stringYaml = strings.ReplaceAll(stringYaml, "TARGETNAME", targetName)
	stringYaml = strings.ReplaceAll(stringYaml, "SOLUTIONNAME", solutionversionName)
	stringYaml = strings.ReplaceAll(stringYaml, "HISTORYNAME", historyName)
	return stringYaml, nil
}

func ReplacePlaceHolderInManifest(manifest string, targetName string, solutionversionContainerName string, solutionversionName string, instanceName string, historyName string) error {
	fullPath, err := filepath.Abs(manifest)
	if err != nil {
		return err
	}
	stringYaml, err := ReplacePlaceHolderInManifestWithString(manifest, targetName, solutionversionContainerName, solutionversionName, instanceName, historyName)
	if err != nil {
		return err
	}
	err = WriteYamlStringsToFile(stringYaml, fullPath)
	return err
}
