package main

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

type RemoteAgentProvider struct{}

func (i *RemoteAgentProvider) getCertificateExpirationOrThumbPrint(certPath string, kind string) (string, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	if kind == "thumbprint" {
		thumbprint := sha1.Sum(cert.Raw)
		return hex.EncodeToString(thumbprint[:]), nil
	} else {
		return cert.NotAfter.Format(time.RFC3339), nil
	}
}

func main() {
	provider := &RemoteAgentProvider{}
	thumbprint, err := provider.getCertificateExpirationOrThumbPrint("test-cert.pem", "thumbprint")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Thumbprint:", thumbprint)
	}

	expiredAt, err := provider.getCertificateExpirationOrThumbPrint("test-cert.pem", "expiration")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("ExpiredAt:", expiredAt)
	}

}
