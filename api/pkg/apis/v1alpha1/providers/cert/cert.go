/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cert

import (
	"context"
	"time"
)

// ICertProvider defines the interface for certificate management
type ICertProvider interface {
	// CreateCert creates a certificate for the specified target
	CreateCert(ctx context.Context, req CertRequest) error

	// DeleteCert deletes the certificate for the specified target
	DeleteCert(ctx context.Context, targetName, namespace string) error

	// GetCert retrieves the certificate for the specified target (read-only)
	GetCert(ctx context.Context, targetName, namespace string) (*CertResponse, error)

	// RotateCert rotates/renews the certificate for the specified target
	RotateCert(ctx context.Context, targetName, namespace string) error

	// CheckCertStatus checks if the certificate is ready and valid
	CheckCertStatus(ctx context.Context, targetName, namespace string) (*CertStatus, error)
}

// CertRequest represents a certificate creation request
type CertRequest struct {
	TargetName  string        `json:"targetName"`
	Namespace   string        `json:"namespace"`
	Duration    time.Duration `json:"duration"`
	RenewBefore time.Duration `json:"renewBefore"`
	CommonName  string        `json:"commonName"`
	DNSNames    []string      `json:"dnsNames"`
	IssuerName  string        `json:"issuerName"`
	ServiceName string        `json:"serviceName"`
}

// CertResponse represents the certificate data
type CertResponse struct {
	PublicKey    string    `json:"publicKey"`
	PrivateKey   string    `json:"privateKey"`
	ExpiresAt    time.Time `json:"expiresAt"`
	SerialNumber string    `json:"serialNumber"`
}

// CertStatus represents the certificate status
type CertStatus struct {
	Ready       bool      `json:"ready"`
	Reason      string    `json:"reason"`
	Message     string    `json:"message"`
	LastUpdate  time.Time `json:"lastUpdate"`
	NextRenewal time.Time `json:"nextRenewal"`
}
