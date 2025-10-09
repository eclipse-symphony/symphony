package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

// Package-level variable for test directory
var scenario3TestDir string

func TestScenario3SingleTargetMultiInstance(t *testing.T) {
	// Test configuration - use relative path from test directory
	projectRoot := utils.GetProjectRoot(t) // Get project root dynamically
	namespace := "default"

	// Setup test environment
	scenario3TestDir = utils.SetupTestDirectory(t)
	t.Logf("Running Scenario 3 single target multi-instance test in: %s", scenario3TestDir)

	// Step 1: Start fresh minikube cluster
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Generate test certificates (with MyRootCA subject)
	certs := utils.GenerateTestCertificates(t, scenario3TestDir)

	// Setup test namespace
	setupScenario3Namespace(t, namespace)

	var caSecretName, clientSecretName string
	var configPath, topologyPath string
	var symphonyCAPath, baseURL string

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
		symphonyCAPath = utils.DownloadSymphonyCA(t, scenario3TestDir)
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
		configPath = utils.CreateHTTPConfig(t, scenario3TestDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, scenario3TestDir)
	})

	config := utils.TestConfig{
		ProjectRoot:    projectRoot,
		ConfigPath:     configPath,
		ClientCertPath: certs.ClientCert,
		ClientKeyPath:  certs.ClientKey,
		CACertPath:     symphonyCAPath,
		Namespace:      namespace,
		TopologyPath:   topologyPath,
		Protocol:       "http",
		BaseURL:        baseURL,
	}

	// Test single target with multiple instances
	t.Run("SingleTarget_MultiInstance", func(t *testing.T) {
		testSingleTargetMultiInstance(t, &config)
	})

	// Cleanup
	t.Cleanup(func() {
		// Clean up Symphony and other resources
		utils.CleanupSymphony(t, "remote-agent-scenario3-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("Scenario 3: Single target multi-instance test completed successfully")
}

func testSingleTargetMultiInstance(t *testing.T, config *utils.TestConfig) {
	targetName := "test-single-target"

	// Step 1: Create target and bootstrap remote agent
	t.Logf("=== Creating target and bootstrapping remote agent ===")

	err := createSingleTarget(t, config, targetName)
	require.NoError(t, err, "Failed to create target")

	// Bootstrap remote agent
	err = bootstrapSingleTargetAgent(t, config, targetName)
	require.NoError(t, err, "Failed to bootstrap remote agent")
	t.Logf("✓ Remote agent bootstrapped for target %s", targetName)

	// Wait for target to be ready
	utils.WaitForTargetReady(t, targetName, config.Namespace, 3*time.Minute)
	t.Logf("✓ Target %s is ready", targetName)

	// Step 2: Create 3 solutions in parallel
	solutionConfigs := []struct {
		name     string
		provider string
	}{
		{"single-target-script-solution-1", "script"},
		{"single-target-helm-solution-2", "script"},
		{"single-target-script-solution-3", "script"},
	}

	t.Logf("=== Creating 3 solutions in parallel ===")

	var solutionWg sync.WaitGroup
	solutionErrors := make(chan error, len(solutionConfigs))

	for _, solutionConfig := range solutionConfigs {
		solutionWg.Add(1)
		go func(solConfig struct{ name, provider string }) {
			defer solutionWg.Done()
			if err := createSingleTargetSolution(t, config, solConfig.name, solConfig.provider); err != nil {
				solutionErrors <- fmt.Errorf("failed to create solution %s: %v", solConfig.name, err)
			}
		}(solutionConfig)
	}

	solutionWg.Wait()
	close(solutionErrors)

	// Check for solution creation errors
	for err := range solutionErrors {
		require.NoError(t, err)
	}

	// Wait for all solutions to be ready
	for _, solutionConfig := range solutionConfigs {
		t.Logf("✓ Solution %s (%s provider) created successfully", solutionConfig.name, solutionConfig.provider)
	}

	// Step 3: Create 3 instances in parallel (all targeting the same target)
	instanceConfigs := []struct {
		instanceName string
		solutionName string
		provider     string
	}{
		{"single-target-instance-1", "single-target-script-solution-1", "script"},
		{"single-target-instance-2", "single-target-helm-solution-2", "script"},
		{"single-target-instance-3", "single-target-script-solution-3", "script"},
	}

	t.Logf("=== Creating 3 instances in parallel (all targeting same target) ===")

	var instanceWg sync.WaitGroup
	instanceErrors := make(chan error, len(instanceConfigs))

	for _, instanceConfig := range instanceConfigs {
		instanceWg.Add(1)
		go func(instConfig struct{ instanceName, solutionName, provider string }) {
			defer instanceWg.Done()
			if err := createSingleTargetInstance(t, config, instConfig.instanceName, instConfig.solutionName, targetName); err != nil {
				instanceErrors <- fmt.Errorf("failed to create instance %s: %v", instConfig.instanceName, err)
			}
		}(instanceConfig)
	}

	instanceWg.Wait()
	close(instanceErrors)

	// Check for instance creation errors
	for err := range instanceErrors {
		require.NoError(t, err)
	}

	// Wait for all instances to be ready and verify deployments
	for _, instanceConfig := range instanceConfigs {
		utils.WaitForInstanceReady(t, instanceConfig.instanceName, config.Namespace, 5*time.Minute)
		verifySingleTargetDeployment(t, instanceConfig.provider, instanceConfig.instanceName)
		t.Logf("✓ Instance %s (%s provider) is ready and deployed successfully on target %s",
			instanceConfig.instanceName, instanceConfig.provider, targetName)
	}

	// Step 4: Delete instances in parallel
	t.Logf("=== Deleting 3 instances in parallel ===")

	var deleteInstanceWg sync.WaitGroup
	deleteInstanceErrors := make(chan error, len(instanceConfigs))

	for _, instanceConfig := range instanceConfigs {
		deleteInstanceWg.Add(1)
		go func(instConfig struct{ instanceName, solutionName, provider string }) {
			defer deleteInstanceWg.Done()
			if err := deleteSingleTargetInstance(t, config, instConfig.instanceName); err != nil {
				deleteInstanceErrors <- fmt.Errorf("failed to delete instance %s: %v", instConfig.instanceName, err)
			}
		}(instanceConfig)
	}

	deleteInstanceWg.Wait()
	close(deleteInstanceErrors)

	// Check for instance deletion errors
	for err := range deleteInstanceErrors {
		require.NoError(t, err)
	}

	// Wait for all instances to be deleted
	for _, instanceConfig := range instanceConfigs {
		utils.WaitForResourceDeleted(t, "instance", instanceConfig.instanceName, config.Namespace, 2*time.Minute)
		t.Logf("✓ Instance %s deleted successfully", instanceConfig.instanceName)
	}

	// Step 5: Delete solutions in parallel
	t.Logf("=== Deleting 3 solutions in parallel ===")

	var deleteSolutionWg sync.WaitGroup
	deleteSolutionErrors := make(chan error, len(solutionConfigs))

	for _, solutionConfig := range solutionConfigs {
		deleteSolutionWg.Add(1)
		go func(solConfig struct{ name, provider string }) {
			defer deleteSolutionWg.Done()
			if err := deleteSingleTargetSolution(t, config, solConfig.name); err != nil {
				deleteSolutionErrors <- fmt.Errorf("failed to delete solution %s: %v", solConfig.name, err)
			}
		}(solutionConfig)
	}

	deleteSolutionWg.Wait()
	close(deleteSolutionErrors)

	// Check for solution deletion errors
	for err := range deleteSolutionErrors {
		require.NoError(t, err)
	}

	// Wait for all solutions to be deleted
	for _, solutionConfig := range solutionConfigs {
		utils.WaitForResourceDeleted(t, "solution", solutionConfig.name, config.Namespace, 2*time.Minute)
		t.Logf("✓ Solution %s deleted successfully", solutionConfig.name)
	}

	// Step 6: Delete target
	t.Logf("=== Deleting target ===")

	err = deleteSingleTarget(t, config, targetName)
	require.NoError(t, err, "Failed to delete target")

	// Wait for target to be deleted
	utils.WaitForResourceDeleted(t, "target", targetName, config.Namespace, 2*time.Minute)
	t.Logf("✓ Target %s deleted successfully", targetName)

	t.Logf("=== Scenario 3: Single target with multiple instances completed successfully ===")
}

// Helper functions for single target multi-instance operations

func createSingleTarget(t *testing.T, config *utils.TestConfig, targetName string) error {
	// Use the standard CreateTargetYAML function from utils
	targetPath := utils.CreateTargetYAML(t, scenario3TestDir, targetName, config.Namespace)
	return utils.ApplyKubernetesManifest(t, targetPath)
}

func bootstrapSingleTargetAgent(t *testing.T, config *utils.TestConfig, targetName string) error {
	// Start remote agent using direct process (no systemd service) without automatic cleanup
	// Create a config specific to this target
	targetConfig := *config
	targetConfig.TargetName = targetName

	processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, targetConfig)
	if processCmd == nil {
		return fmt.Errorf("failed to start remote agent process for target %s", targetName)
	}

	// Wait for the process to be healthy
	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)

	// Add cleanup for this process (will be handled when test completes)
	t.Cleanup(func() {
		t.Logf("Cleaning up remote agent process for target %s...", targetName)
		utils.CleanupRemoteAgentProcess(t, processCmd)
	})

	return nil
}

