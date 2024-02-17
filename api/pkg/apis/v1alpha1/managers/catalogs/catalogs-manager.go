/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"

	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

var log = logger.NewLogger("coa.runtime")

type CatalogsManager struct {
	managers.Manager
	StateProvider states.IStateProvider
	GraphProvider graph.IGraphProvider
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
	for _, provider := range providers {
		if cProvider, ok := provider.(graph.IGraphProvider); ok {
			s.GraphProvider = cProvider
		}
	}
	return nil
}

func (s *CatalogsManager) GetState(ctx context.Context, name string) (model.CatalogState, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
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

	//read spec
	spec := dict["spec"]
	j, _ := json.Marshal(spec)
	var rSpec model.CatalogSpec
	err := json.Unmarshal(j, &rSpec)
	if err != nil {
		return model.CatalogState{}, err
	}
	rSpec.Generation = etag

	//read status
	status := dict["status"]
	j, _ = json.Marshal(status)
	var rStatus model.CatalogStatus
	err = json.Unmarshal(j, &rStatus)
	if err != nil {
		return model.CatalogState{}, err
	}

	//read metadata
	metadata := dict["metadata"]
	j, _ = json.Marshal(metadata)
	var rMetadata map[string]interface{}
	err = json.Unmarshal(j, &rMetadata)
	if err != nil {
		return model.CatalogState{}, err
	}

	state := model.CatalogState{
		Id:       id,
		Spec:     &rSpec,
		Status:   &rStatus,
		Metadata: rMetadata,
	}
	return state, nil
}
func (m *CatalogsManager) ValidateState(ctx context.Context, state model.CatalogState) (utils.SchemaResult, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "ValidateSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if schemaName, ok := state.Metadata["schema"]; ok {
		var schema model.CatalogState
		schema, err = m.GetState(ctx, schemaName.(string))
		if err != nil {
			err = v1alpha2.NewCOAError(err, "schema not found", v1alpha2.ValidateFailed)
			return utils.SchemaResult{Valid: false}, err
		}
		if s, ok := schema.Spec.Properties["spec"]; ok {
			var schemaObj utils.Schema
			jData, _ := json.Marshal(s)
			err = json.Unmarshal(jData, &schemaObj)
			if err != nil {
				err = v1alpha2.NewCOAError(err, "invalid schema", v1alpha2.ValidateFailed)
				return utils.SchemaResult{Valid: false}, err
			}
			return schemaObj.CheckProperties(state.Spec.Properties, nil)
		} else {
			err = v1alpha2.NewCOAError(fmt.Errorf("schema not found"), "schema validation error", v1alpha2.ValidateFailed)
			return utils.SchemaResult{Valid: false}, err
		}
	}
	return utils.SchemaResult{Valid: true}, nil
}
func (m *CatalogsManager) UpsertState(ctx context.Context, name string, state model.CatalogState) error {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "UpsertSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	metadata := map[string]interface{}{
		"name": name,
	}
	for k, v := range state.Metadata {
		metadata[k] = v
	}
	jMetadata, _ := json.Marshal(metadata)

	result, err := m.ValidateState(ctx, state)
	if err != nil {
		return err
	}
	if !result.Valid {
		err = v1alpha2.NewCOAError(nil, "schema validation error", v1alpha2.ValidateFailed)
		return err
	}
	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Catalog",
				"metadata":   metadata,
				"spec":       state.Spec,
			},
		},
		Metadata: map[string]interface{}{
			"template":  fmt.Sprintf(`{"apiVersion":"%s/v1", "kind": "Catalog", "metadata": %s}`, model.FederationGroup, string(jMetadata)),
			"namespace": "",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	m.Context.Publish("catalog", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": state.Spec.Type,
		},
		Body: v1alpha2.JobData{
			Id:     state.Spec.Name,
			Action: "UPDATE",
			Body:   state,
		},
	})
	return nil
}

func (m *CatalogsManager) DeleteState(ctx context.Context, name string) error {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "DeleteSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	//TODO: publish DELETE event
	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": "",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
		},
	})
	return err
}

func (t *CatalogsManager) ListState(ctx context.Context) ([]model.CatalogState, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "ListSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
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
		var rt model.CatalogState
		rt, err = getCatalogState(t.ID, t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}
func (g *CatalogsManager) setProviderDataIfNecessary(ctx context.Context) error {
	if !g.GraphProvider.IsPure() {
		catalogs, err := g.ListState(ctx)
		if err != nil {
			return err
		}
		data := make([]v1alpha2.INode, 0)
		for _, catalog := range catalogs {
			data = append(data, catalog)
		}
		err = g.GraphProvider.SetData(data)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *CatalogsManager) GetChains(ctx context.Context, filter string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetChains",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Debug(" M (Graph): GetChains")
	err = g.setProviderDataIfNecessary(ctx)
	if err != nil {
		return nil, err
	}
	ret, err := g.GraphProvider.GetChains(ctx, graph.ListRequest{Filter: filter})
	if err != nil {
		return nil, err
	}
	res := make(map[string][]v1alpha2.INode)
	for key, set := range ret.Sets {
		res[key] = set.Nodes
	}
	return res, nil
}
func (g *CatalogsManager) GetTrees(ctx context.Context, filter string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetTrees",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Debug(" M (Graph): GetTrees")
	err = g.setProviderDataIfNecessary(ctx)
	if err != nil {
		return nil, err
	}
	ret, err := g.GraphProvider.GetTrees(ctx, graph.ListRequest{Filter: filter})
	if err != nil {
		return nil, err
	}
	res := make(map[string][]v1alpha2.INode)
	for key, set := range ret.Sets {
		res[key] = set.Nodes
	}
	return res, nil
}
