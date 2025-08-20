package testhelpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	shell.Exec(ctx, "kubectl get solutions.solution.symphony -A -o yaml")
	shell.Exec(ctx, "kubectl get instances.solution.symphony -A -o yaml")
	shell.Exec(ctx, "helm list -A -o yaml")
}

func CleanupManifests(ctx context.Context) error {
	return shell.ExecAll(
		ctx,
		"kubectl delete instances.solution.symphony --all -A",
		"kubectl delete targets.fabric.symphony --all -A",
		"kubectl delete solutions.solution.symphony --all -A",
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

// Prepare the cluster
// Run this manually to prepare your local environment for testing/debugging
func SetupCluster() error {
	// Deploy symphony
	err := localenvCmd("cluster:deploy", "")
	if err != nil {
		return err
	}

	// Wait a few secs for symphony cert to be ready;
	// otherwise we will see error when creating symphony manifests in the cluster
	// <Error from server (InternalError): error when creating
	// "/mnt/vss/_work/1/s/test/integration/scenarios/basic/manifest/target.yaml":
	// Internal error occurred: failed calling webhook "mtarget.kb.io": failed to
	// call webhook: Post
	// "https://symphony-webhook-service.default.svc:443/mutate-symphony-microsoft-com-v1-target?timeout=10s":
	// x509: certificate signed by unknown authority>
	time.Sleep(time.Second * 10)
	return nil
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
