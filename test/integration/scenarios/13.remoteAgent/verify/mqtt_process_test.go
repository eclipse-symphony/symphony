package verify

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent/utils"
	"github.com/stretchr/testify/require"
)

func TestE2EMQTTCommunicationWithProcess(t *testing.T) {
	// Test configuration
	targetName := "test-mqtt-process-target"
	namespace := "default"
	mqttBrokerPort := 8883

	// Clean up any stale processes from previous test runs
	utils.CleanupStaleRemoteAgentProcesses(t)

	// Setup test environment
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running MQTT process test in: %s", testDir)

	// Step 1: Start fresh minikube cluster
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Setup test namespace
	setupMQTTProcessNamespace(t, namespace)
	defer utils.CleanupNamespace(t, namespace)

	var configPath, topologyPath, targetYamlPath string
	var processCmd *exec.Cmd
	var config utils.TestConfig
	var detectedBrokerAddress string

	// Use our new MQTT process test setup function with detected broker address
	t.Run("SetupMQTTProcessTestWithDetectedAddress", func(t *testing.T) {
		config, detectedBrokerAddress = utils.SetupMQTTProcessTestWithDetectedAddress(t, testDir, targetName, namespace)
		t.Logf("MQTT process test setup completed with broker address: %s", detectedBrokerAddress)
	})

	t.Run("StartSymphonyWithMQTTConfig", func(t *testing.T) {
		// Deploy Symphony with MQTT configuration using detected broker address
		symphonyBrokerAddress := fmt.Sprintf("tls://%s:%d", detectedBrokerAddress, mqttBrokerPort)
		t.Logf("Starting Symphony with MQTT broker address: %s", symphonyBrokerAddress)
		utils.StartSymphonyWithMQTTConfig(t, symphonyBrokerAddress)

		// Wait for Symphony server certificate to be created
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	// Create test configurations AFTER Symphony is running
	t.Run("CreateTestConfigurations", func(t *testing.T) {
		// Use the config path that was already created with the correct broker address
		configPath = config.ConfigPath
		topologyPath = config.TopologyPath
		fmt.Printf("Topology path: %s", topologyPath)
		targetYamlPath = utils.CreateTargetYAML(t, testDir, targetName, namespace)
		fmt.Printf("Target YAML path: %s", targetYamlPath)
		// Apply Target YAML to create the target resource
		err := utils.ApplyKubernetesManifest(t, targetYamlPath)
		require.NoError(t, err)

		// Wait for target to be created
		utils.WaitForTargetCreated(t, targetName, namespace, 30*time.Second)
	})

	// Start the remote agent process at main test level so it persists across subtests
	t.Logf("Starting MQTT remote agent process...")
	// The config was already properly set up in SetupMQTTProcessTestWithDetectedAddress
	// Just update the paths that were created in CreateTestConfigurations
	config.ConfigPath = configPath
	config.TopologyPath = topologyPath
	fmt.Printf("Starting remote agent process with config: %+v\n", config)

	// Start remote agent using direct process (no systemd service) without automatic cleanup
	processCmd = utils.StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	// Set up cleanup at main test level to ensure process runs for entire test
	t.Cleanup(func() {
		if processCmd != nil {
			t.Logf("Cleaning up MQTT remote agent process from main test...")
			utils.CleanupRemoteAgentProcess(t, processCmd)
		} else {
			t.Logf("No process to cleanup (processCmd is nil)")
		}
	})

	// Also set up a signal handler for immediate cleanup on test interruption
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Test panicked, performing emergency cleanup: %v", r)
			if processCmd != nil {
				utils.CleanupRemoteAgentProcess(t, processCmd)
			}
			panic(r) // Re-panic after cleanup
		}
	}()

	// Add process monitoring to detect early exits
	processExited := make(chan bool, 1)
	go func() {
		processCmd.Wait()
		processExited <- true
	}()

	// Wait for process to be ready and healthy
	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
	t.Logf("MQTT remote agent process started successfully and will persist across all subtests")

	// Additional monitoring: check process didn't exit early
	select {
	case <-processExited:
		t.Fatalf("Remote agent process exited unexpectedly during startup")
	case <-time.After(2 * time.Second):
		// Process is still running after health check + buffer time
		t.Logf("Process stability confirmed - continuing with tests")
	}

	// Start continuous process monitoring throughout the test
	processMonitoring := make(chan bool, 1)
	go func() {
		defer close(processMonitoring)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-processExited:
				t.Logf("WARNING: Remote agent process exited during test execution")
				return
			case <-ticker.C:
				if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
					t.Logf("WARNING: Remote agent process has exited (state: %s)", processCmd.ProcessState.String())
					return
				}
				t.Logf("Process monitoring: Remote agent PID %d is still running", processCmd.Process.Pid)
			}
		}
	}()

	// Set up cleanup for the monitoring goroutine
	t.Cleanup(func() {
		// Stop the monitoring
		if processMonitoring != nil {
			close(processMonitoring)
		}
	})

	t.Run("VerifyProcessStarted", func(t *testing.T) {
		// Just verify the process is running
		require.NotNil(t, processCmd)
		require.NotNil(t, processCmd.Process)

		// Check if process has already exited
		if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
			t.Fatalf("Remote agent process has already exited: %s", processCmd.ProcessState.String())
		}

		// Additional check: try to send a harmless signal to verify process is alive
		if err := processCmd.Process.Signal(syscall.Signal(0)); err != nil {
			t.Fatalf("Process is not responding to signals (likely dead): %v", err)
		}

		t.Logf("MQTT remote agent process verified running with PID: %d", processCmd.Process.Pid)

		// Log current process status for debugging
		t.Logf("Process state: running=%t, exited=%t",
			processCmd.ProcessState == nil,
			processCmd.ProcessState != nil && processCmd.ProcessState.Exited())
	})

	t.Run("VerifyTargetStatus", func(t *testing.T) {
		// First check if our process is still running
		if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
			t.Fatalf("Remote agent process exited before target verification: %s", processCmd.ProcessState.String())
		}

		// Wait for target to reach ready state
		utils.WaitForTargetReady(t, targetName, namespace, 360*time.Second)

		// Check again after waiting - process should still be running
		if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
			t.Logf("WARNING: Remote agent process exited during target status verification: %s", processCmd.ProcessState.String())
		}
	})

	t.Run("VerifyMQTTProcessDataInteraction", func(t *testing.T) {
		// Verify process is still running before starting data interaction test
		if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
			t.Fatalf("Remote agent process exited before data interaction test: %s", processCmd.ProcessState.String())
		}

		// Verify that data flows through MQTT correctly
		// This would check that the remote agent successfully communicates
		// with Symphony through the MQTT broker
		testMQTTProcessDataInteraction(t, targetName, namespace, testDir)

		// Final check - process should still be running after all tests
		if processCmd.ProcessState != nil && processCmd.ProcessState.Exited() {
			t.Logf("WARNING: Remote agent process exited during data interaction test: %s", processCmd.ProcessState.String())
		} else {
			t.Logf("SUCCESS: Remote agent process survived all tests and is still running")
		}
	})

	// Cleanup
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "remote-agent-mqtt-process-test")
		utils.CleanupExternalMQTTBroker(t) // Use external broker cleanup
		utils.CleanupMQTTCASecret(t, "mqtt-ca")
		utils.CleanupMQTTClientSecret(t, namespace, "mqtt-client-secret")
	})

	t.Logf("MQTT communication test with direct process completed successfully")
}

