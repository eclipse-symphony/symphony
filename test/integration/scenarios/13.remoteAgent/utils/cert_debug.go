package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// DebugCertificateInfo prints detailed information about a certificate file
func DebugCertificateInfo(t *testing.T, certPath, certType string) {
	t.Logf("=== DEBUG %s at %s ===", certType, certPath)

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		t.Logf("ERROR: Failed to read certificate file %s: %v", certPath, err)
		return
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		t.Logf("ERROR: Failed to decode PEM block from %s", certPath)
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Logf("ERROR: Failed to parse certificate %s: %v", certPath, err)
		return
	}

	t.Logf("Certificate Subject: %s", cert.Subject.String())
	t.Logf("Certificate Issuer: %s", cert.Issuer.String())
	t.Logf("Certificate Serial Number: %s", cert.SerialNumber.String())
	t.Logf("Certificate Valid From: %s", cert.NotBefore.Format(time.RFC3339))
	t.Logf("Certificate Valid Until: %s", cert.NotAfter.Format(time.RFC3339))
	t.Logf("Certificate Is CA: %t", cert.IsCA)

	if len(cert.DNSNames) > 0 {
		t.Logf("Certificate DNS Names: %v", cert.DNSNames)
	}

	if len(cert.IPAddresses) > 0 {
		t.Logf("Certificate IP Addresses: %v", cert.IPAddresses)
	}

	if len(cert.Extensions) > 0 {
		t.Logf("Certificate has %d extensions", len(cert.Extensions))
	}

	// Check if certificate is expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		t.Logf("WARNING: Certificate is not yet valid (starts %s)", cert.NotBefore.Format(time.RFC3339))
	}
	if now.After(cert.NotAfter) {
		t.Logf("WARNING: Certificate has expired (expired %s)", cert.NotAfter.Format(time.RFC3339))
	}

	t.Logf("=== END DEBUG %s ===", certType)
}

// DebugMQTTBrokerCertificates prints information about all MQTT broker certificates
func DebugMQTTBrokerCertificates(t *testing.T, testDir string) {
	t.Logf("=== DEBUG MQTT BROKER CERTIFICATES ===")

	// Check for common certificate files
	certFiles := []struct {
		name string
		path string
	}{
		{"CA Certificate", filepath.Join(testDir, "ca.crt")},
		{"MQTT Server Certificate", filepath.Join(testDir, "mqtt-server.crt")},
		{"Symphony Server Certificate", filepath.Join(testDir, "symphony-server.crt")},
		{"Remote Agent Certificate", filepath.Join(testDir, "remote-agent.crt")},
	}

	for _, certFile := range certFiles {
		if FileExists(certFile.path) {
			DebugCertificateInfo(t, certFile.path, certFile.name)
		} else {
			t.Logf("Certificate file not found: %s", certFile.path)
		}
	}

	t.Logf("=== END DEBUG MQTT BROKER CERTIFICATES ===")
}

// DebugTLSConnection attempts to connect to MQTT broker and debug TLS handshake
func DebugTLSConnection(t *testing.T, brokerAddress string, port int, caCertPath, clientCertPath, clientKeyPath string) {
	t.Logf("=== DEBUG TLS CONNECTION to %s:%d ===", brokerAddress, port)

	// Load CA certificate
	var caCertPool *x509.CertPool
	if caCertPath != "" {
		caCertBytes, err := os.ReadFile(caCertPath)
		if err != nil {
			t.Logf("ERROR: Failed to read CA certificate: %v", err)
			return
		}

		caCertPool = x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCertBytes) {
			t.Logf("ERROR: Failed to parse CA certificate")
			return
		}
		t.Logf("✓ Loaded CA certificate from %s", caCertPath)
	}

	// Load client certificate and key
	var clientCerts []tls.Certificate
	if clientCertPath != "" && clientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			t.Logf("ERROR: Failed to load client certificate: %v", err)
			return
		}
		clientCerts = []tls.Certificate{clientCert}
		t.Logf("✓ Loaded client certificate from %s", clientCertPath)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: clientCerts,
		ServerName:   brokerAddress, // Important: set server name for SNI
	}

	// Try connecting
	address := fmt.Sprintf("%s:%d", brokerAddress, port)
	t.Logf("Attempting TLS connection to %s", address)

	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		t.Logf("ERROR: TLS connection failed: %v", err)

		// Try with InsecureSkipVerify to see if it's a certificate issue
		tlsConfig.InsecureSkipVerify = true
		t.Logf("Retrying with InsecureSkipVerify=true...")

		conn2, err2 := tls.Dial("tcp", address, tlsConfig)
		if err2 != nil {
			t.Logf("ERROR: Even with InsecureSkipVerify, connection failed: %v", err2)
		} else {
			t.Logf("✓ Connection succeeded with InsecureSkipVerify=true")
			t.Logf("This indicates a certificate validation issue, not a network connectivity issue")

			// Get server certificate info
			state := conn2.ConnectionState()
			if len(state.PeerCertificates) > 0 {
				serverCert := state.PeerCertificates[0]
				t.Logf("Server certificate subject: %s", serverCert.Subject.String())
				t.Logf("Server certificate issuer: %s", serverCert.Issuer.String())
				t.Logf("Server certificate DNS names: %v", serverCert.DNSNames)
				t.Logf("Server certificate IP addresses: %v", serverCert.IPAddresses)
			}

			conn2.Close()
		}
		return
	}

	t.Logf("✓ TLS connection successful!")

	// Get connection state
	state := conn.ConnectionState()
	t.Logf("TLS Version: %x", state.Version)
	t.Logf("Cipher Suite: %x", state.CipherSuite)
	t.Logf("Server certificates: %d", len(state.PeerCertificates))

	if len(state.PeerCertificates) > 0 {
		serverCert := state.PeerCertificates[0]
		t.Logf("Server certificate subject: %s", serverCert.Subject.String())
		t.Logf("Server certificate issuer: %s", serverCert.Issuer.String())
	}

	conn.Close()
	t.Logf("=== END DEBUG TLS CONNECTION ===")
}

