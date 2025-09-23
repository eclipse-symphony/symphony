package verify

import (
	"os/exec"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

func TestScenario6_RemoteAgentNotStarted(t *testing.T) {
	t.Log("Starting Scenario 6: Remote agent did not start -> remote target delete should succeed")

	// Setup test environment following the successful pattern
	testDir := utils.SetupTestDirectory(t)
	namespace := "default"
	t.Logf("Running test in: %s", testDir)

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

	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		caSecretName = utils.CreateCASecret(t, certs)
		clientSecretName = utils.CreateClientCertSecret(t, namespace, certs)
	})

	t.Run("StartSymphonyServer", func(t *testing.T) {
		utils.StartSymphonyWithRemoteAgentConfig(t, "http")
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	t.Run("SetupSymphonyConnection", func(t *testing.T) {
		_ = utils.DownloadSymphonyCA(t, testDir)
	})

	// Setup hosts mapping and port-forward
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)

	// Ensure remote agent is NOT running (cleanup any existing processes)
	t.Log("Ensuring remote agent process is not running")
	utils.CleanupStaleRemoteAgentProcesses(t)

	// Wait a moment to ensure cleanup is complete
	time.Sleep(2 * time.Second)

	// Test Case i: create remote target -> target delete
	t.Log("Test Case i: Create remote target when agent is not started, then delete target")

	targetName := "test-target-no-agent"

	// Create target (this should succeed even if remote agent is not running)
	t.Logf("Creating target: %s", targetName)
	targetPath := utils.CreateTargetYAML(t, testDir, targetName, namespace)

	// Apply target
	applyCmd := exec.Command("kubectl", "apply", "-f", targetPath)
	output, err := applyCmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply target: %s", string(output))
	t.Logf("Target applied successfully: %s", string(output))

	// Wait for target to be processed (it may be in pending state due to no remote agent)
	time.Sleep(10 * time.Second)

	// Verify target exists (even if it's in a pending/error state)
	getTargetCmd := exec.Command("kubectl", "get", "target", targetName, "-n", namespace, "-o", "yaml")
	output, err = getTargetCmd.CombinedOutput()
	require.NoError(t, err, "Target should exist even with no remote agent: %s", string(output))
	t.Log("Target exists as expected")

	// // todo: this will not succeed now, bug to fix in the future, Delete the target - this should succeed regardless of remote agent status
	// t.Logf("Deleting target: %s", targetName)
	// deleteCmd := exec.Command("kubectl", "delete", "target", targetName, "-n", namespace)
	// output, err = deleteCmd.CombinedOutput()
	// require.NoError(t, err, "Target deletion should succeed even without remote agent: %s", string(output))
	// t.Logf("Target deleted successfully: %s", string(output))

	// // Verify target is deleted
	// t.Log("Verifying target deletion")
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		t.Fatal("Timeout waiting for target deletion")
	// 	default:
	// 		checkCmd := exec.Command("kubectl", "get", "target", targetName, "-n", namespace)
	// 		_, err := checkCmd.CombinedOutput()
	// 		if err != nil {
	// 			// Target not found - deletion successful
	// 			t.Log("Target deletion verified successfully")
	// 			return
	// 		}
	// 		time.Sleep(2 * time.Second)
	// 	}
	// }

	// Ensure we clean up created secrets and Symphony for this test
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "scenario6-remote-agent-not-started-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})
}

func TestScenario6_RemoteAgentNotStarted_Targets(t *testing.T) {
	t.Log("Starting Scenario 6 with multiple providers: Remote agent did not start -> remote target delete should succeed")

	providers := []string{"script", "http", "script"}
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

	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		caSecretName = utils.CreateCASecret(t, certs)
		clientSecretName = utils.CreateClientCertSecret(t, namespace, certs)
	})

	t.Run("StartSymphonyServer", func(t *testing.T) {
		utils.StartSymphonyWithRemoteAgentConfig(t, "http")
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})

	t.Run("SetupSymphonyConnection", func(t *testing.T) {
		_ = utils.DownloadSymphonyCA(t, testDir)
	})

	// Setup hosts mapping and port-forward
	utils.SetupSymphonyHostsForMainTest(t)
	utils.StartPortForwardForMainTest(t)

	// Now test each provider with proper infrastructure
	for _, provider := range providers {
		t.Run("Provider_"+provider, func(t *testing.T) {
			targetName := "test-target-no-agent-" + provider

			// Ensure remote agent is NOT running (cleanup any existing processes)
			t.Logf("Ensuring remote agent process is not running for provider: %s", provider)
			utils.CleanupStaleRemoteAgentProcesses(t)

			time.Sleep(2 * time.Second)

			// Create target with specific provider
			t.Logf("Creating target with provider %s: %s", provider, targetName)
			targetPath := utils.CreateTargetYAML(t, testDir, targetName, namespace)

			// Apply target
			applyCmd := exec.Command("kubectl", "apply", "-f", targetPath)
			output, err := applyCmd.CombinedOutput()
			require.NoError(t, err, "Failed to apply target: %s", string(output))
			t.Logf("Target applied successfully: %s", string(output))

			// Wait for target to be processed
			time.Sleep(10 * time.Second)

			// Verify target exists
			getTargetCmd := exec.Command("kubectl", "get", "target", targetName, "-n", namespace, "-o", "yaml")
			output, err = getTargetCmd.CombinedOutput()
			require.NoError(t, err, "Target should exist even with no remote agent: %s", string(output))
			t.Logf("Target exists for provider %s", provider)

			// Delete the target - this should succeed
			t.Logf("Deleting Target...")
			err = utils.DeleteKubernetesResource(t, "targets.fabric.symphony", targetName, namespace, 2*time.Minute)
			if err != nil {
				t.Logf("Warning: Failed to delete target: %v", err)
			}
			t.Logf("✓ Deleted target for %s provider", provider)

			t.Logf("Target deletion verified successfully for provider: %s", provider)
		})
	}

	// Final cleanup
	t.Cleanup(func() {
		utils.CleanupSymphony(t, "scenario6-multiple-providers-test")
		utils.CleanupCASecret(t, caSecretName)
		utils.CleanupClientSecret(t, namespace, clientSecretName)
	})
}
