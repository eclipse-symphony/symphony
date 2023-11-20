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

package activations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	observability "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
)

var lock sync.Mutex

type ActivationsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *ActivationsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (m *ActivationsManager) GetSpec(ctx context.Context, name string) (model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.WorkflowGroup,
			"resource": "activations",
		},
	}
	entry, err := m.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.ActivationState{}, err
	}

	ret, err := getActivationState(name, entry.Body, entry.ETag)
	if err != nil {
		return model.ActivationState{}, err
	}
	return ret, nil
}

func getActivationState(id string, body interface{}, etag string) (model.ActivationState, error) {
	dict := body.(map[string]interface{})
	spec := dict["spec"]
	status := dict["status"]
	j, _ := json.Marshal(spec)
	var rSpec model.ActivationSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.ActivationState{}, err
	}
	j, _ = json.Marshal(status)
	var rStatus model.ActivationStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.ActivationState{}, err
	}
	rSpec.Generation = etag
	state := model.ActivationState{
		Id:     id,
		Spec:   &rSpec,
		Status: &rStatus,
	}
	return state, nil
}

func (m *ActivationsManager) UpsertSpec(ctx context.Context, name string, spec model.ActivationSpec) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata": map[string]interface{}{
					"name": name,
				},
				"spec": spec,
			},
			ETag: spec.Generation,
		},
		Metadata: map[string]string{
			"template": fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Activation", "metadata": {"name": "$activation()"}}`, model.WorkflowGroup),
			"scope":    "",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *ActivationsManager) DeleteSpec(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	return m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]string{
			"scope":    "",
			"group":    model.WorkflowGroup,
			"version":  "v1",
			"resource": "activations",
		},
	})
}

func (t *ActivationsManager) ListSpec(ctx context.Context) ([]model.ActivationState, error) {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	listRequest := states.ListRequest{
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.WorkflowGroup,
			"resource": "activations",
		},
	}
	solutions, _, err := t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.ActivationState, 0)
	for _, t := range solutions {
		rt, err := getActivationState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
func (t *ActivationsManager) ReportStatus(ctx context.Context, name string, current model.ActivationStatus) error {
	ctx, span := observability.StartSpan("Activations Manager", ctx, &map[string]string{
		"method": "ReportStatus",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	lock.Lock()
	defer lock.Unlock()
	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.WorkflowGroup,
			"resource": "activations",
		},
	}
	entry, err := t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return err
	}
	dict := entry.Body.(map[string]interface{})
	delete(dict, "spec")
	dict["status"] = current
	entry.Body = dict
	upsertRequest := states.UpsertRequest{
		Value: entry,
		Metadata: map[string]string{
			"version":  "v1",
			"group":    model.WorkflowGroup,
			"resource": "activations",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}
