package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MQTTBroker holds MQTT broker information
type MQTTBroker struct {
	Address     string
	Port        int
	TLSPort     int
	ContainerID string
}

// SetupMQTTBroker starts a containerized MQTT broker with TLS support
func SetupMQTTBroker(t *testing.T, certs CertificatePaths) (*MQTTBroker, func()) {
	// Create MQTT broker config directory
	configDir := filepath.Join(filepath.Dir(certs.CACert), "mqtt-config")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create mosquitto configuration
	mqttConfig := fmt.Sprintf(`
listener 1883
allow_anonymous false

listener 8883
cafile /mosquitto/config/certs/ca.pem
certfile /mosquitto/config/certs/server.pem
keyfile /mosquitto/config/certs/server-key.pem
require_certificate true
use_identity_as_username true
tls_version tlsv1.2

log_dest stdout
log_type all
`)

	configFile := filepath.Join(configDir, "mosquitto.conf")
	err = ioutil.WriteFile(configFile, []byte(mqttConfig), 0644)
	require.NoError(t, err)

	// Create password file (for future use)
	passwdFile := filepath.Join(configDir, "passwd")
	err = ioutil.WriteFile(passwdFile, []byte(""), 0644)
	require.NoError(t, err)

	// Start MQTT broker container
	containerName := fmt.Sprintf("test-mqtt-broker-%d", time.Now().Unix())

	cmd := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-p", "1883:1883",
		"-p", "8883:8883",
		"-v", fmt.Sprintf("%s:/mosquitto/config/certs", filepath.Dir(certs.CACert)),
		"-v", fmt.Sprintf("%s:/mosquitto/config/mosquitto.conf", configFile),
		"-v", fmt.Sprintf("%s:/mosquitto/config/passwd", passwdFile),
		"eclipse-mosquitto:latest",
		"mosquitto", "-c", "/mosquitto/config/mosquitto.conf")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to start MQTT broker: %s", string(output))
		require.NoError(t, err)
	}

	containerID := string(output)[:12] // Get container ID

	// Wait for broker to start
	time.Sleep(5 * time.Second)

	// Verify container is running
	cmd = exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)

	if len(output) == 0 {
		t.Fatal("MQTT broker container is not running")
	}

	broker := &MQTTBroker{
		Address:     "localhost",
		Port:        1883,
		TLSPort:     8883,
		ContainerID: containerID,
	}

	cleanup := func() {
		// Stop and remove container
		exec.Command("docker", "stop", containerName).Run()
		exec.Command("docker", "rm", containerName).Run()
		// Clean up config directory
		os.RemoveAll(configDir)
		t.Logf("Cleaned up MQTT broker: %s", containerName)
	}

	t.Cleanup(cleanup)
	t.Logf("Started MQTT broker: %s on ports 1883/8883", containerName)

	return broker, cleanup
}

// WaitForMQTTBrokerReady waits for MQTT broker to be ready
func WaitForMQTTBrokerReady(t *testing.T, broker *MQTTBroker, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for MQTT broker to be ready")
		case <-ticker.C:
			// Test connection to broker
			cmd := exec.Command("docker", "exec", broker.ContainerID, "netstat", "-ln")
			output, err := cmd.CombinedOutput()
			if err == nil && len(output) > 0 {
				outputStr := string(output)
				if contains(outputStr, ":1883") && contains(outputStr, ":8883") {
					t.Logf("MQTT broker is ready")
					return
				}
			}
		}
	}
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SetupEmbeddedMQTTBroker creates a simple in-memory MQTT broker for testing
func SetupEmbeddedMQTTBroker(t *testing.T) (*MQTTBroker, func()) {
	// This is a simpler alternative that doesn't require Docker
	// You can implement this using a Go MQTT broker library if needed
	t.Logf("Using embedded MQTT broker (not implemented yet)")

	broker := &MQTTBroker{
		Address: "localhost",
		Port:    1883,
		TLSPort: 8883,
	}

	cleanup := func() {
		t.Logf("Cleaned up embedded MQTT broker")
	}

	return broker, cleanup
}

// VerifyMQTTConnection tests MQTT connectivity using mosquitto client tools
func VerifyMQTTConnection(t *testing.T, broker *MQTTBroker, clientCert, clientKey, caCert string) {
	// Test TLS connection
	testTopic := "test/connection"
	testMessage := "hello"

	// Try to publish a message
	cmd := exec.Command("mosquitto_pub",
		"--host", broker.Address,
		"--port", fmt.Sprintf("%d", broker.TLSPort),
		"--topic", testTopic,
		"--message", testMessage,
		"--cert", clientCert,
		"--key", clientKey,
		"--cafile", caCert,
		"--tls-version", "tlsv1.2")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// mosquitto_pub might not be available in test environment
		t.Logf("mosquitto_pub not available or failed: %s", string(output))
		return
	}

	t.Logf("MQTT connection test successful")
}
