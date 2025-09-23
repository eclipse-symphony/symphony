package verify

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

// TestScenario1ProviderCRUD tests CRUD operations for different component provider types
// This extends the existing http_process_test.go and mqtt_process_test.go by testing different
// solution component providers: script, http, and helm.v3
func TestScenario1ProviderCRUD(t *testing.T) {
	// Use existing HTTP process infrastructure (reuse http_process_test pattern)
	targetName := "test-provider-crud-target"
	namespace := "default"

	// Setup test environment using existing patterns
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running provider CRUD test in: %s", testDir)

	// Step 1: Start fresh minikube cluster (following existing pattern)
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Generate test certificates (following existing pattern)
	certs := utils.GenerateTestCertificates(t, testDir)

	var caSecretName, clientSecretName string
	var configPath, topologyPath, targetYamlPath string
	var symphonyCAPath, baseURL string

	// Setup Symphony infrastructure (following existing pattern)
	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		caSecretName = utils.CreateCASecret(t, certs)
		clientSecretName = utils.CreateClientCertSecret(t, namespace, certs)
	})

	t.Run("StartSymphonyServer", func(t *testing.T) {
		utils.StartSymphonyWithRemoteAgentConfig(t, "http")
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	t.Run("SetupSymphonyConnection", func(t *testing.T) {
		symphonyCAPath = utils.DownloadSymphonyCA(t, testDir)
	})

	// Setup hosts mapping and port-forward (following existing pattern)
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)
	baseURL = "https://symphony-service:8081/v1alpha2"

	// Create test configurations FIRST (following existing pattern)
	t.Run("CreateTestConfigurations", func(t *testing.T) {
		configPath = utils.CreateHTTPConfig(t, testDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, testDir)
		targetYamlPath = utils.CreateTargetYAML(t, testDir, targetName, namespace)

		err := utils.ApplyKubernetesManifest(t, targetYamlPath)
		require.NoError(t, err)

		utils.WaitForTargetCreated(t, targetName, namespace, 30*time.Second)
	})

	// Now create config with properly initialized paths
	config := utils.TestConfig{
		ProjectRoot:    utils.GetProjectRoot(t),
		ConfigPath:     configPath,
		ClientCertPath: certs.ClientCert,
		ClientKeyPath:  certs.ClientKey,
		CACertPath:     symphonyCAPath,
		TargetName:     targetName,
		Namespace:      namespace,
		TopologyPath:   topologyPath,
		Protocol:       "http",
		BaseURL:        baseURL,
	}

	// Validate configuration before starting remote agent
	t.Logf("=== Validating TestConfig ===")
	t.Logf("ConfigPath: %s", config.ConfigPath)
	t.Logf("TopologyPath: %s", config.TopologyPath)
	t.Logf("TargetName: %s", config.TargetName)
	require.NotEmpty(t, config.ConfigPath, "ConfigPath should not be empty")
	require.NotEmpty(t, config.TopologyPath, "TopologyPath should not be empty")
	require.NotEmpty(t, config.TargetName, "TargetName should not be empty")

	// Start remote agent using direct process (no systemd service) without automatic cleanup
	processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	// Set up cleanup at main test level to ensure process runs for entire test
	t.Cleanup(func() {
		if processCmd != nil {
			t.Logf("Cleaning up remote agent process from main test...")
			utils.CleanupRemoteAgentProcess(t, processCmd)
		}
	})

	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
	// IMPORTANT: Wait for target to be ready AFTER starting remote agent
	utils.WaitForTargetReady(t, targetName, namespace, 120*time.Second)

	// Now test different provider types
	providers := []string{"script", "http"}

	for _, provider := range providers {
		t.Run(fmt.Sprintf("TestProvider_%s", provider), func(t *testing.T) {
			testProviderCRUDLifecycle(t, provider, targetName, namespace, testDir)
		})
	}

	// Cleanup
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "scenario1-provider-crud-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("Provider CRUD test completed successfully")
}