func setupMQTTProcessNamespace(t *testing.T, namespace string) {
	// Create namespace if it doesn't exist
	_, err := utils.GetKubeClient()
	if err != nil {
		t.Logf("Warning: Could not get kube client to create namespace: %v", err)
		return
	}

	nsYaml := fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)

	nsPath := filepath.Join(utils.SetupTestDirectory(t), "namespace.yaml")
	err = utils.CreateYAMLFile(t, nsPath, nsYaml)
	if err == nil {
		utils.ApplyKubernetesManifest(t, nsPath)
	}
}

func testMQTTProcessDataInteraction(t *testing.T, targetName, namespace, testDir string) {
	// Step 1: Create a simple Solution first
	solutionName := "test-mqtt-process-solution"
	solutionVersion := "test-mqtt-process-solution-v-version1"
	solutionYaml := fmt.Sprintf(`
apiVersion: solution.symphony/v1
kind: SolutionContainer
metadata:
  name: %s
  namespace: %s
spec:
---
apiVersion: solution.symphony/v1
kind: Solution
metadata:
  name: %s
  namespace: %s
spec:
  rootResource: %s
  components:
  - name: test-component
    type: script
    properties:
      script: |
        echo "MQTT Process test component deployed successfully"
        echo "Target: %s"
        echo "Namespace: %s"
`, solutionName, namespace, solutionVersion, namespace, solutionName, targetName, namespace)

	solutionPath := filepath.Join(testDir, "solution.yaml")
	err := utils.CreateYAMLFile(t, solutionPath, solutionYaml)
	require.NoError(t, err)

	// Apply the solution
	t.Logf("Creating Solution %s...", solutionName)
	err = utils.ApplyKubernetesManifest(t, solutionPath)
	require.NoError(t, err)

	// Step 2: Create an Instance that references the Solution and Target
	instanceName := "test-mqtt-process-instance"
	instanceYaml := fmt.Sprintf(`
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: %s
  namespace: %s
spec:
  displayName: %s
  solution: %s:version1
  target:
    name: %s
  scope: %s-scope
`, instanceName, namespace, instanceName, solutionName, targetName, namespace)

	instancePath := filepath.Join(testDir, "instance.yaml")
	err = utils.CreateYAMLFile(t, instancePath, instanceYaml)
	require.NoError(t, err)

	// Apply the instance
	t.Logf("Creating Instance %s that references Solution %s and Target %s...", instanceName, solutionName, targetName)
	err = utils.ApplyKubernetesManifest(t, instancePath)
	require.NoError(t, err)

	// Wait for Instance deployment to complete or reach a stable state
	t.Logf("Waiting for Instance %s to complete deployment...", instanceName)
	utils.WaitForInstanceReady(t, instanceName, namespace, 5*time.Minute)

	t.Cleanup(func() {
		// Delete in correct order: Instance -> Solution -> Target
		// Following the pattern from CleanUpSymphonyObjects function

		// First delete Instance and ensure it's completely removed
		t.Logf("Deleting Instance first...")
		err := utils.DeleteKubernetesResource(t, "instances.solution.symphony", instanceName, namespace, 2*time.Minute)
		if err != nil {
			t.Logf("Warning: Failed to delete instance: %v", err)
		} else {
			// Wait for Instance to be completely deleted before proceeding
			utils.WaitForResourceDeleted(t, "instance", instanceName, namespace, 1*time.Minute)
		}

		// Then delete Solution and ensure it's completely removed
		t.Logf("Deleting Solution...")
		err = utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
		if err != nil {
			t.Logf("Warning: Failed to delete solution: %v", err)
		} else {
			// Wait for Solution to be completely deleted before proceeding
			utils.WaitForResourceDeleted(t, "solution", solutionVersion, namespace, 1*time.Minute)
		}

		// Finally delete Target
		t.Logf("Deleting Target...")
		err = utils.DeleteKubernetesResource(t, "targets.fabric.symphony", targetName, namespace, 2*time.Minute)
		if err != nil {
			t.Logf("Warning: Failed to delete target: %v", err)
		}

		t.Logf("Cleanup completed")
	})

	// Give a short additional wait to ensure stability
	t.Logf("Instance deployment phase completed, test continuing...")
	time.Sleep(2 * time.Second)

	// Verify instance status
	// In a real test, you would check that:
	// 1. The instance was processed by Symphony
	// 2. The remote agent received deployment instructions
	// 3. The agent successfully executed the deployment
	// 4. Status was reported back to Symphony

	t.Logf("MQTT Process data interaction test completed - Solution and Instance created successfully")
}
