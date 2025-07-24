//go:build azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package rust

import (
	"context"
	"errors"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	log = logger.NewLogger("providers.target.rust")
)

const (
	providerName = "P (Rust Target)"
)

type RustTargetProviderConfig struct {
	Name    string `json:"name"`
	LibFile string `json:"libFile"`
	LibHash string `json:"libHash"`
}

type RustTargetProvider struct {
	Context *contexts.ManagerContext
}

func RustTargetProviderConfigFromMap(properties map[string]string) (RustTargetProviderConfig, error) {
	return RustTargetProviderConfig{}, errors.New("Rust provider is not available in Azure")
}

func toRustTargetProviderConfig(config providers.IProviderConfig) (RustTargetProviderConfig, error) {
	return RustTargetProviderConfig{}, errors.New("Rust provider is not available in Azure")
}

func (i *RustTargetProvider) InitWithMap(properties map[string]string) error {
	return errors.New("Rust provider is not available in Azure")
}

func (s *RustTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (r *RustTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Rust Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfoCtx(ctx, "  P (Rust Target): Init()")

	return errors.New("Rust provider is not available in Azure")
}

func (r *RustTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{}
}

func (r *RustTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Rust Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (Rust Target Provider): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	return nil, errors.New("Rust provider is not available in Azure")
}

func (r *RustTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Rust Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfofCtx(ctx, "  P (Rust Target Provider): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	return nil, errors.New("Rust provider is not available in Azure")
}

func (r *RustTargetProvider) Close() {
	log.Info("  P (Rust Target): Close(), Rust provider is not available in Azure")
	return
}
