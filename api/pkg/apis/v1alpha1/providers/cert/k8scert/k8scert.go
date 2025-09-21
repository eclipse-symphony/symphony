/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
	Config        K8sCertProviderConfig
	Context       *contexts.ManagerContext
	dynamicClient dynamic.Interface
	kubeClient    kubernetes.Interface
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

	k.dynamicClient, err = dynamic.NewForConfig(kubeConfig)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create dynamic kubernetes client: %+v", err)
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

// CreateCert creates a minimal cert-manager Certificate resource matching targets-vendor pattern
func (k *K8sCertProvider) CreateCert(ctx context.Context, req cert.CertRequest) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "CreateCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): creating certificate for target %s in namespace %s", req.TargetName, req.Namespace)

	// Use simple naming pattern like targets-vendor
	certName := fmt.Sprintf("%s-working-cert", req.TargetName)
	secretName := certName

	// Create minimal Certificate resource matching solution-manager pattern
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      certName,
				"namespace": req.Namespace,
			},
			"spec": map[string]interface{}{
				"secretName":  secretName,
				"commonName":  "symphony-service",
				"dnsNames":    req.DNSNames,
				"duration":    req.Duration.String(),
				"renewBefore": req.RenewBefore.String(),
				"issuerRef": map[string]interface{}{
					"name": req.IssuerName,
					"kind": "Issuer",
				},
			},
		},
	}

	// Create the Certificate resource
	certificateGVR := k.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}).Namespace(req.Namespace)

	_, err = certificateGVR.Create(ctx, certificate, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			sLog.InfofCtx(ctx, "  P (K8sCert): certificate %s already exists", certName)
			// Even if certificate already exists, wait for it to be ready
		} else {
			sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create certificate: %+v", err)
			return err
		}
	} else {
		sLog.InfofCtx(ctx, "  P (K8sCert): created certificate %s in namespace %s", certName, req.Namespace)
	}

	// Wait for certificate and secret to be ready
	err = k.waitForCertificateReady(ctx, certName, req.Namespace, secretName)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to wait for certificate ready: %+v", err)
		return err
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): certificate %s is ready with secret %s", certName, secretName)
	return nil
}

// DeleteCert deletes the certificate resource
func (k *K8sCertProvider) DeleteCert(ctx context.Context, targetName, namespace string) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "DeleteCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): deleting certificate for target %s in namespace %s", targetName, namespace)

	certName := fmt.Sprintf("%s-working-cert", targetName)

	// Delete Certificate resource
	certificateGVR := k.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}).Namespace(namespace)

	err = certificateGVR.Delete(ctx, certName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			sLog.InfofCtx(ctx, "  P (K8sCert): certificate %s not found (already deleted)", certName)
			return nil
		}
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to delete certificate: %+v", err)
		return err
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): deleted certificate %s in namespace %s", certName, namespace)
	return nil
}

// GetCert retrieves the certificate from the cert-manager created secret with retry logic
func (k *K8sCertProvider) GetCert(ctx context.Context, targetName, namespace string) (*cert.CertResponse, error) {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "GetCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): getting certificate for target %s in namespace %s", targetName, namespace)

	secretName := fmt.Sprintf("%s-working-cert", targetName)

	// Retry logic: 30 seconds timeout, retry every 2 seconds (safety net for client-side timing issues)
	timeout := time.Now().Add(30 * time.Second)
	retryCount := 0

	for time.Now().Before(timeout) {
		secret, err := k.kubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err == nil {
			// Secret found, check if it has valid certificate data
			certPEM := secret.Data["tls.crt"]
			keyPEM := secret.Data["tls.key"]

			if len(certPEM) > 0 && len(keyPEM) > 0 {
				// Convert PEM format to match target vendor format (replace newlines with spaces)
				// This ensures compatibility with existing test code that parses certificate responses
				publicCert := strings.ReplaceAll(string(certPEM), "\n", " ")
				privateCert := strings.ReplaceAll(string(keyPEM), "\n", " ")

				response := &cert.CertResponse{
					PublicKey:    publicCert,
					PrivateKey:   privateCert,
					ExpiresAt:    time.Now().Add(90 * 24 * time.Hour), // Default 90 days
					SerialNumber: "cert-manager-generated",
				}

				sLog.InfofCtx(ctx, "  P (K8sCert): retrieved certificate for target %s after %d retries", targetName, retryCount)
				return response, nil
			} else {
				sLog.InfofCtx(ctx, "  P (K8sCert): certificate secret %s exists but missing certificate or key data, retrying...", secretName)
			}
		} else {
			if !errors.IsNotFound(err) {
				// If it's not a "not found" error, return immediately
				sLog.ErrorfCtx(ctx, "  P (K8sCert): unexpected error getting certificate secret %s: %v", secretName, err)
				return nil, fmt.Errorf("certificate not found for target %s: %v", targetName, err)
			}
		}

		// Log retry attempt
		retryCount++
		sLog.InfofCtx(ctx, "  P (K8sCert): certificate secret %s not ready yet, retrying in 2 seconds (attempt %d)...", secretName, retryCount)
		time.Sleep(2 * time.Second)
	}

	// 30 seconds timeout reached without finding valid certificate
	sLog.ErrorfCtx(ctx, "  P (K8sCert): certificate secret %s not found after 30 seconds timeout", secretName)
	return nil, fmt.Errorf("certificate not found for target %s after 30 seconds: secret %s not available", targetName, secretName)
}

