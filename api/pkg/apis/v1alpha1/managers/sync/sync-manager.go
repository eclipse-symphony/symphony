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
	"os"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type SyncManager struct {
	managers.Manager
	SiteId string
}

func (s *SyncManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	s.SiteId = os.Getenv("SYMPHONY_SITE_ID")
	if s.SiteId == "" {
		return v1alpha2.NewCOAError(nil, "siteId is required", v1alpha2.BadConfig)
	}
	return nil
}
func (s *SyncManager) Enabled() bool {
	return s.Config.Properties["sync.enabled"] == "true"
}
func (s *SyncManager) Poll() []error {
	baseUrl, err := utils.GetString(s.Manager.Config.Properties, "baseUrl")
	if err != nil {
		return []error{err}
	}
	user, err := utils.GetString(s.Manager.Config.Properties, "user")
	if err != nil {
		return []error{err}
	}
	password, err := utils.GetString(s.Manager.Config.Properties, "password")
	if err != nil {
		return []error{err}
	}
	batch, err := utils.GetABatchForSite(baseUrl, s.SiteId, user, password)
	if err != nil {
		return []error{err}
	}
	for _, catalog := range batch {
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
	if err != nil {
		return []error{err}
	}
	return nil
}
func (s *SyncManager) Reconcil() []error {
	return nil
}
