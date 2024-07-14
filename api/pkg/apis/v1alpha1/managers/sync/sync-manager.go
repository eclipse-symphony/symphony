/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sync

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type SyncManager struct {
	managers.Manager
	apiClient utils.ApiClient
}

func (s *SyncManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	if s.Context.SiteInfo.SiteId == "" {
		return v1alpha2.NewCOAError(nil, "siteId is required", v1alpha2.BadConfig)
	}
	s.apiClient, err = utils.GetParentApiClient(s.VendorContext.SiteInfo.ParentSite.BaseUrl)
	if err != nil {
		return err
	}
	return nil
}
func (s *SyncManager) Enabled() bool {
	return s.Config.Properties["sync.enabled"] == "true"
}
func (s *SyncManager) Poll() []error {
	ctx, span := observability.StartSpan("Sync Manager", context.Background(), &map[string]string{
		"method": "Poll",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	if s.VendorContext.SiteInfo.ParentSite.BaseUrl == "" {
		return nil
	}
	batch, err := s.apiClient.GetABatchForSite(ctx, s.VendorContext.SiteInfo.SiteId,
		s.VendorContext.SiteInfo.ParentSite.Username,
		s.VendorContext.SiteInfo.ParentSite.Password)
	if err != nil {
		return []error{err}
	}
	if batch.Catalogs != nil {
		for _, catalog := range batch.Catalogs {
			s.Context.Publish("catalog-sync", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": catalog.Spec.CatalogType,
					"origin":     batch.Origin,
				},
				Body: v1alpha2.JobData{
					Id:     catalog.ObjectMeta.Name,
					Action: v1alpha2.JobUpdate, //TODO: handle deletion, this probably requires BetBachForSites return flags
					Body:   catalog,
				},
				Context: ctx,
			})
		}
	}
	if batch.Jobs != nil {
		for _, job := range batch.Jobs {
			s.Context.Publish("remote-job", v1alpha2.Event{
				Metadata: map[string]string{
					"origin": batch.Origin,
				},
				Body:    job,
				Context: ctx,
			})
		}
	}
	if err != nil {
		return []error{err}
	}
	return nil
}
func (s *SyncManager) Reconcil() []error {
	return nil
}
