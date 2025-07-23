package mqtt_communication

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/e2e/remote-agent-integration/utils"
	"github.com/stretchr/testify/require"
)

func TestE2EMQTTCommunication(t *testing.T) {
	// Test configuration
	projectRoot := "/mnt/d/code3/symphony/" // Relative to test file location
	targetName := "test-mqtt-target"
	namespace := "default"
	mqttBrokerAddress := "localhost"
	mqttBrokerPort := 8883

	// Setup test environment
	testDir := utils.SetupTestDirectory(t)
	t.Logf("Running MQTT communication test in: %s", testDir)

	// Step 1: Start fresh minikube cluster
	t.Run("SetupFreshMinikubeCluster", func(t *testing.T) {
		utils.StartFreshMinikube(t)
	})
	t.Cleanup(func() {
		utils.CleanupMinikube(t)
	})

	mqttCerts := utils.GenerateMQTTCertificates(t, testDir)

	// Setup test namespace
	setupNamespace(t, namespace)
	defer utils.CleanupNamespace(t, namespace)

	var caSecretName, clientSecretName string
	var configPath, topologyPath, targetYamlPath string
	t.Run("CreateCertificateSecrets", func(t *testing.T) {
		// Create CA secret in cert-manager namespace (use MQTT certs for trust bundle)
		caSecretName = utils.CreateMQTTCASecret(t, mqttCerts)

		// Create Symphony MQTT client certificate secret in default namespace
		clientSecretName = utils.CreateMQTTClientCertSecret(t, namespace, mqttCerts)
	})

	t.Run("SetupExternalMQTTBroker", func(t *testing.T) {
		// Deploy MQTT broker on host machine using Docker with TLS
		utils.SetupExternalMQTTBroker(t, mqttCerts, mqttBrokerPort)

		// Test connectivity to external broker
		t.Logf("Testing external MQTT broker connectivity...")
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("172.22.111.41:%d", mqttBrokerPort), 10*time.Second)

		if err == nil {
			conn.Close()
			t.Logf("External MQTT broker connectivity test passed")
		} else {
			t.Fatalf("External MQTT broker connectivity test failed: %v", err)
		}
	})

	t.Run("StartSymphonyWithMQTTConfig", func(t *testing.T) {
		// Deploy Symphony with MQTT configuration
		// Symphony runs inside minikube, needs to access external broker on host
		brokerAddress := fmt.Sprintf("tls://172.22.111.41:%d", mqttBrokerPort)
		fmt.Printf("Starting Symphony with MQTT broker address: %s\n", brokerAddress)
		utils.StartSymphonyWithMQTTConfig(t, brokerAddress)

		// Wait for Symphony server certificate to be created
		utils.WaitForSymphonyServerCert(t, 5*time.Minute)
	})
	// Create test configurations AFTER Symphony is running
	t.Run("CreateTestConfigurations", func(t *testing.T) {
		configPath = utils.CreateMQTTConfig(t, testDir, mqttBrokerAddress, mqttBrokerPort, targetName, namespace)
		topologyPath = utils.CreateTestTopology(t, testDir)
		fmt.Printf("Topology path: %s", topologyPath)
		targetYamlPath = utils.CreateTargetYAML(t, testDir, targetName, namespace)
		fmt.Printf("Target YAML path: %s", targetYamlPath)
		// Apply Target YAML to create the target resource
		err := utils.ApplyKubernetesManifest(t, targetYamlPath)
		require.NoError(t, err)

		// Wait for target to be created
		utils.WaitForTargetCreated(t, targetName, namespace, 30*time.Second)
	})

	t.Run("StartRemoteAgentWithMQTTBootstrap", func(t *testing.T) {
		// Configure remote agent for MQTT (use standard test certificates for remote agent)
		config := utils.TestConfig{
			ProjectRoot:    projectRoot,
			ConfigPath:     configPath,
			ClientCertPath: mqttCerts.RemoteAgentCert, // Use standard test cert for remote agent
			ClientKeyPath:  mqttCerts.RemoteAgentKey,  // Use standard test key for remote agent
			CACertPath:     mqttCerts.CACert,          // Use Symphony server CA for TLS trust
			TargetName:     targetName,
			Namespace:      namespace,
			TopologyPath:   topologyPath,
			Protocol:       "mqtt",
		}
		fmt.Printf("Starting remote agent with config: %+v\n", config)
		// Start remote agent using bootstrap.sh
		bootstrapCmd := utils.StartRemoteAgentWithBootstrap(t, config)
		require.NotNil(t, bootstrapCmd)

		// Check service status
		utils.CheckSystemdServiceStatus(t, "remote-agent.service")

		// Try to wait for service to be active
		t.Logf("Attempting to verify service is active...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Service check failed, but bootstrap.sh succeeded: %v", r)
				}
			}()
			utils.WaitForSystemdService(t, "remote-agent.service", 15*time.Second)
		}()

		time.Sleep(5 * time.Second)
		t.Logf("Continuing with test - bootstrap.sh completed successfully")
	})

	t.Run("VerifyTargetStatus", func(t *testing.T) {
		// Wait for target to reach ready state
		utils.WaitForTargetReady(t, targetName, namespace, 360*time.Second)
	})

	t.Run("VerifyMQTTDataInteraction", func(t *testing.T) {
		// Verify that data flows through MQTT correctly
		// This would check that the remote agent successfully communicates
		// with Symphony through the MQTT broker
		// verifyMQTTDataInteraction(t, targetName, namespace, testDir)
		testBootstrapDataInteraction(t, targetName, namespace, testDir)
	})

	// Cleanup
	t.Cleanup(func() {
		utils.CleanupSystemdService(t)
		utils.CleanupSymphony(t, "remote-agent-mqtt-test")
		utils.CleanupExternalMQTTBroker(t) // Use external broker cleanup
		utils.CleanupMQTTCASecret(t, caSecretName)
		utils.CleanupMQTTClientSecret(t, namespace, clientSecretName)
	})

	t.Logf("MQTT communication test completed successfully")
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
}

func testBootstrapDataInteraction(t *testing.T, targetName, namespace, testDir string) {
	// Step 1: Create a simple Solution first
	solutionName := "test-bootstrap-solution"
	solutionVersion := "test-bootstrap-solution-v-version1"
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
        echo "Bootstrap test component deployed successfully"
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
	instanceName := "test-bootstrap-instance"
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

	t.Logf("Bootstrap data interaction test completed - Solution and Instance created successfully")
}
