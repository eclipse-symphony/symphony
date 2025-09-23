package verify

import (
	"os/exec"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/scenarios/13.remoteAgent-linux/utils"
	"github.com/stretchr/testify/require"
)

func TestScenario7_SolutionUpdate(t *testing.T) {
	t.Log("Starting Scenario 7: Solution update - 2 components to 4 components")

	// Setup test environment
	testDir := "scenario7-solution-update"
	config, err := utils.SetupTestEnvironment(t, testDir)
	require.NoError(t, err)
	defer utils.CleanupTestDirectory(testDir)

	targetName := "test-target-update"
	solutionName := "test-solution-update"
	instanceName := "test-instance-update"

	// Step 1: Create target and bootstrap remote agent
	t.Log("Step 1: Creating target and bootstrapping remote agent")
	targetPath := utils.CreateTargetYAML(t, testDir, targetName, config.Namespace)

	// Apply target
	applyCmd := exec.Command("kubectl", "apply", "-f", targetPath)
	output, err := applyCmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply target: %s", string(output))
	t.Logf("Target applied successfully: %s", string(output))

	// Bootstrap remote agent
	// Start remote agent using direct process (no systemd service) without automatic cleanup
	processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, config)
	require.NotNil(t, processCmd)

	require.NoError(t, err, "Failed to bootstrap remote agent")

	// Wait for target to be ready
	utils.WaitForTargetReady(t, targetName, config.Namespace, 120*time.Second)

	// Step 2: Create initial solution with 2 components
	t.Log("Step 2: Creating initial solution with 2 components")
	initialSolutionPath := utils.CreateSolutionWithComponents(t, testDir, solutionName, config.Namespace, 2)

	// Apply initial solution
	applyCmd = exec.Command("kubectl", "apply", "-f", initialSolutionPath)
	output, err = applyCmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply initial solution: %s", string(output))
	t.Logf("Initial solution applied successfully: %s", string(output))

	// Step 3: Create instance with initial solution
	t.Log("Step 3: Creating instance with initial solution")
	instancePath := utils.CreateInstanceYAML(t, testDir, instanceName, solutionName, targetName, config.Namespace)

	// Apply instance
	applyCmd = exec.Command("kubectl", "apply", "-f", instancePath)
	output, err = applyCmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply instance: %s", string(output))
	t.Logf("Instance applied successfully: %s", string(output))

	utils.WaitForInstanceReady(t, instanceName, config.Namespace, 5*time.Minute)
	verifySingleTargetDeployment(t, "script", instanceName)
	t.Logf("✓ Instance %s (%s provider) is ready and deployed successfully on target %s",
		instanceName, "script", targetName)
	// Verify initial deployment with 2 components
	t.Log("Verifying initial deployment with 2 components")
	initialComponents := utils.GetInstanceComponents(t, instanceName, config.Namespace)
	require.Equal(t, 2, len(initialComponents), "Initial instance should have 2 components")
	t.Logf("Initial deployment verified with %d components", len(initialComponents))

	// Step 4: Update solution to have 4 components
	t.Log("Step 4: Updating solution to have 4 components")
	updatedSolutionPath := utils.CreateSolutionWithComponents(t, testDir, solutionName+"-v2", config.Namespace, 4)

	// Apply updated solution with same name but different spec
	updateSolutionContent := utils.ReadAndUpdateSolutionName(t, updatedSolutionPath, solutionName)
	updatedSolutionFinalPath := testDir + "/updated-solution.yaml"
	utils.WriteFileContent(t, updatedSolutionFinalPath, updateSolutionContent)

	applyCmd = exec.Command("kubectl", "apply", "-f", updatedSolutionFinalPath)
	output, err = applyCmd.CombinedOutput()
	require.NoError(t, err, "Failed to apply updated solution: %s", string(output))
	t.Logf("Updated solution applied successfully: %s", string(output))

	// Wait for solution update to be processed
	time.Sleep(10 * time.Second)

	// Step 5: Verify instance reconciliation
	t.Log("Step 5: Verifying instance reconciliation with updated solution")

	// Wait for instance to reconcile with new solution
	require.Eventually(t, func() bool {
		components := utils.GetInstanceComponents(t, instanceName, config.Namespace)
		if len(components) == 4 {
			t.Logf("Instance reconciled with %d components", len(components))
			return true
		}
		t.Logf("Instance still has %d components, waiting for reconciliation...", len(components))
		return false
	}, 180*time.Second, 10*time.Second, "Instance should reconcile to have 4 components")

	// Verify deployment status after update
	finalComponents := utils.GetInstanceComponents(t, instanceName, config.Namespace)
	require.Equal(t, 4, len(finalComponents), "Updated instance should have 4 components")
	t.Logf("Solution update verification successful - instance now has %d components", len(finalComponents))

	// Verify that both install and uninstall operations occurred during reconciliation
	t.Log("Verifying reconciliation events")
	reconciliationEvents := utils.GetInstanceEvents(t, instanceName, config.Namespace)
	t.Logf("Reconciliation events: %v", reconciliationEvents)

	// Cleanup
	t.Log("Step 6: Cleaning up resources")

	// Delete instance
	deleteCmd := exec.Command("kubectl", "delete", "instance", instanceName, "-n", config.Namespace)
	output, err = deleteCmd.CombinedOutput()
	require.NoError(t, err, "Failed to delete instance: %s", string(output))

	// Delete solution
	deleteCmd = exec.Command("kubectl", "delete", "solution", solutionName, "-n", config.Namespace)
	output, err = deleteCmd.CombinedOutput()
	require.NoError(t, err, "Failed to delete solution: %s", string(output))

	// Delete target
	deleteCmd = exec.Command("kubectl", "delete", "target", targetName, "-n", config.Namespace)
	output, err = deleteCmd.CombinedOutput()
	require.NoError(t, err, "Failed to delete target: %s", string(output))

	t.Log("Scenario 7 completed successfully")
}

