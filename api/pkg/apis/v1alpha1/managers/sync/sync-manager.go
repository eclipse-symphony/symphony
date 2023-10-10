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

package sync

import (
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type SyncManager struct {
	managers.Manager
}

func (s *SyncManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	if s.Context.SiteInfo.SiteId == "" {
		return v1alpha2.NewCOAError(nil, "siteId is required", v1alpha2.BadConfig)
	}
	return nil
}
func (s *SyncManager) Enabled() bool {
	return s.Config.Properties["sync.enabled"] == "true"
}
func (s *SyncManager) Poll() []error {
	if s.VendorContext.SiteInfo.ParentSite.BaseUrl == "" {
		return nil
	}
	batch, err := utils.GetABatchForSite(
		s.VendorContext.SiteInfo.ParentSite.BaseUrl,
		s.VendorContext.SiteInfo.SiteId,
		s.VendorContext.SiteInfo.ParentSite.Username,
		s.VendorContext.SiteInfo.ParentSite.Password)
	if err != nil {
		return []error{err}
	}
	if batch.Catalogs != nil {
		for _, catalog := range batch.Catalogs {
			s.Context.Publish("catalog-sync", v1alpha2.Event{
				Metadata: map[string]string{
					"objectType": catalog.Type,
				},
				Body: v1alpha2.JobData{
					Id:     catalog.Name,
					Action: "UPDATE", //TODO: handle deletion, this probably requires BetBachForSites return flags
					Body:   catalog,
				},
			})
		}
	}
	if batch.Jobs != nil {
		for _, job := range batch.Jobs {
			s.Context.Publish("remote-job", v1alpha2.Event{
				Metadata: map[string]string{
					"origin": batch.Origin,
				},
				Body: job,
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
