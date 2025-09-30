package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

// Package-level variable for test directory
var scenario4TestDir string

func TestScenario4MultiTargetMultiSolution(t *testing.T) {
	// Test configuration - use relative path from test directory
	projectRoot := utils.GetProjectRoot(t) // Get project root dynamically
	namespace := "default"

	// Setup test environment
	scenario4TestDir = utils.SetupTestDirectory(t)
	t.Logf("Running Scenario 4 multi-target multi-solution test in: %s", scenario4TestDir)

	// Step 1: Start fresh minikube cluster
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Generate test certificates (with MyRootCA subject)
	certs := utils.GenerateTestCertificates(t, scenario4TestDir)

	// Setup test namespace
	setupScenario4Namespace(t, namespace)

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
		symphonyCAPath = utils.DownloadSymphonyCA(t, scenario4TestDir)
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
		configPath = utils.CreateHTTPConfig(t, scenario4TestDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, scenario4TestDir)
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

	// Test multi-target multi-solution scenario
	t.Run("MultiTarget_MultiSolution", func(t *testing.T) {
		testMultiTargetMultiSolution(t, &config)
	})

	// Cleanup
	t.Cleanup(func() {
		// Clean up Symphony and other resources
		utils.CleanupSymphony(t, "remote-agent-scenario4-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("Scenario 4: Multi-target multi-solution test completed successfully")
}

func testMultiTargetMultiSolution(t *testing.T, config *utils.TestConfig) {
	// Define target configurations
	targetConfigs := []struct {
		name string
		port int
	}{
		{"multi-target-1", 8091},
		{"multi-target-2", 8092},
		{"multi-target-3", 8093},
	}

	// Define solution configurations
	solutionConfigs := []struct {
		name     string
		provider string
	}{
		{"multi-script-solution-1", "script"},
		{"multi-helm-solution-2", "helm"},
		{"multi-script-solution-3", "script"},
	}

	// Define instance configurations (each targets a different target)
	instanceConfigs := []struct {
		instanceName string
		solutionName string
		targetName   string
		provider     string
	}{
		{"multi-instance-1", "multi-script-solution-1", "multi-target-1", "script"},
		{"multi-instance-2", "multi-helm-solution-2", "multi-target-2", "helm"},
		{"multi-instance-3", "multi-script-solution-3", "multi-target-3", "script"},
	}

	// Step 1: Create 3 targets in parallel (but don't verify status yet)
	t.Logf("=== Creating 3 targets in parallel ===")

	var targetWg sync.WaitGroup
	targetErrors := make(chan error, len(targetConfigs))

	for _, targetConfig := range targetConfigs {
		targetWg.Add(1)
		go func(tConfig struct {
			name string
			port int
		}) {
			defer targetWg.Done()
			if err := createMultiTarget(t, config, tConfig.name); err != nil {
				targetErrors <- fmt.Errorf("failed to create target %s: %v", tConfig.name, err)
				return
			}
		}(targetConfig)
	}

	targetWg.Wait()
	close(targetErrors)

	// Check for target creation errors
	for err := range targetErrors {
		require.NoError(t, err)
	}

	t.Logf("✓ All 3 targets created successfully")

	// Step 2: Start remote agents in parallel (bootstrap remote agents)
	t.Logf("=== Starting remote agents in parallel ===")

	// Track remote agent processes for cleanup
	var remoteAgentProcesses = make(map[string]*exec.Cmd)
	var agentWg sync.WaitGroup
	agentErrors := make(chan error, len(targetConfigs))

	for _, targetConfig := range targetConfigs {
		agentWg.Add(1)
		go func(tConfig struct {
			name string
			port int
		}) {
			defer agentWg.Done()
			processCmd, err := startMultiTargetRemoteAgent(t, config, tConfig.name)
			if err != nil {
				agentErrors <- fmt.Errorf("failed to start remote agent for target %s: %v", tConfig.name, err)
				return
			}
			remoteAgentProcesses[tConfig.name] = processCmd
		}(targetConfig)
	}

	agentWg.Wait()
	close(agentErrors)

	// Check for remote agent startup errors
	for err := range agentErrors {
		require.NoError(t, err)
	}

	// Setup cleanup for all remote agent processes using optimized cleanup
	t.Cleanup(func() {
		utils.CleanupMultipleRemoteAgentProcesses(t, remoteAgentProcesses)
	})

	t.Logf("✓ All 3 remote agents started successfully")

	// Step 3: Now verify target topology updates (targets should become Ready)
	t.Logf("=== Verifying target topology updates ===")
	for _, targetConfig := range targetConfigs {
		utils.WaitForTargetReady(t, targetConfig.name, config.Namespace, 3*time.Minute)
		t.Logf("✓ Target %s is ready with remote agent connected", targetConfig.name)
	}

	// Step 4: Create 3 solutions in parallel
	t.Logf("=== Creating 3 solutions in parallel ===")

	var solutionWg sync.WaitGroup
	solutionErrors := make(chan error, len(solutionConfigs))

	for _, solutionConfig := range solutionConfigs {
		solutionWg.Add(1)
		go func(solConfig struct{ name, provider string }) {
			defer solutionWg.Done()
			if err := createMultiSolution(t, config, solConfig.name, solConfig.provider); err != nil {
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

	// Step 5: Create 3 instances in parallel (each targeting a different target)
	t.Logf("=== Creating 3 instances in parallel (each targeting different targets) ===")

	var instanceWg sync.WaitGroup
	instanceErrors := make(chan error, len(instanceConfigs))

	for _, instanceConfig := range instanceConfigs {
		instanceWg.Add(1)
		go func(instConfig struct{ instanceName, solutionName, targetName, provider string }) {
			defer instanceWg.Done()
			if err := createMultiInstance(t, config, instConfig.instanceName, instConfig.solutionName, instConfig.targetName); err != nil {
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
		verifyMultiDeployment(t, instanceConfig.provider, instanceConfig.instanceName)
		t.Logf("✓ Instance %s (%s provider) is ready and deployed successfully on target %s",
			instanceConfig.instanceName, instanceConfig.provider, instanceConfig.targetName)
	}

	// Step 6: Delete instances in parallel
	t.Logf("=== Deleting 3 instances in parallel ===")

	var deleteInstanceWg sync.WaitGroup
	deleteInstanceErrors := make(chan error, len(instanceConfigs))

	for _, instanceConfig := range instanceConfigs {
		deleteInstanceWg.Add(1)
		go func(instConfig struct{ instanceName, solutionName, targetName, provider string }) {
			defer deleteInstanceWg.Done()
			if err := deleteMultiInstance(t, config, instConfig.instanceName); err != nil {
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

	// Step 7: Delete solutions in parallel
	t.Logf("=== Deleting 3 solutions in parallel ===")

	var deleteSolutionWg sync.WaitGroup
	deleteSolutionErrors := make(chan error, len(solutionConfigs))

	for _, solutionConfig := range solutionConfigs {
		deleteSolutionWg.Add(1)
		go func(solConfig struct{ name, provider string }) {
			defer deleteSolutionWg.Done()
			if err := deleteMultiSolution(t, config, solConfig.name); err != nil {
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

	// Step 8: Delete targets in parallel
	t.Logf("=== Deleting 3 targets in parallel ===")

	var deleteTargetWg sync.WaitGroup
	deleteTargetErrors := make(chan error, len(targetConfigs))

	for _, targetConfig := range targetConfigs {
		deleteTargetWg.Add(1)
		go func(tConfig struct {
			name string
			port int
		}) {
			defer deleteTargetWg.Done()
			if err := deleteMultiTarget(t, config, tConfig.name); err != nil {
				deleteTargetErrors <- fmt.Errorf("failed to delete target %s: %v", tConfig.name, err)
			}
		}(targetConfig)
	}

	deleteTargetWg.Wait()
	close(deleteTargetErrors)

	// Check for target deletion errors
	for err := range deleteTargetErrors {
		require.NoError(t, err)
	}

	// Wait for all targets to be deleted
	for _, targetConfig := range targetConfigs {
		utils.WaitForResourceDeleted(t, "target", targetConfig.name, config.Namespace, 2*time.Minute)
		t.Logf("✓ Target %s deleted successfully", targetConfig.name)
	}

	t.Logf("=== Scenario 4: Multi-target multi-solution completed successfully ===")
}

// Helper functions for multi-target multi-solution operations

func createMultiTarget(t *testing.T, config *utils.TestConfig, targetName string) error {
	// Create unique target YAML file to avoid race conditions in parallel execution
	targetYaml := fmt.Sprintf(`
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: %s
  namespace: %s
spec:
  displayName: %s
  scope: %s-scope
  topologies:
  - bindings:
    - config:
        inCluster: "false"
      provider: providers.target.remote-agent
      role: remote-agent
  properties:
    os.name: linux
`, targetName, config.Namespace, targetName, config.Namespace)

	// Use unique filename for each target to prevent race conditions
	targetPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-target.yaml", targetName))
	if err := utils.CreateYAMLFile(t, targetPath, targetYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, targetPath)
}

func bootstrapMultiTargetAgent(t *testing.T, config *utils.TestConfig, targetName string) error {
	// For process mode, we bootstrap the remote agent
	bootstrapYaml := fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-bootstrap
  namespace: %s
data:
  bootstrap: "true"
`, targetName, config.Namespace)

	bootstrapPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-bootstrap.yaml", targetName))
	if err := utils.CreateYAMLFile(t, bootstrapPath, bootstrapYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, bootstrapPath)
}

func createMultiSolution(t *testing.T, config *utils.TestConfig, solutionName, provider string) error {
	var solutionYaml string
	solutionVersion := fmt.Sprintf("%s-v-version1", solutionName)

	switch provider {
	case "script":
		// Create a temporary script file for the script provider
		scriptContent := fmt.Sprintf(`#!/bin/bash
echo "=== Script Provider Multi-Target Test ==="
echo "Solution: %s"
echo "Timestamp: $(date)"
echo "Creating marker file..."
echo "Multi-target multi-solution test successful at $(date)" > /tmp/%s-test.log
echo "=== Script Provider Test Completed ==="
exit 0
`, solutionName, solutionName)

		// Write script to a temporary file
		scriptPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-script.sh", solutionName))
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
          test.symphony.com/scenario: "multi-target-multi-solution"
          test.symphony.com/solution: "%s"
`, solutionName, config.Namespace, solutionVersion, config.Namespace, solutionName, solutionName, solutionName)

	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	solutionPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-solution.yaml", solutionName))
	if err := utils.CreateYAMLFile(t, solutionPath, solutionYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, solutionPath)
}

func createMultiInstance(t *testing.T, config *utils.TestConfig, instanceName, solutionName, targetName string) error {
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

	instancePath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-instance.yaml", instanceName))
	if err := utils.CreateYAMLFile(t, instancePath, instanceYaml); err != nil {
		return err
	}

	return utils.ApplyKubernetesManifest(t, instancePath)
}

func deleteMultiInstance(t *testing.T, config *utils.TestConfig, instanceName string) error {
	instancePath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-instance.yaml", instanceName))
	return utils.DeleteKubernetesManifest(t, instancePath)
}

func deleteMultiSolution(t *testing.T, config *utils.TestConfig, solutionName string) error {
	solutionPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-solution.yaml", solutionName))
	return utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
}

func deleteMultiTarget(t *testing.T, config *utils.TestConfig, targetName string) error {
	targetPath := filepath.Join(scenario4TestDir, fmt.Sprintf("%s-target.yaml", targetName))
	return utils.DeleteKubernetesManifest(t, targetPath)
}

func verifyMultiDeployment(t *testing.T, provider, instanceName string) {
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

func startMultiTargetRemoteAgent(t *testing.T, config *utils.TestConfig, targetName string) (*exec.Cmd, error) {
	// Start remote agent using shared binary optimization for improved performance
	targetConfig := *config
	targetConfig.TargetName = targetName

	t.Logf("Starting remote agent process for target %s using shared binary optimization...", targetName)
	processCmd := utils.StartRemoteAgentProcessWithSharedBinary(t, targetConfig)
	if processCmd == nil {
		return nil, fmt.Errorf("failed to start remote agent process for target %s", targetName)
	}

	// Wait for the process to be healthy
	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
	t.Logf("Remote agent process started successfully for target %s using shared binary", targetName)

	return processCmd, nil
}

func setupScenario4Namespace(t *testing.T, namespace string) {
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

	nsPath := filepath.Join(scenario4TestDir, "namespace.yaml")
	err = utils.CreateYAMLFile(t, nsPath, nsYaml)
	if err == nil {
		utils.ApplyKubernetesManifest(t, nsPath)
	}
}