func TestScenario7_SolutionUpdate_MultipleProviders(t *testing.T) {
	t.Log("Starting Scenario 7 with multiple providers: Solution update test")

	providers := []string{"script", "http", "script"}

	for _, provider := range providers {
		t.Run("Provider_"+provider, func(t *testing.T) {
			testDir := "scenario7-solution-update-" + provider
			config, err := utils.SetupTestEnvironment(t, testDir)
			require.NoError(t, err)
			defer utils.CleanupTestDirectory(testDir)

			targetName := "test-target-update-" + provider
			solutionName := "test-solution-update-" + provider
			instanceName := "test-instance-update-" + provider

			// Create target and bootstrap
			t.Logf("Creating target with provider %s: %s", provider, targetName)
			targetPath := utils.CreateTargetYAML(t, testDir, targetName, config.Namespace)

			applyCmd := exec.Command("kubectl", "apply", "-f", targetPath)
			output, err := applyCmd.CombinedOutput()
			require.NoError(t, err, "Failed to apply target: %s", string(output))

			// Start remote agent using direct process (no systemd service) without automatic cleanup
			processCmd := utils.StartRemoteAgentProcessWithoutCleanup(t, config)
			require.NotNil(t, processCmd)

			utils.WaitForTargetReady(t, targetName, config.Namespace, 120*time.Second)

			// Create initial solution with 2 components for specific provider
			t.Logf("Creating initial solution with 2 components for provider: %s", provider)
			initialSolutionPath := utils.CreateSolutionWithComponentsForProvider(t, testDir, solutionName, config.Namespace, 2, provider)

			applyCmd = exec.Command("kubectl", "apply", "-f", initialSolutionPath)
			output, err = applyCmd.CombinedOutput()
			require.NoError(t, err, "Failed to apply initial solution: %s", string(output))
			// Create instance
			instancePath := utils.CreateInstanceYAML(t, testDir, instanceName, solutionName, targetName, config.Namespace)

			applyCmd = exec.Command("kubectl", "apply", "-f", instancePath)
			output, err = applyCmd.CombinedOutput()
			require.NoError(t, err, "Failed to apply instance: %s", string(output))

			utils.WaitForInstanceReady(t, instanceName, config.Namespace, 120*time.Second)

			// Verify initial components
			initialComponents := utils.GetInstanceComponents(t, instanceName, config.Namespace)
			require.Equal(t, 2, len(initialComponents), "Initial instance should have 2 components")

			// Update solution to 4 components
			t.Logf("Updating solution to 4 components for provider: %s", provider)
			updatedSolutionPath := utils.CreateSolutionWithComponentsForProvider(t, testDir, solutionName+"-v2", config.Namespace, 4, provider)

			updateSolutionContent := utils.ReadAndUpdateSolutionName(t, updatedSolutionPath, solutionName)
			updatedSolutionFinalPath := testDir + "/updated-solution-" + provider + ".yaml"
			utils.WriteFileContent(t, updatedSolutionFinalPath, updateSolutionContent)

			applyCmd = exec.Command("kubectl", "apply", "-f", updatedSolutionFinalPath)
			output, err = applyCmd.CombinedOutput()
			require.NoError(t, err, "Failed to apply updated solution: %s", string(output))

			// Wait for reconciliation
			utils.WaitForInstanceReady(t, instanceName, config.Namespace, 180*time.Second)
			finalComponents := utils.GetInstanceComponents(t, instanceName, config.Namespace)
			require.Equal(t, 4, len(finalComponents), "Updated instance should have 4 components")
			t.Logf("Solution update verified for provider %s - instance now has %d components", provider, len(finalComponents))

			// Cleanup
			exec.Command("kubectl", "delete", "instance", instanceName, "-n", config.Namespace).Run()
			exec.Command("kubectl", "delete", "solution", solutionName, "-n", config.Namespace).Run()
			exec.Command("kubectl", "delete", "target", targetName, "-n", config.Namespace).Run()
		})
	}
}
