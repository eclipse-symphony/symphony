/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/cert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8SCertProviderConfig struct {
	Name            string `json:"name"`
	DefaultDuration string `json:"defaultDuration"`
	RenewBefore     string `json:"renewBefore"`
}

type K8SCertProvider struct {
	Config        K8SCertProviderConfig
	Context       context.Context
	K8sClient     kubernetes.Interface
	DynamicClient dynamic.Interface
}

func (p *K8SCertProvider) ID() string {
	return "k8s-cert"
}

func (p *K8SCertProvider) SetContext(ctx context.Context) {
	p.Context = ctx
}

func toK8SCertProviderConfig(config providers.IProviderConfig) (K8SCertProviderConfig, error) {
	ret := K8SCertProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (p *K8SCertProvider) Init(config providers.IProviderConfig) error {
	aConfig, err := toK8SCertProviderConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}
	p.Config = aConfig

	// Get in-cluster config
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	p.K8sClient = clientset

	// Create dynamic client for cert-manager CRDs
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	p.DynamicClient = dynamicClient

	return nil
}

// getConfigDuration reads the defaultDuration from provider configuration
func (p *K8SCertProvider) getConfigDuration() time.Duration {
	if p.Config.DefaultDuration == "" {
		return 4320 * time.Hour // 180 days default
	}

	duration, err := time.ParseDuration(p.Config.DefaultDuration)
	if err != nil {
		return 4320 * time.Hour // 180 days default
	}

	return duration
}

// getConfigRenewBefore reads the renewBefore from provider configuration
func (p *K8SCertProvider) getConfigRenewBefore() time.Duration {
	if p.Config.RenewBefore == "" {
		return 360 * time.Hour // 15 days default
	}

	renewBefore, err := time.ParseDuration(p.Config.RenewBefore)
	if err != nil {
		return 360 * time.Hour // 15 days default
	}

	return renewBefore
}

// parseCertificateInfo extracts serial number and expiration time from PEM-encoded certificate data
func parseCertificateInfo(certData []byte) (string, time.Time, error) {
	// Decode PEM block
	block, _ := pem.Decode(certData)
	if block == nil {
		return "", time.Time{}, fmt.Errorf("failed to decode PEM block")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.SerialNumber.String(), cert.NotAfter, nil
}

func (p *K8SCertProvider) CreateCert(ctx context.Context, req cert.CertRequest) error {
	// Validate required fields
	if err := p.validateCertRequest(req); err != nil {
		return fmt.Errorf("invalid certificate request: %w", err)
	}

	// Define the Certificate resource
	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	// Always use provider config for duration and renewBefore
	duration := p.getConfigDuration()
	renewBefore := p.getConfigRenewBefore()

	// Use consistent naming: targetname-working-cert
	secretName := fmt.Sprintf("%s-working-cert", req.TargetName)

	// Build certificate spec with proper field handling
	spec, err := p.buildCertificateSpec(req, secretName, duration, renewBefore)
	if err != nil {
		return fmt.Errorf("failed to build certificate spec: %w", err)
	}

	// Create the Certificate object
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      req.TargetName,
				"namespace": req.Namespace,
			},
			"spec": spec,
		},
	}

	// Create the certificate with better error handling
	_, err = p.DynamicClient.Resource(certificateGVR).Namespace(req.Namespace).Create(
		ctx, certificate, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return fmt.Errorf("certificate '%s' already exists in namespace '%s'", req.TargetName, req.Namespace)
		}
		return fmt.Errorf("failed to create certificate '%s' in namespace '%s': %w", req.TargetName, req.Namespace, err)
	}

	return nil
}

// validateCertRequest validates the required fields in the certificate request
func (p *K8SCertProvider) validateCertRequest(req cert.CertRequest) error {
	if req.TargetName == "" {
		return fmt.Errorf("targetName is required")
	}
	if req.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if req.IssuerName == "" {
		return fmt.Errorf("issuerName is required")
	}
	if req.CommonName == "" {
		return fmt.Errorf("commonName is required")
	}
	return nil
}