// DebugMQTTSecrets prints information about MQTT-related Kubernetes secrets
func DebugMQTTSecrets(t *testing.T, namespace string) {
	t.Logf("=== DEBUG MQTT KUBERNETES SECRETS ===")

	secretNames := []string{
		"mqtt-ca",
		"mqtt-client-secret",
		"remote-agent-client-secret",
		"mqtt-server-certs",
	}

	for _, secretName := range secretNames {
		t.Logf("Checking secret: %s in namespace %s", secretName, namespace)

		// Try to get the secret data
		cmd := fmt.Sprintf("kubectl get secret %s -n %s -o yaml", secretName, namespace)
		if _, err := executeCommand(cmd); err == nil {
			t.Logf("Secret %s exists with data", secretName)
			// Don't print the full secret for security, just confirm it exists
		} else {
			t.Logf("Secret %s not found or error: %v", secretName, err)
		}
	}

	t.Logf("=== END DEBUG MQTT KUBERNETES SECRETS ===")
}

// DebugSymphonyPodCertificates checks certificates mounted in Symphony pods
func DebugSymphonyPodCertificates(t *testing.T) {
	t.Logf("=== DEBUG SYMPHONY POD CERTIFICATES ===")

	// Get Symphony API pod name
	podCmd := "kubectl get pods -n default -l app.kubernetes.io/name=symphony-api -o jsonpath='{.items[0].metadata.name}'"
	podName, err := executeCommand(podCmd)
	if err != nil {
		t.Logf("Failed to get Symphony API pod name: %v", err)
		return
	}

	if podName == "" {
		t.Logf("No Symphony API pod found")
		return
	}

	t.Logf("Found Symphony API pod: %s", podName)

	// Check mounted certificates in the pod
	certPaths := []string{
		"/etc/mqtt-ca/ca.crt",
		"/etc/mqtt-client/client.crt",
		"/etc/mqtt-client/client.key",
	}

	for _, certPath := range certPaths {
		cmd := fmt.Sprintf("kubectl exec %s -n default -- ls -la %s", podName, certPath)
		if output, err := executeCommand(cmd); err == nil {
			t.Logf("Certificate found in pod at %s: %s", certPath, output)

			// Also try to get certificate info
			if certPath != "/etc/mqtt-client/client.key" { // Don't cat private keys
				catCmd := fmt.Sprintf("kubectl exec %s -n default -- cat %s", podName, certPath)
				if certContent, err := executeCommand(catCmd); err == nil {
					// Parse and display certificate info
					t.Logf("Certificate content at %s (first 200 chars): %.200s...", certPath, certContent)
				}
			}
		} else {
			t.Logf("Certificate not found in pod at %s: %v", certPath, err)
		}
	}

	t.Logf("=== END DEBUG SYMPHONY POD CERTIFICATES ===")
}

// executeCommand is a helper to execute shell commands
func executeCommand(cmd string) (string, error) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	command := exec.Command(parts[0], parts[1:]...)
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.ReadFile(filename)
	return err == nil
}

// TestMQTTCertificateChain tests the complete certificate chain
func TestMQTTCertificateChain(t *testing.T, caCertPath, serverCertPath string) {
	t.Logf("=== TESTING MQTT CERTIFICATE CHAIN ===")

	// Load CA certificate
	caCertBytes, err := os.ReadFile(caCertPath)
	if err != nil {
		t.Logf("ERROR: Failed to read CA certificate: %v", err)
		return
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertBytes) {
		t.Logf("ERROR: Failed to parse CA certificate")
		return
	}

	// Load server certificate
	serverCertBytes, err := os.ReadFile(serverCertPath)
	if err != nil {
		t.Logf("ERROR: Failed to read server certificate: %v", err)
		return
	}

	block, _ := pem.Decode(serverCertBytes)
	if block == nil {
		t.Logf("ERROR: Failed to decode server certificate PEM")
		return
	}

	serverCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Logf("ERROR: Failed to parse server certificate: %v", err)
		return
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots: caCertPool,
	}

	chains, err := serverCert.Verify(opts)
	if err != nil {
		t.Logf("ERROR: Certificate chain verification failed: %v", err)
	} else {
		t.Logf("✓ Certificate chain verification successful")
		t.Logf("Found %d certificate chain(s)", len(chains))
	}

	t.Logf("=== END TESTING MQTT CERTIFICATE CHAIN ===")
}
