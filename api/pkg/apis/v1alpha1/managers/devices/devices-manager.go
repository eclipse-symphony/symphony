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

package devices

import (
	"context"
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

type DevicesManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

type DeviceState struct {
	Id   string            `json:"id"`
	Spec *model.DeviceSpec `json:"spec,omitempty"`
}

func (s *DevicesManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		s.StateProvider = stateprovider
	} else {
		return err
	}
	return nil
}

func (t *DevicesManager) DeleteSpec(ctx context.Context, name string) error {
	return t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    "fabric.symphony",
			"version":  "v1",
			"resource": "devices",
		},
	})
}

func (t *DevicesManager) UpsertSpec(ctx context.Context, name string, spec model.DeviceSpec) error {
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": "fabric.symphony/v1",
				"kind":       "device",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
		},
		Metadata: map[string]string{
			"template": `{"apiVersion":"fabric.symphony/v1", "kind": "Device", "metadata": {"name": "$device()"}}`,
			"scope":    "",
			"group":    "fabric.symphony",
			"version":  "v1",
			"resource": "devices",
		},
	}
	_, err := t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *DevicesManager) ListSpec(ctx context.Context) ([]DeviceState, error) {
	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "fabric.symphony",
			"resource": "devices",
		},
	}
	solutions, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]DeviceState, 0)
	for _, t := range solutions {
		rt, err := getDeviceState(t.ID, t.Body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getDeviceState(id string, body interface{}) (DeviceState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]

	j, _ := json.Marshal(spec)
	var rSpec model.DeviceSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return DeviceState{}, err
	}
	state := DeviceState{
		Id:   id,
		Spec: &rSpec,
	}
	return state, nil
}

func (t *DevicesManager) GetSpec(ctx context.Context, id string) (DeviceState, error) {
	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    "fabric.symphony",
			"resource": "devices",
		},
	}
	target, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return DeviceState{}, err
	}

	ret, err := getDeviceState(id, target.Body)
	if err != nil {
		return DeviceState{}, err
	}
	return ret, nil
}
