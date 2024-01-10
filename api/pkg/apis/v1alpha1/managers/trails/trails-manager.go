/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package trails

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type TrailsManager struct {
	managers.Manager
	LedgerProviders []ledger.ILedgerProvider
}

func (s *TrailsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	s.LedgerProviders = make([]ledger.ILedgerProvider, 0)
	for _, provider := range providers {
		if p, ok := provider.(ledger.ILedgerProvider); ok {
			s.LedgerProviders = append(s.LedgerProviders, p.(ledger.ILedgerProvider))
		}
	}
	return nil
}

func (s *TrailsManager) Append(ctx context.Context, trails []v1alpha2.Trail) error {
	ctx, span := observability.StartSpan("Sync Manager", ctx, &map[string]string{
		"method": "Append",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	errMessage := ""
	for _, p := range s.LedgerProviders {
		err = p.Append(ctx, trails)
		if err != nil {
			errMessage += err.Error() + ";"
		}
	}
	if errMessage != "" {
		err := v1alpha2.NewCOAError(nil, errMessage, v1alpha2.InternalError)
		return err
	}
	return nil
}
