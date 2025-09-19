/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const loggerName = "providers.cert.k8scert"

var sLog = logger.NewLogger(loggerName)

type K8sCertProviderConfig struct {
	Name      string `json:"name"`
	InCluster bool   `json:"inCluster,omitempty"`
}

type K8sCertProvider struct {
	Config     K8sCertProviderConfig
	Context    *contexts.ManagerContext
	kubeClient kubernetes.Interface
}

func K8sCertProviderConfigFromMap(properties map[string]string) (K8sCertProviderConfig, error) {
	ret := K8sCertProviderConfig{
		InCluster: true, // default to in-cluster
	}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["inCluster"]; ok {
		ret.InCluster = v == "true"
	}
	return ret, nil
}

func (k *K8sCertProvider) InitWithMap(properties map[string]string) error {
	config, err := K8sCertProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (K8sCert): expected K8sCertProviderConfigFromMap: %+v", err)
		return err
	}
	return k.Init(config)
}

func (k *K8sCertProvider) SetContext(ctx *contexts.ManagerContext) {
	k.Context = ctx
}

func (k *K8sCertProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("K8sCert Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (K8sCert): Init()")

	// convert config to K8sCertProviderConfig type
	certConfig, err := toK8sCertProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): expected K8sCertProviderConfig: %+v", err)
		return err
	}

	k.Config = certConfig

	// Initialize Kubernetes client
	var kubeConfig *rest.Config
	if k.Config.InCluster {
		kubeConfig, err = rest.InClusterConfig()
	} else {
		// For out-of-cluster access, would need to load from kubeconfig file
		// This can be implemented later if needed
		err = fmt.Errorf("out-of-cluster configuration not implemented yet")
	}
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to get kubernetes config: %+v", err)
		return err
	}

	k.kubeClient, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create kubernetes client: %+v", err)
		return err
	}

	return nil
}

func toK8sCertProviderConfig(config providers.IProviderConfig) (K8sCertProviderConfig, error) {
	ret := K8sCertProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

// CreateCert creates a self-signed certificate and stores it as a Kubernetes Secret
func (k *K8sCertProvider) CreateCert(ctx context.Context, req cert.CertRequest) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "CreateCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): creating certificate for target %s in namespace %s", req.TargetName, req.Namespace)

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to generate private key: %+v", err)
		return err
	}

	// Set default duration if not specified
	duration := req.Duration
	if duration == 0 {
		duration = 365 * 24 * time.Hour // 1 year default
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Symphony"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    req.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(duration),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add DNS names if specified
	if len(req.DNSNames) > 0 {
		template.DNSNames = req.DNSNames
	}

	// Set default CommonName if not specified
	if req.CommonName == "" {
		template.Subject.CommonName = fmt.Sprintf("%s.symphony.local", req.TargetName)
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create certificate: %+v", err)
		return err
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to marshal private key: %+v", err)
		return err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	// Create Kubernetes Secret
	secretName := fmt.Sprintf("%s-working-cert", req.TargetName)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: req.Namespace,
			Labels: map[string]string{
				"symphony.microsoft.com/managed-by": "symphony",
				"symphony.microsoft.com/target":     req.TargetName,
				"symphony.microsoft.com/cert-type":  "working-cert",
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": privateKeyPEM,
		},
	}

	// Create or update the secret
	_, err = k.kubeClient.CoreV1().Secrets(req.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Update existing secret
			_, err = k.kubeClient.CoreV1().Secrets(req.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to update certificate secret: %+v", err)
				return err
			}
			sLog.InfofCtx(ctx, "  P (K8sCert): updated certificate secret %s in namespace %s", secretName, req.Namespace)
		} else {
			sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create certificate secret: %+v", err)
			return err
		}
	} else {
		sLog.InfofCtx(ctx, "  P (K8sCert): created certificate secret %s in namespace %s", secretName, req.Namespace)
	}

	return nil
}

// DeleteCert deletes the certificate secret for the specified target
func (k *K8sCertProvider) DeleteCert(ctx context.Context, targetName, namespace string) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "DeleteCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): deleting certificate for target %s in namespace %s", targetName, namespace)

	secretName := fmt.Sprintf("%s-working-cert", targetName)
	err = k.kubeClient.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			sLog.InfofCtx(ctx, "  P (K8sCert): certificate secret %s not found (already deleted)", secretName)
			return nil
		}
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to delete certificate secret: %+v", err)
		return err
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): deleted certificate secret %s in namespace %s", secretName, namespace)
	return nil
}

