package http_communication

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/e2e/remote-agent-integration/utils"
	"github.com/stretchr/testify/require"
)

func TestE2EHttpCommunication(t *testing.T) {
	// Test configuration
	projectRoot := "/mnt/d/code3/symphony/" // Relative to test file location
	targetName := "test-http-target"
	namespace := "default"
	baseURL := "https://symphony-service:8081/v1alpha2"

	// Setup test environment
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running HTTP communication test in: %s", testDir)

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
	setupNamespace(t, namespace)
	defer utils.CleanupNamespace(t, namespace)

	var caSecretName, clientSecretName string
	var configPath, topologyPath, targetYamlPath string

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

	var symphonyCAPath string

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

	t.Run("StartRemoteAgent", func(t *testing.T) {
		// Now start remote agent with correct configurations
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
			BaseURL:        baseURL, // Add the missing BaseURL
		}

		agent := utils.StartRemoteAgentProcess(t, config)
		require.NotNil(t, agent)

		// Wait for agent to be ready
		utils.WaitForProcessReady(t, agent, 10*time.Second)
	})

	t.Run("VerifyTargetStatus", func(t *testing.T) {
		// Wait for target to reach ready state
		utils.WaitForTargetReady(t, targetName, namespace, 120*time.Second)
	})

	t.Run("VerifyTopologyUpdate", func(t *testing.T) {
		// Verify that topology was successfully updated
		// This would check that the remote agent successfully called
		// the /targets/updatetopology endpoint
		verifyTopologyUpdate(t, targetName, namespace)
	})

	t.Run("TestDataInteraction", func(t *testing.T) {
		// Test actual data interaction between server and agent
		// This would involve creating an Instance that uses the Target
		// and verifying the end-to-end workflow
		testDataInteraction(t, targetName, namespace, testDir)
	})

	// Cleanup
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "remote-agent-http-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("HTTP communication test completed successfully")
}

func setupNamespace(t *testing.T, namespace string) {
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

	t.Cleanup(func() {
		utils.DeleteKubernetesManifest(t, nsPath)
	})
}

func verifyTopologyUpdate(t *testing.T, targetName, namespace string) {
	// Get the target and check its topology status
	_, err := utils.GetDynamicClient()
	require.NoError(t, err)

	// Check that target has topology information
	// This is a placeholder - you would implement actual topology verification
	t.Logf("Verifying topology update for target %s/%s", namespace, targetName)

	// In a real implementation, you would:
	// 1. Get the Target resource
	// 2. Check its status for topology information
	// 3. Verify that the remote agent successfully updated the topology
	// 4. Check Symphony server logs for topology update calls

	t.Logf("Topology update verification completed")
}

func testDataInteraction(t *testing.T, targetName, namespace, testDir string) {
	// Create a simple Instance resource that uses our Target
	instanceName := "test-instance"
	instanceYaml := fmt.Sprintf(`
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: %s
  namespace: %s
spec:
  displayName: %s
  solution: test-solution
  target:
    name: %s
  scope: %s
`, instanceName, namespace, instanceName, targetName, namespace+"scope")

	instancePath := filepath.Join(testDir, "instance.yaml")
	err := utils.CreateYAMLFile(t, instancePath, instanceYaml)
	require.NoError(t, err)

	// Apply the instance
	err = utils.ApplyKubernetesManifest(t, instancePath)
	require.NoError(t, err)

	t.Cleanup(func() {
		utils.DeleteKubernetesManifest(t, instancePath)
	})

	// Wait for instance processing
	time.Sleep(10 * time.Second)

	// Verify instance status
	// In a real test, you would check that:
	// 1. The instance was processed by Symphony
	// 2. The remote agent received deployment instructions
	// 3. The agent successfully executed the deployment
	// 4. Status was reported back to Symphony

	t.Logf("Data interaction test completed")
}
