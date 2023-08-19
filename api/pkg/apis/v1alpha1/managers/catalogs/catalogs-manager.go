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

package catalogs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

type CatalogsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *CatalogsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (s *CatalogsManager) GetSpec(ctx context.Context, name string) (model.CatalogState, error) {
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
		},
	}
	entry, err := s.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CatalogState{}, err
	}

	ret, err := getCatalogState(name, entry.Body, entry.ETag)
	if err != nil {
		return model.CatalogState{}, err
	}
	return ret, nil
}

func getCatalogState(id string, body interface{}, etag string) (model.CatalogState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]
	status := dict["status"]
	j, _ := json.Marshal(spec)
	var rSpec model.CatalogSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.CatalogState{}, err
	}
	rSpec.Generation = etag
	j, _ = json.Marshal(status)
	var rStatus model.CatalogStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.CatalogState{}, err
	}
	state := model.CatalogState{
		Id:     id,
		Spec:   &rSpec,
		Status: &rStatus,
	}
	return state, nil
}

func (m *CatalogsManager) UpsertSpec(ctx context.Context, name string, spec model.CatalogSpec) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Catalog",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Catalog", "metadata": {"name": "$catalog()"}}`, model.FederationGroup),
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "catalogs",
		},
	}
	_, err := m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	m.Context.Publish("catalog", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": spec.Type,
		},
		Body: v1alpha2.JobData{
			Id:     spec.Name,
			Action: "UPDATE",
			Body:   spec,
		},
	})
	return nil
}

func (m *CatalogsManager) DeleteSpec(ctx context.Context, name string) error {
	//TODO: publish DELETE event
	return m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.FederationGroup,
			"version":  "v1",
			"resource": "catalogs",
		},
	})
}

func (t *CatalogsManager) ListSpec(ctx context.Context) ([]model.CatalogState, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.FederationGroup,
			"resource": "catalogs",
		},
	}
	catalogs, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CatalogState, 0)
	for _, t := range catalogs {
		rt, err := getCatalogState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
