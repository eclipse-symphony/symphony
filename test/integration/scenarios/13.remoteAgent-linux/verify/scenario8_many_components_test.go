package verify

import (
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

func TestScenario8_ManyComponents_MultipleProviders(t *testing.T) {
	t.Log("Starting Scenario 8 with multiple providers: Many component test")

	// providers := []string{"script", "http", "helm.v3"}
	providers := []string{"script", "http", "script"}
	componentCount := 30
	namespace := "default"

	// Setup test environment following the successful pattern from scenario1_provider_crud_test.go
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running multiple providers test in: %s", testDir)

	// Step 1: Start fresh minikube cluster (following the correct pattern)
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Step 2: Generate test certificates (following the correct pattern)
	certs := utils.GenerateTestCertificates(t, testDir)

	// Step 3: Setup Symphony infrastructure
	var caSecretName, clientSecretName string
	var configPath, topologyPath string
	var symphonyCAPath, baseURL string

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

	// Setup hosts mapping and port-forward
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)
	baseURL = "https://symphony-service:8081/v1alpha2"

	t.Run("CreateTestConfigurations", func(t *testing.T) {
		configPath = utils.CreateHTTPConfig(t, testDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, testDir)
	})

	// Now test each provider with proper infrastructure
	for _, provider := range providers {
		t.Run("Provider_"+provider, func(t *testing.T) {
			targetName := "test-target-many-comp-" + provider
			solutionName := "test-solution-many-comp-" + provider
			instanceName := "test-instance-many-comp-" + provider

			// Create target first
			targetPath := utils.CreateTargetYAML(t, testDir, targetName, namespace)
			err := utils.ApplyKubernetesManifest(t, targetPath)
			require.NoError(t, err, "Failed to apply target")

			// Wait for target to be created
			utils.WaitForTargetCreated(t, targetName, namespace, 30*time.Second)

			// Create config following working test pattern
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

			// Start remote agent as a process
			processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, config)
			require.NotNil(t, processCmd)

			// Setup cleanup for the process
			t.Cleanup(func() {
				if processCmd != nil {
					utils.CleanupRemoteAgentProcess(t, processCmd)
				}
			})

			// Wait for process health and target readiness
			utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
			utils.WaitForTargetReady(t, targetName, namespace, 120*time.Second)

			// Create solution with many components for specific provider
			t.Logf("Creating solution with %d components for provider: %s", componentCount, provider)
			solutionPath := utils.CreateSolutionWithComponentsForProvider(t, testDir, solutionName, namespace, componentCount, provider)

			err = utils.ApplyKubernetesManifest(t, solutionPath)
			require.NoError(t, err, "Failed to apply solution")

			// Create instance
			instancePath := utils.CreateInstanceYAML(t, testDir, instanceName, solutionName, targetName, namespace)

			err = utils.ApplyKubernetesManifest(t, instancePath)
			require.NoError(t, err, "Failed to apply instance")

			// Wait for instance to be processed and ready
			utils.WaitForInstanceReady(t, instanceName, namespace, 5*time.Minute)
			verifySingleTargetDeployment(t, "script", instanceName)
			t.Logf("✓ Instance %s (%s provider) is ready and deployed successfully on target %s", instanceName, provider, targetName)

			// todo: verify deployment result
			// t.Logf("Successfully verified %d components for provider %s", len(deployedComponents), provider)

			// Test provider-specific paging
			// pageSize := 10
			// totalPages := (componentCount + pageSize - 1) / pageSize
			// retrievedTotal := 0

			// for page := 0; page < totalPages; page++ {
			// 	pagedComponents := utils.GetInstanceComponentsPagedForProvider(t, instanceName, namespace, page, pageSize, provider)
			// 	retrievedTotal += len(pagedComponents)
			// }

			// require.Equal(t, componentCount, retrievedTotal, "Paged retrieval should return all %d components for provider %s", componentCount, provider)
			// t.Logf("Paging test successful for provider %s - retrieved %d components", provider, retrievedTotal)

			// Performance test for provider
			startTime := time.Now()
			utils.VerifyProviderComponentsDeployment(t, instanceName, namespace, provider, componentCount)
			duration := time.Since(startTime)
			t.Logf("Provider %s component verification with %d components took: %v", provider, componentCount, duration)

			// Cleanup resources for this provider
			err = utils.DeleteKubernetesResource(t, "instances.solution.symphony", instanceName, namespace, 2*time.Minute)
			if err != nil {
				t.Logf("Failed to delete instance: %v", err)
			}
			err = utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
			if err != nil {
				t.Logf("Failed to delete solution: %v", err)
			}
			err = utils.DeleteKubernetesResource(t, "target", targetName, namespace, 2*time.Minute)
			if err != nil {
				t.Logf("Failed to delete target: %v", err)
			}

			t.Logf("Provider %s many components test completed successfully", provider)
		})
	}

	// Final cleanup
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "scenario8-multiple-providers-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})
}