// GetCert retrieves the certificate for the specified target (read-only)
func (k *K8sCertProvider) GetCert(ctx context.Context, targetName, namespace string) (*cert.CertResponse, error) {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "GetCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): getting certificate for target %s in namespace %s", targetName, namespace)

	secretName := fmt.Sprintf("%s-working-cert", targetName)
	secret, err := k.kubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			sLog.InfofCtx(ctx, "  P (K8sCert): certificate secret %s not found", secretName)
			return nil, fmt.Errorf("certificate not found for target %s", targetName)
		}
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to get certificate secret: %+v", err)
		return nil, err
	}

	certPEM := secret.Data["tls.crt"]
	keyPEM := secret.Data["tls.key"]

	if len(certPEM) == 0 || len(keyPEM) == 0 {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): certificate secret %s is missing certificate or key data", secretName)
		return nil, fmt.Errorf("invalid certificate data for target %s", targetName)
	}

	// Parse certificate to get expiration date and serial number
	block, _ := pem.Decode(certPEM)
	if block == nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to decode certificate PEM")
		return nil, fmt.Errorf("invalid certificate format for target %s", targetName)
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to parse certificate: %+v", err)
		return nil, err
	}

	response := &cert.CertResponse{
		PublicKey:    base64.StdEncoding.EncodeToString(certPEM),
		PrivateKey:   base64.StdEncoding.EncodeToString(keyPEM),
		ExpiresAt:    parsedCert.NotAfter,
		SerialNumber: parsedCert.SerialNumber.String(),
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): retrieved certificate for target %s, expires at %v", targetName, parsedCert.NotAfter)
	return response, nil
}

// RotateCert rotates/renews the certificate for the specified target
func (k *K8sCertProvider) RotateCert(ctx context.Context, targetName, namespace string) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "RotateCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): rotating certificate for target %s in namespace %s", targetName, namespace)

	// Create a new certificate with default settings
	req := cert.CertRequest{
		TargetName: targetName,
		Namespace:  namespace,
		Duration:   365 * 24 * time.Hour, // 1 year
		CommonName: fmt.Sprintf("%s.symphony.local", targetName),
		DNSNames:   []string{targetName, fmt.Sprintf("%s.symphony.local", targetName)},
	}

	return k.CreateCert(ctx, req)
}

// CheckCertStatus checks if the certificate is ready and valid
func (k *K8sCertProvider) CheckCertStatus(ctx context.Context, targetName, namespace string) (*cert.CertStatus, error) {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "CheckCertStatus",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): checking certificate status for target %s in namespace %s", targetName, namespace)

	status := &cert.CertStatus{
		Ready:      false,
		LastUpdate: time.Now(),
	}

	secretName := fmt.Sprintf("%s-working-cert", targetName)
	secret, err := k.kubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.Reason = "NotFound"
			status.Message = "Certificate secret not found"
			return status, nil
		}
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to get certificate secret: %+v", err)
		status.Reason = "Error"
		status.Message = err.Error()
		return status, nil
	}

	certPEM := secret.Data["tls.crt"]
	if len(certPEM) == 0 {
		status.Reason = "InvalidData"
		status.Message = "Certificate data is missing"
		return status, nil
	}

	// Parse certificate to check validity
	block, _ := pem.Decode(certPEM)
	if block == nil {
		status.Reason = "InvalidFormat"
		status.Message = "Certificate format is invalid"
		return status, nil
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to parse certificate: %+v", err)
		status.Reason = "ParseError"
		status.Message = err.Error()
		return status, nil
	}

	now := time.Now()
	if now.Before(parsedCert.NotBefore) {
		status.Reason = "NotYetValid"
		status.Message = "Certificate is not yet valid"
		return status, nil
	}

	if now.After(parsedCert.NotAfter) {
		status.Reason = "Expired"
		status.Message = "Certificate has expired"
		return status, nil
	}

	// Check if renewal is needed (30 days before expiration)
	renewalThreshold := parsedCert.NotAfter.Add(-30 * 24 * time.Hour)
	if now.After(renewalThreshold) {
		status.NextRenewal = renewalThreshold
		status.Message = "Certificate needs renewal soon"
	} else {
		status.NextRenewal = renewalThreshold
	}

	status.Ready = true
	status.Reason = "Ready"
	status.Message = "Certificate is valid and ready"

	sLog.InfofCtx(ctx, "  P (K8sCert): certificate status for target %s: ready=%v, reason=%s", targetName, status.Ready, status.Reason)
	return status, nil
}
