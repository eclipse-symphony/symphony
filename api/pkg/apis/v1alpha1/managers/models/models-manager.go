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

package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

type ModelsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *ModelsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *ModelsManager) DeleteSpec(ctx context.Context, name string) error {
	return t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.AIGroup,
			"version":  "v1",
			"resource": "models",
		},
	})
}

func (t *ModelsManager) UpsertSpec(ctx context.Context, name string, spec model.DeviceSpec) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.AIGroup + "/v1",
				"kind":       "model",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion": "%s/v1", "kind": "Model", "metadata": {"name": "$model()"}}`, model.AIGroup),
			"scope":    "",
			"group":    model.AIGroup,
			"version":  "v1",
			"resource": "models",
		},
	}
	_, err := t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *ModelsManager) ListSpec(ctx context.Context) ([]model.ModelState, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "models",
		},
	}
	models, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.ModelState, 0)
	for _, t := range models {
		rt, err := getModelState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getModelState(id string, body interface{}, etag string) (model.ModelState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]

	j, _ := json.Marshal(spec)
	var rSpec model.ModelSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.ModelState{}, err
	}
	//rSpec.Generation??
	state := model.ModelState{
		Id:   id,
		Spec: &rSpec,
	}
	return state, nil
}

func (t *ModelsManager) GetSpec(ctx context.Context, id string) (model.ModelState, error) {
	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.AIGroup,
			"resource": "models",
		},
	}
	m, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.ModelState{}, err
	}

	ret, err := getModelState(id, m.Body, m.ETag)
	if err != nil {
		return model.ModelState{}, err
	}
	return ret, nil
}