// RotateCert rotates the certificate by recreating it
func (k *K8sCertProvider) RotateCert(ctx context.Context, targetName, namespace string) error {
	ctx, span := observability.StartSpan("K8sCert Provider", ctx, &map[string]string{
		"method": "RotateCert",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (K8sCert): rotating certificate for target %s in namespace %s", targetName, namespace)

	// Create a new certificate request with default values from solution-manager pattern
	req := cert.CertRequest{
		TargetName:  targetName,
		Namespace:   namespace,
		Duration:    time.Hour * 2160, // 90 days default
		RenewBefore: time.Hour * 360,  // 15 days before expiration
		CommonName:  "symphony-service",
		DNSNames:    []string{targetName, fmt.Sprintf("%s.%s", targetName, namespace)},
		IssuerName:  "symphony-ca-issuer",
	}

	return k.CreateCert(ctx, req)
}

// CheckCertStatus checks if the certificate is ready
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

	certName := fmt.Sprintf("%s-working-cert", targetName)

	// Check Certificate resource status
	certificateGVR := k.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}).Namespace(namespace)

	certificate, err := certificateGVR.Get(ctx, certName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			status.Reason = "NotFound"
			status.Message = "Certificate not found"
			return status, nil
		}
		status.Reason = "Error"
		status.Message = err.Error()
		return status, nil
	}

	// Check if certificate is ready
	if statusObj, found := certificate.Object["status"]; found {
		if statusMap, ok := statusObj.(map[string]interface{}); ok {
			if conditions, found := statusMap["conditions"]; found {
				if conditionsArray, ok := conditions.([]interface{}); ok {
					for _, condition := range conditionsArray {
						if condMap, ok := condition.(map[string]interface{}); ok {
							if condType, found := condMap["type"]; found && strings.EqualFold(condType.(string), "ready") {
								if condStatus, found := condMap["status"]; found && strings.EqualFold(condStatus.(string), "true") {
									status.Ready = true
									status.Reason = "Ready"
									status.Message = "Certificate is ready"
									return status, nil
								}
							}
						}
					}
				}
			}
		}
	}

	status.Reason = "NotReady"
	status.Message = "Certificate is not ready yet"
	return status, nil
}

// waitForCertificateReady waits for Certificate to be ready and secret to have the correct type and content
func (k *K8sCertProvider) waitForCertificateReady(ctx context.Context, certName, namespace, secretName string) error {
	sLog.InfofCtx(ctx, "  P (K8sCert): waiting for certificate %s to be ready in namespace %s", certName, namespace)

	// Create a context with timeout for the whole operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	op := func() error {
		// Check Certificate status
		ready, err := k.checkCertificateStatus(timeoutCtx, certName, namespace)
		if err != nil {
			sLog.ErrorfCtx(timeoutCtx, "  P (K8sCert): error checking certificate status: %v", err)
			return err
		}

		if !ready {
			sLog.ErrorfCtx(timeoutCtx, "  P (K8sCert): certificate %s not ready yet", certName)
			return fmt.Errorf("certificate %s not ready", certName)
		}

		// Check if secret exists and has correct type
		secretReady, err := k.checkSecretReady(timeoutCtx, secretName, namespace)
		if err != nil {
			sLog.ErrorfCtx(timeoutCtx, "  P (K8sCert): error checking secret status: %v", err)
			return err
		}

		if !secretReady {
			sLog.ErrorfCtx(timeoutCtx, "  P (K8sCert): secret %s not ready yet", secretName)
			return fmt.Errorf("secret %s not ready", secretName)
		}

		sLog.InfofCtx(timeoutCtx, "  P (K8sCert): certificate %s and secret %s are ready", certName, secretName)
		return nil
	}

	// Use exponential backoff with the timeout context for cancellation
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 2 * time.Second
	bo.MaxInterval = 10 * time.Second
	// Respect the outer timeout via WithContext
	err := backoff.RetryNotify(op, backoff.WithContext(bo, timeoutCtx), func(err error, duration time.Duration) {
		sLog.InfofCtx(timeoutCtx, "  P (K8sCert): retrying certificate check in %v due to: %v", duration, err)
	})

	if err != nil {
		return fmt.Errorf("timeout waiting for certificate %s to be ready: %s", certName, err.Error())
	}

	return nil
}

// checkCertificateStatus checks if Certificate is ready
func (k *K8sCertProvider) checkCertificateStatus(ctx context.Context, certName, namespace string) (bool, error) {
	certificateGVR := k.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}).Namespace(namespace)

	certificate, err := certificateGVR.Get(ctx, certName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get certificate: %s", err.Error())
	}

	// Check Certificate status conditions
	if status, found := certificate.Object["status"]; found {
		if statusMap, ok := status.(map[string]interface{}); ok {
			if conditions, found := statusMap["conditions"]; found {
				if conditionsArray, ok := conditions.([]interface{}); ok {
					for _, condition := range conditionsArray {
						if condMap, ok := condition.(map[string]interface{}); ok {
							if condType, found := condMap["type"]; found && strings.EqualFold(condType.(string), "ready") {
								if condStatus, found := condMap["status"]; found && strings.EqualFold(condStatus.(string), "true") {
									return true, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return false, nil
}

// checkSecretReady checks if secret exists and has the correct type and content
func (k *K8sCertProvider) checkSecretReady(ctx context.Context, secretName, namespace string) (bool, error) {
	// Try to read both tls.crt and tls.key to verify secret is complete
	secret, err := k.kubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): secret %s not ready yet, waiting...", secretName)
		return false, err // Secret not ready yet
	}

	// Check if secret has the required keys
	if _, hasCrt := secret.Data["tls.crt"]; !hasCrt {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): secret %s missing tls.crt, waiting...", secretName)
		return false, fmt.Errorf("secret missing tls.crt")
	}

	if _, hasKey := secret.Data["tls.key"]; !hasKey {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): secret %s missing tls.key, waiting...", secretName)
		return false, fmt.Errorf("secret missing tls.key")
	}

	return true, nil
}
