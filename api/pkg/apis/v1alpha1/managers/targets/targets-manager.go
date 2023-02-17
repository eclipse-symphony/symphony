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

package targets

import (
	"context"
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/registry"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

type TargetsManager struct {
	managers.Manager
	StateProvider    states.IStateProvider
	RegistryProvider registry.IRegistryProvider
}

type TargetState struct {
	Id       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Status   map[string]string `json:"status,omitempty"`
	Spec     *model.TargetSpec `json:"spec,omitempty"`
}

func (s *TargetsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}

	return nil
}

func (t *TargetsManager) DeleteSpec(ctx context.Context, name string) error {
	return t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    "fabric.symphony",
			"version":  "v1",
			"resource": "targets",
		},
	})
}

func (t *TargetsManager) UpsertSpec(ctx context.Context, name string, spec model.TargetSpec) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": "fabric.symphony/v1",
				"kind":       "Target",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": `{"apiVersion":"fabric.symphony/v1", "kind": "Target", "metadata": {"name": "$target()"}}`,
			"scope":    "",
			"group":    "fabric.symphony",
			"version":  "v1",
			"resource": "targets",
		},
	}
	_, err := t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *TargetsManager) ReportState(ctx context.Context, current TargetState) (TargetState, error) {
	getRequest := states.GetRequest{
		ID:       current.Id,
		Metadata: current.Metadata,
	}
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return TargetState{}, err
	}

	dict := target.Body.(map[string]interface{})

	specCol := dict["spec"].(map[string]interface{})
	var metadata map[string]string
	if m, ok := specCol["metadata"]; ok {
		jm, _ := json.Marshal(m)
		json.Unmarshal(jm, &metadata)
	}

	delete(dict, "spec")
	status := dict["status"]

	j, _ := json.Marshal(status)
	var rStatus map[string]interface{}
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return TargetState{}, err
	}
	j, _ = json.Marshal(rStatus["properties"])
	var rProperties map[string]string
	err = json.Unmarshal(j, &rProperties)
	if err != nil {
		return TargetState{}, err
	}

	for k, v := range current.Status {
		rProperties[k] = v
	}

	dict["status"].(map[string]interface{})["properties"] = rProperties

	target.Body = dict

	updateRequest := states.UpsertRequest{
		Value:    target,
		Metadata: current.Metadata,
	}

	_, err = t.StateProvider.Upsert(ctx, updateRequest)
	if err != nil {
		return TargetState{}, err
	}

	return TargetState{
		Id:       current.Id,
		Metadata: metadata,
		Status:   rProperties,
	}, nil
}
func (t *TargetsManager) ListSpec(ctx context.Context) ([]TargetState, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "fabric.symphony",
			"resource": "targets",
		},
	}
	targets, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]TargetState, 0)
	for _, t := range targets {
		rt, err := getTargetState(t.ID, t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getTargetState(id string, body interface{}) (TargetState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]
	status := dict["status"]

	j, _ := json.Marshal(spec)
	var rSpec model.TargetSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return TargetState{}, err
	}

	j, _ = json.Marshal(status)
	var rStatus map[string]interface{}
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return TargetState{}, err
	}
	j, _ = json.Marshal(rStatus["properties"])
	var rProperties map[string]string
	err = json.Unmarshal(j, &rProperties)
	if err != nil {
		return TargetState{}, err
	}

	state := TargetState{
		Id:     id,
		Spec:   &rSpec,
		Status: rProperties,
	}
	return state, nil
}

func (t *TargetsManager) GetSpec(ctx context.Context, id string) (TargetState, error) {
	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "fabric.symphony",
			"resource": "targets",
		},
	}
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return TargetState{}, err
	}

	ret, err := getTargetState(id, target.Body)
	if err != nil {
		return TargetState{}, err
	}
	return ret, nil
}
