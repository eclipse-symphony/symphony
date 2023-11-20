/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package trails

import (
	"context"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/ledger"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
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
	defer observ_utils.CloseSpanWithError(span, err)

	errMessage := ""
	for _, p := range s.LedgerProviders {
		err := p.Append(ctx, trails)
		if err != nil {
			errMessage += err.Error() + ";"
		}
	}
	if errMessage != "" {
		retError := v1alpha2.NewCOAError(nil, errMessage, v1alpha2.InternalError)
		return retError
	}
	return nil
}