// buildCertificateSpec builds the certificate spec with proper field handling
func (p *K8SCertProvider) buildCertificateSpec(req cert.CertRequest, secretName string, duration, renewBefore time.Duration) (map[string]interface{}, error) {
	spec := map[string]interface{}{
		"secretName": secretName,
		"issuerRef": map[string]interface{}{
			"name": req.IssuerName,
			"kind": "Issuer",
		},
		"commonName":  req.CommonName,
		"duration":    duration.String(),
		"renewBefore": renewBefore.String(),
	}

	// Only add dnsNames if it's not empty
	if len(req.DNSNames) > 0 {
		spec["dnsNames"] = req.DNSNames
	}

	// Only add subject if it's not empty to avoid issues with nil maps
	if req.Subject != nil && len(req.Subject) > 0 {
		spec["subject"] = req.Subject
	}

	return spec, nil
}

func (p *K8SCertProvider) DeleteCert(ctx context.Context, targetName, namespace string) error {
	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	err := p.DynamicClient.Resource(certificateGVR).Namespace(namespace).Delete(
		ctx, targetName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}

func (p *K8SCertProvider) GetCert(ctx context.Context, targetName, namespace string) (*cert.CertResponse, error) {
	// Get the Certificate resource to find the secret name
	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	var certificate *unstructured.Unstructured
	certificate, err := p.DynamicClient.Resource(certificateGVR).Namespace(namespace).Get(
		ctx, targetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate for %s after retries: %w", targetName, err)
	}

	// Extract the secret name from the certificate spec
	spec, ok := certificate.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid certificate spec")
	}

	secretName, ok := spec["secretName"].(string)
	if !ok {
		return nil, fmt.Errorf("secret name not found in certificate spec")
	}

	// Get the secret containing the certificate
	secret, err := p.K8sClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	// Extract certificate data
	certData, exists := secret.Data["tls.crt"]
	if !exists {
		return nil, fmt.Errorf("certificate data not found in secret")
	}

	// Extract private key data
	keyData, exists := secret.Data["tls.key"]
	if !exists {
		return nil, fmt.Errorf("private key data not found in secret")
	}

	// Parse certificate to get real serial number and expiration time
	serialNumber, expiresAt, err := parseCertificateInfo(certData)
	if err != nil {
		// If parsing fails, return basic info
		return &cert.CertResponse{
			PublicKey:    string(certData),
			PrivateKey:   string(keyData),
			SerialNumber: "parsing-failed",
			ExpiresAt:    time.Now().Add(90 * 24 * time.Hour), // default fallback
		}, nil
	}

	return &cert.CertResponse{
		PublicKey:    string(certData),
		PrivateKey:   string(keyData),
		SerialNumber: serialNumber,
		ExpiresAt:    expiresAt,
	}, nil
}

func (p *K8SCertProvider) CheckCertStatus(ctx context.Context, targetName, namespace string) (*cert.CertStatus, error) {
	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	certificate, err := p.DynamicClient.Resource(certificateGVR).Namespace(namespace).Get(
		ctx, targetName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	// Check the status
	status, ok := certificate.Object["status"].(map[string]interface{})
	if !ok {
		return &cert.CertStatus{
			Ready:      false,
			LastUpdate: time.Now(),
		}, nil
	}

	conditions, ok := status["conditions"].([]interface{})
	if !ok || len(conditions) == 0 {
		return &cert.CertStatus{
			Ready:      false,
			LastUpdate: time.Now(),
		}, nil
	}

	// Check the Ready condition
	for _, condition := range conditions {
		condMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		if condType, ok := condMap["type"].(string); ok && condType == "Ready" {
			if condStatus, ok := condMap["status"].(string); ok {
				if condStatus == "True" {
					return &cert.CertStatus{
						Ready:      true,
						LastUpdate: time.Now(),
					}, nil
				} else {
					reason, _ := condMap["reason"].(string)
					message, _ := condMap["message"].(string)
					return &cert.CertStatus{
						Ready:      false,
						Reason:     reason,
						Message:    message,
						LastUpdate: time.Now(),
					}, nil
				}
			}
		}
	}

	return &cert.CertStatus{
		Ready:      false,
		LastUpdate: time.Now(),
	}, nil
}
