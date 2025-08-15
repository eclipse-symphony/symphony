package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CertificatePaths holds paths to all generated certificates
type CertificatePaths struct {
	CACert     string
	CAKey      string
	ServerCert string
	ServerKey  string
	ClientCert string
	ClientKey  string
}

// MQTTCertificatePaths holds paths to MQTT-specific certificates
type MQTTCertificatePaths struct {
	CACert             string
	CAKey              string
	MQTTServerCert     string
	MQTTServerKey      string
	SymphonyServerCert string
	SymphonyServerKey  string
	RemoteAgentCert    string
	RemoteAgentKey     string
}

// GenerateTestCertificates generates a complete set of test certificates
func GenerateTestCertificates(t *testing.T, testDir string) CertificatePaths {
	// Generate CA certificate
	caCert, caKey := generateCA(t)

	// Generate server certificate (for MQTT broker and Symphony server)
	serverCert, serverKey := generateServerCert(t, caCert, caKey, "localhost")

	// Generate client certificate (for remote agent)
	clientCert, clientKey := generateClientCert(t, caCert, caKey, "remote-agent-client")

	// Define paths
	paths := CertificatePaths{
		CACert:     filepath.Join(testDir, "ca.pem"),
		CAKey:      filepath.Join(testDir, "ca-key.pem"),
		ServerCert: filepath.Join(testDir, "server.pem"),
		ServerKey:  filepath.Join(testDir, "server-key.pem"),
		ClientCert: filepath.Join(testDir, "client.pem"),
		ClientKey:  filepath.Join(testDir, "client-key.pem"),
	}

	// Save all certificates
	err := saveCertificate(paths.CACert, caCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.CAKey, caKey)
	require.NoError(t, err)

	err = saveCertificate(paths.ServerCert, serverCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.ServerKey, serverKey)
	require.NoError(t, err)

	err = saveCertificate(paths.ClientCert, clientCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.ClientKey, clientKey)
	require.NoError(t, err)

	t.Logf("Generated test certificates in %s", testDir)
	return paths
}

