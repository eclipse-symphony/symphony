package verify

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent/utils"
	"github.com/stretchr/testify/require"
)

func TestE2EHttpCommunicationWithProcess(t *testing.T) {
	// Test configuration - use relative path from test directory
	projectRoot := utils.GetProjectRoot(t) // Get project root dynamically
	targetName := "test-http-process-target"
	namespace := "default"

	// Setup test environment
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running HTTP process test in: %s", testDir)

	// Step 1: Start fresh minikube cluster
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Generate test certificates (with MyRootCA subject)
	certs := utils.GenerateTestCertificates(t, testDir)

	// Setup test namespace
	setupProcessNamespace(t, namespace)

	var caSecretName, clientSecretName string
	var configPath, topologyPath, targetYamlPath string
	var symphonyCAPath, baseURL string
	var processCmd *exec.Cmd

	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		// Create CA secret in cert-manager namespace
		caSecretName = utils.CreateCASecret(t, certs)

		// Create client cert secret in test namespace
		clientSecretName = utils.CreateClientCertSecret(t, namespace, certs)
	})

	t.Run("StartSymphonyServer", func(t *testing.T) {
		utils.StartSymphonyWithRemoteAgentConfig(t, "http")

		// Wait for Symphony server certificate to be created
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	t.Run("SetupSymphonyConnection", func(t *testing.T) {
		// Download Symphony server CA certificate
		symphonyCAPath = utils.DownloadSymphonyCA(t, testDir)

		t.Logf("Symphony server CA certificate downloaded")
	})

	// Setup hosts mapping and port-forward at main test level so they persist
	// across all sub-tests until the main test completes
	t.Logf("Setting up hosts mapping and port-forward...")
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)
	baseURL = "https://symphony-service:8081/v1alpha2"
	t.Logf("Symphony server accessible at: %s", baseURL)

	// Create test configurations AFTER Symphony is running
	t.Run("CreateTestConfigurations", func(t *testing.T) {
		configPath = utils.CreateHTTPConfig(t, testDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, testDir)
		targetYamlPath = utils.CreateTargetYAML(t, testDir, targetName, namespace)

		// Apply Target YAML to create the target resource
		err := utils.ApplyKubernetesManifest(t, targetYamlPath)
		require.NoError(t, err)

		// Wait for target to be created
		utils.WaitForTargetCreated(t, targetName, namespace, 30*time.Second)
	})

	// Start the remote agent process at main test level so it persists across subtests
	t.Logf("Starting remote agent process...")
	config := utils.TestConfig{
		ProjectRoot:    projectRoot,
		ConfigPath:     configPath,
		ClientCertPath: certs.ClientCert,
		ClientKeyPath:  certs.ClientKey,
		CACertPath:     symphonyCAPath, // Use Symphony server CA for TLS trust
		TargetName:     targetName,
		Namespace:      namespace,
		TopologyPath:   topologyPath,
		Protocol:       "http",
		BaseURL:        baseURL,
	}

	// Start remote agent using direct process (no systemd service) without automatic cleanup
	processCmd = utils.StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	// Set up cleanup at main test level to ensure process runs for entire test
	t.Cleanup(func() {
		if processCmd != nil {
			t.Logf("Cleaning up remote agent process from main test...")
			utils.CleanupRemoteAgentProcess(t, processCmd)
		}
	})

	// Wait for process to be ready and healthy
	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
	t.Logf("Remote agent process started successfully and will persist across all subtests")

	t.Run("VerifyProcessStarted", func(t *testing.T) {
		// Just verify the process is running
		require.NotNil(t, processCmd)
		require.NotNil(t, processCmd.Process)
		t.Logf("Remote agent process verified running with PID: %d", processCmd.Process.Pid)
	})

	t.Run("VerifyTargetStatus", func(t *testing.T) {
		// Wait for target to reach ready state
		utils.WaitForTargetReady(t, targetName, namespace, 120*time.Second)
	})

	t.Run("VerifyTopologyUpdate", func(t *testing.T) {
		// Verify that topology was successfully updated
		// This would check that the remote agent successfully called
		// the /targets/updatetopology endpoint
		utils.VerifyTargetTopologyUpdate(t, targetName, namespace, "HTTP process")
	})

	t.Run("TestDataInteraction", func(t *testing.T) {
		// Test actual data interaction between server and agent
		// This would involve creating an Instance that uses the Target
		// and verifying the end-to-end workflow
		testProcessDataInteraction(t, targetName, namespace, testDir)
	})

	// Cleanup
	t.Cleanup(func() {
		// Clean up Symphony and other resources
		utils.CleanupSymphony(t, "remote-agent-http-process-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("HTTP communication test with direct process completed successfully")
}

func setupProcessNamespace(t *testing.T, namespace string) {
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

func testProcessDataInteraction(t *testing.T, targetName, namespace, testDir string) {
	// Step 1: Create a simple Solution first
	solutionName := "test-process-solution"
	solutionVersion := "test-process-solution-v-version1"
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
        echo "Process test component deployed successfully"
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
	instanceName := "test-process-instance"
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

	t.Logf("Process data interaction test completed - Solution and Instance created successfully")
}
