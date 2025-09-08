/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cert

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	cLog = logger.NewLogger("coa.runtime")
)

type CertManager struct {
	managers.Manager
	StateProvider  states.IStateProvider
	SecretProvider secret.ISecretProvider
	Config         CertManagerConfig
}

type CertManagerConfig struct {
	CAIssuer               string
	ServiceName            string
	WorkingCertDuration    string
	WorkingCertRenewBefore string
}

func (c *CertManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := c.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}

	stateProvider, err := managers.GetPersistentStateProvider(config, providers)
	if err == nil {
		c.StateProvider = stateProvider
	} else {
		return err
	}

	secretProvider, err := managers.GetSecretProvider(config, providers)
	if err == nil {
		if sp, ok := secretProvider.(secret.ISecretProvider); ok {
			c.SecretProvider = sp
		} else {
			// Try to get from providers map directly
			for _, p := range providers {
				if sp, ok := p.(secret.ISecretProvider); ok {
					c.SecretProvider = sp
					break
				}
			}
			if c.SecretProvider == nil {
				return v1alpha2.NewCOAError(nil, "secret provider not found", v1alpha2.MissingConfig)
			}
		}
	} else {
		return err
	}

	// Initialize config with defaults
	c.Config = CertManagerConfig{
		CAIssuer:               getConfigValue(config, "caIssuer", "symphony-issuer"),
		ServiceName:            getConfigValue(config, "serviceName", "symphony-service"),
		WorkingCertDuration:    getConfigValue(config, "workingCertDuration", "2160h"),   // 90 days
		WorkingCertRenewBefore: getConfigValue(config, "workingCertRenewBefore", "360h"), // 15 days
	}

	return nil
}

func getConfigValue(config managers.ManagerConfig, key, defaultValue string) string {
	if val, exists := config.Properties[key]; exists && val != "" {
		return val
	}
	return defaultValue
}

// CreateWorkingCert creates a working certificate for the specified target
func (c *CertManager) CreateWorkingCert(ctx context.Context, targetName, namespace string) error {
	cLog.InfofCtx(ctx, "Creating working cert for target %s in namespace %s", targetName, namespace)

	subject := fmt.Sprintf("CN=%s-%s.%s", namespace, targetName, c.Config.ServiceName)
	secretName := fmt.Sprintf("%s-tls", targetName)

	// Create a new GroupVersionKind for the certificate
	gvk := schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "Certificate",
	}

	// Create a new unstructured object for the certificate
	cert := &unstructured.Unstructured{}
	cert.SetGroupVersionKind(gvk)
	cert.SetName(targetName)
	cert.SetNamespace(namespace)

	spec := map[string]interface{}{
		"secretName":  secretName,
		"duration":    c.Config.WorkingCertDuration,
		"renewBefore": c.Config.WorkingCertRenewBefore,
		"commonName":  subject,
		"dnsNames": []string{
			subject,
		},
		"issuerRef": map[string]interface{}{
			"name": c.Config.CAIssuer,
			"kind": "Issuer",
		},
		"subject": map[string]interface{}{
			"organizations": []interface{}{
				c.Config.ServiceName,
			},
		},
		"privateKey": map[string]interface{}{
			"algorithm": "RSA",
			"size":      2048,
		},
	}

	cert.Object["spec"] = spec

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   targetName,
			Body: cert.Object,
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     gvk.Group,
			"version":   gvk.Version,
			"resource":  "certificates",
			"kind":      gvk.Kind,
		},
	}

	// Check if Certificate already exists
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     gvk.Group,
			"version":   gvk.Version,
			"resource":  "certificates",
			"kind":      gvk.Kind,
		},
	}

	_, err := c.StateProvider.Get(ctx, getRequest)
	if err == nil {
		cLog.InfofCtx(ctx, "Certificate %s already exists, skipping creation", targetName)
		return nil
	}

	// Certificate doesn't exist, create it
	jsonData, _ := json.Marshal(upsertRequest)
	cLog.InfofCtx(ctx, "Creating certificate object - %s", jsonData)
	_, err = c.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		cLog.ErrorfCtx(ctx, "Failed to create certificate: %s", err.Error())
		return err
	}

	cLog.InfofCtx(ctx, "Successfully created working cert for target %s", targetName)
	return nil
}

