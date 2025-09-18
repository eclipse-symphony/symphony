/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package certmanager

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

// OSSCMCertProvider is a placeholder cert provider that implements the cert.ICertProvider interface
// This is used for backward compatibility in the provider factory
type OSSCMCertProvider struct {
	Config  providers.IProviderConfig
	Context *contexts.ManagerContext
}

func (o *OSSCMCertProvider) Init(config providers.IProviderConfig) error {
	o.Config = config
	return nil
}

func (o *OSSCMCertProvider) SetContext(ctx *contexts.ManagerContext) {
	o.Context = ctx
}

func (o *OSSCMCertProvider) CreateCert(ctx context.Context, req cert.CertRequest) error {
	return nil // placeholder implementation
}

func (o *OSSCMCertProvider) DeleteCert(ctx context.Context, targetName, namespace string) error {
	return nil // placeholder implementation
}

func (o *OSSCMCertProvider) GetCert(ctx context.Context, targetName, namespace string) (*cert.CertResponse, error) {
	return nil, nil // placeholder implementation
}

func (o *OSSCMCertProvider) RotateCert(ctx context.Context, targetName, namespace string) error {
	return nil // placeholder implementation
}

func (o *OSSCMCertProvider) CheckCertStatus(ctx context.Context, targetName, namespace string) (*cert.CertStatus, error) {
	return nil, nil // placeholder implementation
