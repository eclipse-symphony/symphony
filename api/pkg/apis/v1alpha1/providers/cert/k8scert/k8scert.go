/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
				"commonName":  req.CommonName,
				"dnsNames":    req.DNSNames,
				"duration":    req.Duration.String(),
				"renewBefore": req.RenewBefore.String(),
				"issuerRef": map[string]interface{}{
					"name": req.IssuerName,
					"kind": "ClusterIssuer",
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
			return nil
		}
		sLog.ErrorfCtx(ctx, "  P (K8sCert): failed to create certificate: %+v", err)
		return err
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): created certificate %s in namespace %s", certName, req.Namespace)
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

// GetCert retrieves the certificate from the cert-manager created secret
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
		sLog.ErrorfCtx(ctx, "  P (K8sCert): certificate secret %s not found: %v", secretName, err)
		return nil, fmt.Errorf("certificate not found for target %s: %v", targetName, err)
	}

	certPEM := secret.Data["tls.crt"]
	keyPEM := secret.Data["tls.key"]

	if len(certPEM) == 0 || len(keyPEM) == 0 {
		sLog.ErrorfCtx(ctx, "  P (K8sCert): certificate secret %s is missing certificate or key data", secretName)
		return nil, fmt.Errorf("invalid certificate data for target %s", targetName)
	}

	response := &cert.CertResponse{
		PublicKey:    base64.StdEncoding.EncodeToString(certPEM),
		PrivateKey:   base64.StdEncoding.EncodeToString(keyPEM),
		ExpiresAt:    time.Now().Add(90 * 24 * time.Hour), // Default 90 days
		SerialNumber: "cert-manager-generated",
	}

	sLog.InfofCtx(ctx, "  P (K8sCert): retrieved certificate for target %s", targetName)
	return response, nil
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
		CommonName:  fmt.Sprintf("symphony-%s", targetName),
		DNSNames:    []string{targetName, fmt.Sprintf("%s.%s", targetName, namespace)},
		IssuerName:  "symphony-ca",
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