func createSingleTargetSolution(t *testing.T, config *utils.TestConfig, solutionName, provider string) error {
	var solutionYaml string
	solutionVersion := fmt.Sprintf("%s-v-version1", solutionName)

	switch provider {
	case "script":
		// Create a temporary script file for the script provider
		scriptContent := fmt.Sprintf(`#!/bin/bash
echo "=== Script Provider Single Target Test ==="
echo "Solution: %s"
echo "Timestamp: $(date)"
echo "Creating marker file..."
echo "Single target multi-instance test successful at $(date)" > /tmp/%s-test.log
echo "=== Script Provider Test Completed ==="
exit 0
`, solutionName, solutionName)

		// Write script to a temporary file
		scriptPath := filepath.Join(scenario3TestDir, fmt.Sprintf("%s-script.sh", solutionName))
		err := utils.CreateYAMLFile(t, scriptPath, scriptContent) // CreateYAMLFile can handle any text content
		if err != nil {
			return err
		}

		// Make script executable
		err = os.Chmod(scriptPath, 0755)
		if err != nil {
			return err
		}

		solutionYaml = fmt.Sprintf(`
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
  - name: %s-script-component
    type: script
    properties:
      path: %s
`, solutionName, config.Namespace, solutionVersion, config.Namespace, solutionName, solutionName, scriptPath)

	case "helm":
		solutionYaml = fmt.Sprintf(`
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
  - name: %s-helm-component
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
          test.symphony.com/scenario: "single-target-multi-instance"
          test.symphony.com/solution: "%s"
`, solutionName, config.Namespace, solutionVersion, config.Namespace, solutionName, solutionName, solutionName)

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	solutionPath := filepath.Join(scenario3TestDir, fmt.Sprintf("%s-solution.yaml", solutionName))
	if err := utils.CreateYAMLFile(t, solutionPath, solutionYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, solutionPath)
}

func createSingleTargetInstance(t *testing.T, config *utils.TestConfig, instanceName, solutionName, targetName string) error {
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
`, instanceName, config.Namespace, instanceName, solutionName, targetName, config.Namespace)

	instancePath := filepath.Join(scenario3TestDir, fmt.Sprintf("%s-instance.yaml", instanceName))
	if err := utils.CreateYAMLFile(t, instancePath, instanceYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, instancePath)
}

func deleteSingleTargetInstance(t *testing.T, config *utils.TestConfig, instanceName string) error {
	instancePath := filepath.Join(scenario3TestDir, fmt.Sprintf("%s-instance.yaml", instanceName))
	return utils.DeleteKubernetesManifest(t, instancePath)
}

func deleteSingleTargetSolution(t *testing.T, config *utils.TestConfig, solutionName string) error {
	solutionPath := filepath.Join(scenario3TestDir, fmt.Sprintf("%s-solution.yaml", solutionName))
	return utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
}

func deleteSingleTarget(t *testing.T, config *utils.TestConfig, targetName string) error {
	targetPath := filepath.Join(scenario3TestDir, "target.yaml")
	return utils.DeleteKubernetesManifest(t, targetPath)
}

func verifySingleTargetDeployment(t *testing.T, provider, instanceName string) {
	switch provider {
	case "script":
		t.Logf("Verifying script deployment for instance: %s", instanceName)
	// Script verification would check for marker files or logs
	// For now, we rely on the instance being ready
	case "helm":
		t.Logf("Verifying helm deployment for instance: %s", instanceName)
	// Helm verification would check for deployed charts and running pods
	// For now, we rely on the instance being ready
	default:
		t.Logf("Unknown provider type for verification: %s", provider)
	}
}

func setupScenario3Namespace(t *testing.T, namespace string) {
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

	nsPath := filepath.Join(scenario3TestDir, "namespace.yaml")
	err = utils.CreateYAMLFile(t, nsPath, nsYaml)
	if err == nil {
		utils.ApplyKubernetesManifest(t, nsPath)
	}
}