// DeleteWorkingCert deletes the working certificate for the specified target
func (c *CertManager) DeleteWorkingCert(ctx context.Context, targetName, namespace string) error {
	cLog.InfofCtx(ctx, "Deleting working cert for target %s in namespace %s", targetName, namespace)

	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	// first check if exists
	_, err := c.StateProvider.Get(ctx, getRequest)
	if err != nil {
		cLog.ErrorfCtx(ctx, "Working cert %s not found, cannot delete: %s", targetName, err.Error())
		return err
	}

	// if found,  then  delete
	deleteRequest := states.DeleteRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	err = c.StateProvider.Delete(ctx, deleteRequest)
	if err != nil && !v1alpha2.IsNotFound(err) {
		cLog.ErrorfCtx(ctx, "Failed to delete certificate: %s", err.Error())
		return err
	}

	// double check deletion
	_, err = c.StateProvider.Get(ctx, getRequest)
	if v1alpha2.IsNotFound(err) {
		cLog.InfofCtx(ctx, "Successfully deleted working cert for target %s", targetName)
		return nil
	}

	cLog.ErrorfCtx(ctx, "Certificate %s still exists after delete", targetName)
	if err != nil {
		return err
	}
	return fmt.Errorf("certificate %s still exists after delete", targetName)
}

// GetWorkingCert retrieves the working certificate for the specified target (read-only)
func (c *CertManager) GetWorkingCert(ctx context.Context, targetName, namespace string) (string, string, error) {
	cLog.InfofCtx(ctx, "Getting working cert for target %s in namespace %s", targetName, namespace)

	secretName := fmt.Sprintf("%s-tls", targetName)
	evalCtx := utils.EvaluationContext{Namespace: namespace}

	// Check if certificate exists first
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	_, err := c.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return "", "", fmt.Errorf("working certificate not found for target %s: %w", targetName, err)
	}

	// Read the certificate and private key from the secret
	public, err := c.SecretProvider.Read(ctx, secretName, "tls.crt", evalCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to read public certificate: %w", err)
	}

	private, err := c.SecretProvider.Read(ctx, secretName, "tls.key", evalCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key: %w", err)
	}

	// Format certificates (remove newlines)
	public = strings.ReplaceAll(public, "\n", " ")
	private = strings.ReplaceAll(private, "\n", " ")

	cLog.InfofCtx(ctx, "Successfully retrieved working cert for target %s", targetName)
	return public, private, nil
}

// CheckCertificateReady checks if the certificate is ready and the secret is available
func (c *CertManager) CheckCertificateReady(ctx context.Context, targetName, namespace string) (bool, error) {
	// Check Certificate status
	ready, err := c.checkCertificateStatus(ctx, targetName, namespace)
	if err != nil {
		return false, err
	}

	if !ready {
		return false, nil
	}

	// Check if secret exists and has correct type
	secretName := fmt.Sprintf("%s-tls", targetName)
	secretReady, err := c.checkSecretReady(ctx, secretName, namespace)
	if err != nil {
		return false, err
	}

	return secretReady, nil
}

// checkCertificateStatus checks if Certificate is ready
func (c *CertManager) checkCertificateStatus(ctx context.Context, certName, namespace string) (bool, error) {
	getRequest := states.GetRequest{
		ID: certName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	entry, err := c.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return false, fmt.Errorf("failed to get certificate: %s", err.Error())
	}

	// Check Certificate status conditions
	if status, found := entry.Body.(map[string]interface{})["status"]; found {
		if statusMap, ok := status.(map[string]interface{}); ok {
			if conditions, found := statusMap["conditions"]; found {
				if conditionsArray, ok := conditions.([]interface{}); ok {
					for _, condition := range conditionsArray {
						if condMap, ok := condition.(map[string]interface{}); ok {
							if condType, found := condMap["type"]; found && strings.ToLower(condType.(string)) == "ready" {
								if condStatus, found := condMap["status"]; found && strings.ToLower(condStatus.(string)) == "true" {
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
func (c *CertManager) checkSecretReady(ctx context.Context, secretName, namespace string) (bool, error) {
	evalCtx := utils.EvaluationContext{Namespace: namespace}

	// Try to read both tls.crt and tls.key to verify secret is complete
	_, err := c.SecretProvider.Read(ctx, secretName, "tls.crt", evalCtx)
	if err != nil {
		return false, nil // Secret not ready yet
	}

	_, err = c.SecretProvider.Read(ctx, secretName, "tls.key", evalCtx)
	if err != nil {
		return false, nil // Secret not complete yet
	}

	return true, nil
}

// WaitForCertificateReady waits for Certificate to be ready and secret to have the correct type and content
func (c *CertManager) WaitForCertificateReady(ctx context.Context, targetName, namespace string) error {
	cLog.InfofCtx(ctx, "Waiting for certificate %s to be ready in namespace %s", targetName, namespace)

	// Create a context with timeout for the whole operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for certificate %s to be ready", targetName)
		case <-ticker.C:
			ready, err := c.CheckCertificateReady(timeoutCtx, targetName, namespace)
			if err != nil {
				cLog.ErrorfCtx(timeoutCtx, "Error checking certificate status: %v", err)
				continue
			}

			if ready {
				cLog.InfofCtx(timeoutCtx, "Certificate %s is ready", targetName)
				return nil
			}

			cLog.InfofCtx(timeoutCtx, "Certificate %s not ready yet, waiting...", targetName)
		}
	}
}
