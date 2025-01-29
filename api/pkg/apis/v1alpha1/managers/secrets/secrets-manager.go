/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package secrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type SecretsManager struct {
	managers.Manager
	SecretProviders map[string]secret.ISecretProvider
	Precedence      []string
}

func (s *SecretsManager) Init(context *contexts.VendorContext, cfg managers.ManagerConfig, providers map[string]providers.IProvider) error {
	log.Debug(" M (Secret): Init")
	err := s.Manager.Init(context, cfg, providers)
	if err != nil {
		return err
	}
	s.SecretProviders = make(map[string]secret.ISecretProvider)
	for key, provider := range providers {
		if cProvider, ok := provider.(secret.ISecretProvider); ok {
			s.SecretProviders[key] = cProvider
		}
	}
	if val, ok := cfg.Properties["precedence"]; ok {
		s.Precedence = strings.Split(val, ",")
	}
	if len(s.SecretProviders) == 0 {
		log.Error(" M (Secret): No secret providers found")
		return v1alpha2.NewCOAError(nil, "No secret providers found", v1alpha2.BadConfig)
	}
	if len(s.Precedence) < len(s.SecretProviders) && len(s.SecretProviders) > 1 {
		log.Error(" M (Secret): Not enough precedence values")
		return v1alpha2.NewCOAError(nil, "Not enough precedence values", v1alpha2.BadConfig)
	}
	if len(s.SecretProviders) > 1 {
		var provderKeys []string
		for key := range s.SecretProviders {
			provderKeys = append(provderKeys, key)
		}
		if !utils.AreSlicesEqual(provderKeys, s.Precedence) {
			log.Error(" M (Secret): Precedence does not match with secret providers")
			return v1alpha2.NewCOAError(nil, "Precedence does not match with secret providers", v1alpha2.BadConfig)
		}
	}
	return nil
}

func (s *SecretsManager) Get(ctx context.Context, object string, field string, localContext interface{}) (string, error) {
	ctx, span := observability.StartSpan("Secret Manager", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugfCtx(ctx, " M (Secret): Get %v, secret provider size %d", object, len(s.SecretProviders))
	if field == "" {
		log.ErrorfCtx(ctx, " M (Secret): field is empty")
		err = v1alpha2.NewCOAError(nil, "Field is empty", v1alpha2.BadRequest)
		return "", err
	}
	if strings.Index(object, "::") > 0 {
		parts := strings.Split(object, "::")
		if len(parts) != 2 {
			log.ErrorfCtx(ctx, " M (Secret): Invalid object: %s", object)
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid object: %s", object), v1alpha2.BadRequest)
			return "", err
		}
		if provider, ok := s.SecretProviders[parts[0]]; ok {
			return provider.Read(ctx, parts[1], field, localContext)
		}

		log.ErrorfCtx(ctx, " M (Secret): Invalid provider: %s", parts[0])
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Invalid provider: %s", parts[0]), v1alpha2.BadRequest)
		return "", err
	}

	if len(s.SecretProviders) == 1 {
		for _, provider := range s.SecretProviders {
			return provider.Read(ctx, object, field, localContext)
		}
	}
	for _, key := range s.Precedence {
		if provider, ok := s.SecretProviders[key]; ok {
			ret, err := provider.Read(ctx, object, field, localContext)
			if err == nil {
				return ret, nil

			}
		}
	}

	log.ErrorfCtx(ctx, " M (Secret): No provider found for object: %s", object)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("No provider found for object: %s", object), v1alpha2.NotFound)
	return "", err
}
