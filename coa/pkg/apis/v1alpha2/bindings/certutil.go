// certutil.go: PEM/certificate helper functions for MQTT binding and other packages
package bindings

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// ValidatePEMCertificates checks if the PEM data contains at least one valid X.509 certificate
func ValidatePEMCertificates(pemData []byte) error {
	if len(pemData) == 0 {
		return fmt.Errorf("CA certificate file is empty")
	}
	rest := pemData
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		if _, err := x509.ParseCertificate(block.Bytes); err == nil {
			return nil // found at least one valid cert
		}
	}
	return fmt.Errorf("CA certificate file does not contain a valid X.509 certificate")
}

// LoadCACertPool loads and validates a CA certificate file, returning a CertPool
func LoadCACertPool(caCertPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	if err := ValidatePEMCertificates(caCert); err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate - invalid PEM format or corrupted certificate")
	}
	return caCertPool, nil
}
