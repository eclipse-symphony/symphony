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

package sites

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

type SitesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *SitesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

// GetCampaign retrieves a CampaignSpec object by name
func (m *SitesManager) GetSpec(ctx context.Context, name string) (model.SiteState, error) {
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	entry, err := m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.SiteState{}, err
	}

	ret, err := getSiteState(name, entry.Body)
	if err != nil {
		return model.SiteState{}, err
	}
	return ret, nil
}

func getSiteState(id string, body interface{}) (model.SiteState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]
	status := dict["status"]

	j, _ := json.Marshal(spec)
	var rSpec model.SiteSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.SiteState{}, err
	}

	var rStatus model.SiteStatus

	if status != nil {
		j, _ = json.Marshal(status)
		err = json.Unmarshal(j, &rStatus)
		if err != nil {
			return model.SiteState{}, err
		}
	}
	state := model.SiteState{
		Id:     id,
		Spec:   &rSpec,
		Status: &rStatus,
	}
	return state, nil
}

func (t *SitesManager) ReportState(ctx context.Context, current model.SiteState) error {
	current.Metadata = map[string]string{
		"version":  "v1",
		"group":    model.FederationGroup,
		"resource": "sites",
	}
	getRequest := states.GetRequest{
		ID: current.Id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}

	entry, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			return err
		}
		err = t.UpsertSpec(ctx, current.Id, *current.Spec)
		if err != nil {
			return err
		}
		entry, err = t.StateProvider.Get(ctx, getRequest)
		if err != nil {
			return err
		}
	}

	// This copy is necessary becasue otherwise you could be modifying data in memory stage provider
	jTransfer, _ := json.Marshal(entry.Body)
	var dict map[string]interface{}
	json.Unmarshal(jTransfer, &dict)

	delete(dict, "spec")
	status := dict["status"]

	j, _ := json.Marshal(status)
	var rStatus model.SiteStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return err
	}
	rStatus.LastReported = time.Now().UTC().Format(time.RFC3339)
	dict["status"] = rStatus

	entry.Body = dict

	updateRequest := states.UpsertRequest{
		Value:    entry,
		Metadata: current.Metadata,
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) UpsertSpec(ctx context.Context, name string, spec model.SiteSpec) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Site",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Site", "metadata": {"name": "$site()"}}`, model.FederationGroup),
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "sites",
		},
	}
	_, err := m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *SitesManager) DeleteSpec(ctx context.Context, name string) error {
	return m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "sites",
		},
	})
}

func (t *SitesManager) ListSpec(ctx context.Context) ([]model.SiteState, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "sites",
		},
	}
	sites, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.SiteState, 0)
	for _, t := range sites {
		rt, err := getSiteState(t.ID, t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
func (s *SitesManager) Enabled() bool {
	return s.VendorContext.SiteInfo.ParentSite.BaseUrl != ""
}
func (s *SitesManager) Poll() []error {
	thisSite, err := s.GetSpec(context.Background(), s.VendorContext.SiteInfo.SiteId)
	if err != nil {
		//TOOD: only ignore not found, and log the error
		return nil
	}
	thisSite.Spec.IsSelf = false
	jData, _ := json.Marshal(thisSite)
	utils.UpdateSite(
		s.VendorContext.SiteInfo.ParentSite.BaseUrl,
		s.VendorContext.SiteInfo.SiteId,
		s.VendorContext.SiteInfo.ParentSite.Username,
		s.VendorContext.SiteInfo.ParentSite.Password,
		jData,
	)
	return nil
}
func (s *SitesManager) Reconcil() []error {
	return nil
}