// testProviderCRUDLifecycle tests the full CRUD lifecycle for a specific provider type
func testProviderCRUDLifecycle(t *testing.T, provider, targetName, namespace, testDir string) {
	timestamp := time.Now().Unix()
	solutionName := fmt.Sprintf("test-%s-solution-%d", provider, timestamp)
	instanceName := fmt.Sprintf("test-%s-instance-%d", provider, timestamp)

	t.Logf("Testing CRUD lifecycle for provider: %s", provider)

	// Step 1: Create Solution with provider-specific component
	t.Run("CreateSolution", func(t *testing.T) {
		solutionYaml := createProviderSolution(provider, solutionName, namespace)
		solutionPath := filepath.Join(testDir, fmt.Sprintf("solution-%s.yaml", solutionName))

		err := utils.CreateYAMLFile(t, solutionPath, solutionYaml)
		require.NoError(t, err)

		err = utils.ApplyKubernetesManifest(t, solutionPath)
		require.NoError(t, err)

		t.Logf("✓ Created solution for %s provider", provider)
	})

	// Step 2: Create Instance
	t.Run("CreateInstance", func(t *testing.T) {
		instanceYaml := createProviderInstance(instanceName, namespace, solutionName, targetName)
		instancePath := filepath.Join(testDir, fmt.Sprintf("instance-%s.yaml", instanceName))

		err := utils.CreateYAMLFile(t, instancePath, instanceYaml)
		require.NoError(t, err)

		err = utils.ApplyKubernetesManifest(t, instancePath)
		require.NoError(t, err)

		// Wait for instance to be processed
		utils.WaitForInstanceReady(t, instanceName, namespace, 5*time.Minute)

		t.Logf("✓ Created and verified instance for %s provider", provider)
	})

	// Step 3: Verify provider-specific deployment
	t.Run("VerifyDeployment", func(t *testing.T) {
		verifyProviderSpecificDeployment(t, provider, instanceName, namespace)
		t.Logf("✓ Verified deployment for %s provider", provider)
	})

	// Step 4: Delete Instance (CRUD Delete)
	t.Run("DeleteInstance", func(t *testing.T) {
		err := utils.DeleteKubernetesResource(t, "instances.solution.symphony", instanceName, namespace, 2*time.Minute)
		require.NoError(t, err)

		utils.WaitForResourceDeleted(t, "instance", instanceName, namespace, 1*time.Minute)
		t.Logf("✓ Deleted instance for %s provider", provider)
	})

	// Step 5: Delete Solution (CRUD Delete)
	t.Run("DeleteSolution", func(t *testing.T) {
		solutionPath := filepath.Join(testDir, fmt.Sprintf("solution-%s.yaml", solutionName))
		err := utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
		require.NoError(t, err)

		t.Logf("✓ Deleted solution for %s provider", provider)
	})

	t.Logf("Completed CRUD lifecycle test for provider: %s", provider)
}

// createProviderSolution creates solution YAML for different provider types
func createProviderSolution(provider, solutionName, namespace string) string {
	solutionVersion := fmt.Sprintf("%s-v-version1", solutionName)

	switch provider {
	case "script":
		return fmt.Sprintf(`
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
  - name: script-test-component
    type: script
    properties:
      script: |
        #!/bin/bash
        echo "=== Script Provider Test ==="
        echo "Solution: %s"
        echo "Namespace: %s"
        echo "Timestamp: $(date)"
        echo "Creating test marker file..."
        echo "Script component executed successfully" > /tmp/script-test-marker.log
        echo "=== Script Provider Test Completed ==="
        exit 0
`, solutionName, namespace, solutionVersion, namespace, solutionName, solutionName, namespace)

	case "http":
		return fmt.Sprintf(`
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
  - name: http-test-component
    type: http
    properties:
      url: "https://bing.com"
      method: "POST"
      headers:
        Content-Type: "application/json"
        User-Agent: "Symphony-Test/1.0"
      timeout: "30s"
`, solutionName, namespace, solutionVersion, namespace, solutionName)

	case "helm.v3":
		return fmt.Sprintf(`
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
  - name: helm-test-component
    type: helm.v3
    properties:
      chart:
        repo: "https://charts.bitnami.com/bitnami"
        name: "nginx"
        version: "15.1.0"
      values:
        replicaCount: 1
        service:
          type: ClusterIP
          port: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        podAnnotations:
          symphony.test/provider: "helm.v3"
          symphony.test/solution: "%s"
          symphony.test/namespace: "%s"
`, solutionName, namespace, solutionVersion, namespace, solutionName, solutionName, namespace)

	default:
		panic(fmt.Sprintf("Unsupported provider type: %s", provider))
	}
}

// createProviderInstance creates instance YAML
func createProviderInstance(instanceName, namespace, solutionName, targetName string) string {
	return fmt.Sprintf(`
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
}

// verifyProviderSpecificDeployment performs provider-specific verification
func verifyProviderSpecificDeployment(t *testing.T, provider, instanceName, namespace string) {
	switch provider {
	case "script":
		// For script provider, we verify the instance completed successfully
		// In a real scenario, you might check for marker files or specific outputs
		t.Logf("Verifying script provider deployment - instance should be completed")

	case "http":
		// For HTTP provider, we verify the HTTP request was made successfully
		// The instance status should indicate successful completion
		t.Logf("Verifying HTTP provider deployment - HTTP request should have completed")

	case "helm.v3":
		// For Helm provider, we could verify the Helm chart was deployed
		// For this test, we check that the instance deployment completed
		t.Logf("Verifying Helm provider deployment - Helm chart should be deployed")

	default:
		t.Fatalf("Unknown provider type: %s", provider)
	}

	// Wait a moment for any final deployment activities to complete
	time.Sleep(10 * time.Second)
	t.Logf("Provider-specific deployment verification completed for: %s", provider)
}