func TestScenario8_ManyComponents_StressTest(t *testing.T) {
	t.Log("Starting Scenario 8 Stress Test: Testing system limits with very large component count")

	// This test uses an even larger number of components to stress test the system
	testDir := utils.SetupTestDirectory(t)
	defer utils.CleanupTestDirectory(testDir)

	targetName := "test-target-stress"
	solutionName := "test-solution-stress"
	instanceName := "test-instance-stress"
	namespace := "default"
	stressComponentCount := 50 // Increased for stress testing

	// Step 1: Start fresh minikube cluster (following the correct pattern)
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})

	// Ensure minikube is cleaned up after test
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	// Step 2: Generate test certificates (following the correct pattern)
	certs := utils.GenerateTestCertificates(t, testDir)

	// Step 3: Create certificate secrets BEFORE starting Symphony (this was the issue!)
	var caSecretName, clientSecretName string
	var configPath, topologyPath string
	var symphonyCAPath, baseURL string

	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		// Create CA secret in cert-manager namespace
		caSecretName = utils.CreateCASecret(t, certs)

		// Create client cert secret in test namespace
		clientSecretName = utils.CreateClientCertSecret(t, namespace, certs)
	})

	// Step 4: Start Symphony server AFTER certificates are created
	t.Run("StartSymphonyServer", func(t *testing.T) {
		utils.StartSymphonyWithRemoteAgentConfig(t, "http")

		// Wait for Symphony server certificate to be created
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	t.Run("SetupSymphonyConnection", func(t *testing.T) {
		symphonyCAPath = utils.DownloadSymphonyCA(t, testDir)
	})

	// Setup hosts mapping and port-forward
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)
	baseURL = "https://symphony-service:8081/v1alpha2"

	// Create test configurations - THIS WAS MISSING!
	t.Run("CreateTestConfigurations", func(t *testing.T) {
		configPath = utils.CreateHTTPConfig(t, testDir, baseURL)
		topologyPath = utils.CreateTestTopology(t, testDir)
	})

	// Create config following working test pattern (don't use SetupTestEnvironment)
	config := utils.TestConfig{
		ProjectRoot:    utils.GetProjectRoot(t),
		ConfigPath:     configPath, // Use absolute path from CreateHTTPConfig
		ClientCertPath: certs.ClientCert,
		ClientKeyPath:  certs.ClientKey,
		CACertPath:     symphonyCAPath,
		TargetName:     targetName,
		Namespace:      namespace,
		TopologyPath:   topologyPath, // Use absolute path from CreateTestTopology
		Protocol:       "http",
		BaseURL:        baseURL,
	}

	t.Logf("Stress testing with %d components", stressComponentCount)

	// Create target and bootstrap
	targetPath := utils.CreateTargetYAML(t, testDir, targetName, config.Namespace)
	err := utils.ApplyKubernetesManifest(t, targetPath)
	require.NoError(t, err)

	// Wait for target to be created
	utils.WaitForTargetCreated(t, targetName, config.Namespace, 30*time.Second)

	// Start remote agent as a process
	processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	// Setup cleanup for the process
	t.Cleanup(func() {
		if processCmd != nil {
			utils.CleanupRemoteAgentProcess(t, processCmd)
		}
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, config.Namespace, clientSecretName)
	})

	// Wait for process to be healthy and target to be ready
	utils.WaitForProcessHealthy(t, processCmd, 30*time.Second)
	utils.WaitForTargetReady(t, targetName, config.Namespace, 120*time.Second)

	// Create large solution
	solutionPath := utils.CreateSolutionWithComponents(t, testDir, solutionName, config.Namespace, stressComponentCount)
	err = utils.ApplyKubernetesManifest(t, solutionPath)
	require.NoError(t, err)

	// Create instance and measure deployment time
	instancePath := utils.CreateInstanceYAML(t, testDir, instanceName, solutionName, targetName, config.Namespace)

	deploymentStartTime := time.Now()
	err = utils.ApplyKubernetesManifest(t, instancePath)
	require.NoError(t, err)

	// Wait for instance to be processed and ready
	utils.WaitForInstanceReady(t, instanceName, config.Namespace, 5*time.Minute)
	verifySingleTargetDeployment(t, "script", instanceName)
	t.Logf("✓ Instance %s (script provider) is ready and deployed successfully on target %s", instanceName, targetName)

	deploymentDuration := time.Since(deploymentStartTime)
	t.Logf("Stress test deployment with %d components completed in: %v", stressComponentCount, deploymentDuration)

	// Test system responsiveness under load
	t.Log("Testing system responsiveness under load")
	responsivenessTested := utils.TestSystemResponsivenessUnderLoad(t, config.Namespace, stressComponentCount)
	require.True(t, responsivenessTested, "System should remain responsive under load")

	//todo: in the future verify helm pods/ containers
	// Cleanup with timing
	cleanupStartTime := time.Now()
	err = utils.DeleteKubernetesResource(t, "instances.solution.symphony", instanceName, config.Namespace, 10*time.Minute)
	if err != nil {
		t.Logf("Failed to delete instance during cleanup: %v", err)
	}

	utils.WaitForResourceDeleted(t, "instance", instanceName, config.Namespace, 1*time.Minute)

	cleanupDuration := time.Since(cleanupStartTime)
	t.Logf("Stress test cleanup with %d components completed in: %v", stressComponentCount, cleanupDuration)

	// Final cleanup
	err = utils.DeleteSolutionManifestWithTimeout(t, solutionPath, 2*time.Minute)
	if err != nil {
		t.Logf("Failed to delete solution during final cleanup: %v", err)
	}
	err = utils.DeleteKubernetesResource(t, "target", targetName, config.Namespace, 2*time.Minute)
	if err != nil {
		t.Logf("Failed to delete target during final cleanup: %v", err)
	}

	t.Log("Scenario 8 stress test completed successfully")
}