func generateCA(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Symphony Test"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "MyRootCA", // This is what Symphony will check for trust
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

func generateServerCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey, hostname string) (*x509.Certificate, *rsa.PrivateKey) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Build comprehensive list of IP addresses to include in certificate
	// Start with standard localhost addresses
	ipAddresses := []net.IP{
		net.IPv4(127, 0, 0, 1), // localhost IPv4
		net.IPv6loopback,       // localhost IPv6
		net.IPv4zero,           // 0.0.0.0 - any IPv4
	}

	// Dynamically detect all available network interfaces and their IPs
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			// Skip loopback and down interfaces, but include all others
			if iface.Flags&net.FlagUp == 0 {
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil {
					// Add both IPv4 and IPv6 addresses
					ipAddresses = append(ipAddresses, ip)
					t.Logf("Added detected IP to certificate: %s (interface: %s)", ip.String(), iface.Name)
				}
			}
		}
	} else {
		t.Logf("Warning: Could not detect network interfaces: %v", err)
	}

	// Also try to detect common container/VM host IPs dynamically
	commonHostIPs := []string{
		"host.docker.internal",
		"host.minikube.internal",
		"gateway.docker.internal",
	}

	for _, hostname := range commonHostIPs {
		if ips, err := net.LookupIP(hostname); err == nil {
			for _, ip := range ips {
				ipAddresses = append(ipAddresses, ip)
				t.Logf("Added resolved IP to certificate: %s (from %s)", ip.String(), hostname)
			}
		}
	}

	// Add some fallback IPs for common scenarios (but fewer than before since we have dynamic detection)
	fallbackIPs := []string{
		"172.17.0.1",   // Docker bridge IP
		"192.168.49.1", // Common minikube host IP
		"10.0.2.2",     // VirtualBox host IP
	}

	for _, ipStr := range fallbackIPs {
		if ip := net.ParseIP(ipStr); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		}
	}

	// Add comprehensive DNS names for maximum compatibility
	dnsNames := []string{
		hostname,
		"localhost",
		"*.local",
		"*.localhost",
		"host.docker.internal",   // Docker Desktop
		"host.minikube.internal", // Minikube
	}

	// Log the final list of IPs in the certificate for debugging
	t.Logf("Certificate will be valid for %d IP addresses:", len(ipAddresses))
	for i, ip := range ipAddresses {
		t.Logf("  [%d] %s", i+1, ip.String())
	}

	// Log DNS names for debugging
	t.Logf("Certificate will be valid for %d DNS names:", len(dnsNames))
	for i, name := range dnsNames {
		t.Logf("  DNS[%d] %s", i+1, name)
	}

	// Create certificate template with very permissive settings for testing
	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:  []string{"Symphony Test"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, // Allow both server and client auth
		IPAddresses: ipAddresses,
		DNSNames:    dnsNames,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	require.NoError(t, err)

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

func generateClientCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey, commonName string) (*x509.Certificate, *rsa.PrivateKey) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization:  []string{"Symphony Test"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    commonName, // Use the provided common name for client cert
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	require.NoError(t, err)

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

func saveCertificate(filename string, cert *x509.Certificate) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

func savePrivateKey(filename string, key *rsa.PrivateKey) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// CleanupCertificates removes all generated certificate files
func CleanupCertificates(paths CertificatePaths) {
	os.Remove(paths.CACert)
	os.Remove(paths.CAKey)
	os.Remove(paths.ServerCert)
	os.Remove(paths.ServerKey)
	os.Remove(paths.ClientCert)
	os.Remove(paths.ClientKey)
}

// GenerateMQTTCertificates generates a complete set of MQTT-specific test certificates
func GenerateMQTTCertificates(t *testing.T, testDir string) MQTTCertificatePaths {
	// Generate CA certificate (same CA signs all certificates)
	caCert, caKey := generateCA(t)

	// Generate MQTT server certificate (for MQTT broker)
	mqttServerCert, mqttServerKey := generateServerCert(t, caCert, caKey, "localhost")

	// Generate Symphony server certificate (Symphony as MQTT client)
	symphonyServerCert, symphonyServerKey := generateClientCert(t, caCert, caKey, "symphony-client")

	// Generate remote agent certificate (Remote agent as MQTT client)
	remoteAgentCert, remoteAgentKey := generateClientCert(t, caCert, caKey, "remote-agent-client")

	// Define paths with MQTT-specific naming
	paths := MQTTCertificatePaths{
		CACert:             filepath.Join(testDir, "ca.crt"),
		CAKey:              filepath.Join(testDir, "ca.key"),
		MQTTServerCert:     filepath.Join(testDir, "mqtt-server.crt"),
		MQTTServerKey:      filepath.Join(testDir, "mqtt-server.key"),
		SymphonyServerCert: filepath.Join(testDir, "symphony-server.crt"),
		SymphonyServerKey:  filepath.Join(testDir, "symphony-server.key"),
		RemoteAgentCert:    filepath.Join(testDir, "remote-agent.crt"),
		RemoteAgentKey:     filepath.Join(testDir, "remote-agent.key"),
	}

	// Save all certificates
	err := saveCertificate(paths.CACert, caCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.CAKey, caKey)
	require.NoError(t, err)

	err = saveCertificate(paths.MQTTServerCert, mqttServerCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.MQTTServerKey, mqttServerKey)
	require.NoError(t, err)

	err = saveCertificate(paths.SymphonyServerCert, symphonyServerCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.SymphonyServerKey, symphonyServerKey)
	require.NoError(t, err)

	err = saveCertificate(paths.RemoteAgentCert, remoteAgentCert)
	require.NoError(t, err)
	err = savePrivateKey(paths.RemoteAgentKey, remoteAgentKey)
	require.NoError(t, err)

	t.Logf("Generated MQTT test certificates in %s", testDir)
	t.Logf("  CA Certificate: %s", paths.CACert)
	t.Logf("  MQTT Server Certificate: %s", paths.MQTTServerCert)
	t.Logf("  Symphony Server Certificate: %s", paths.SymphonyServerCert)
	t.Logf("  Remote Agent Certificate: %s", paths.RemoteAgentCert)
	return paths
}

// CleanupMQTTCertificates removes all generated MQTT certificate files
func CleanupMQTTCertificates(paths MQTTCertificatePaths) {
	os.Remove(paths.CACert)
	os.Remove(paths.CAKey)
	os.Remove(paths.MQTTServerCert)
	os.Remove(paths.MQTTServerKey)
	os.Remove(paths.SymphonyServerCert)
	os.Remove(paths.SymphonyServerKey)
	os.Remove(paths.RemoteAgentCert)
	os.Remove(paths.RemoteAgentKey)
}
