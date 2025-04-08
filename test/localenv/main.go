package main

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

type RemoteAgentProvider struct{}

func (i *RemoteAgentProvider) getCertificateExpirationOrThumbPrintOrSubject(certPath string, kind string) (string, error) {
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

	switch strings.ToLower(kind) {
	case "thumbprint":
		thumbprint := sha1.Sum(cert.Raw)
		return hex.EncodeToString(thumbprint[:]), nil
	case "expiration":
		return cert.NotAfter.Format(time.RFC3339), nil
	case "subject":
		return cert.Subject.String(), nil
	default:
		return "", fmt.Errorf("invalid kind: %s", kind)
	}
}

func main() {
	provider := &RemoteAgentProvider{}
	thumbprint, err := provider.getCertificateExpirationOrThumbPrintOrSubject("test-cert.pem", "thumbprint")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Thumbprint:", thumbprint)
	}

	expiredAt, err := provider.getCertificateExpirationOrThumbPrintOrSubject("test-cert.pem", "expiration")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("ExpiredAt:", expiredAt)
	}

	subject, err := provider.getCertificateExpirationOrThumbPrintOrSubject("test-cert.pem", "subject")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Subject:", subject)
	}
}
