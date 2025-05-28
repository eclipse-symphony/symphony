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

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
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
	StateProvider    states.IStateProvider
	GraphProvider    graph.IGraphProvider
	needValidate     bool
	CatalogValidator validation.CatalogValidator
}

func (s *CatalogsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	err := s.Manager.Init(context, config, providers)
	if err != nil {
		return err
	}
	stateprovider, err := managers.GetPersistentStateProvider(config, providers)
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
	s.needValidate = managers.NeedObjectValidate(config, providers)
	if s.needValidate {
		// Turn off validation of differnt types: https://github.com/eclipse-symphony/symphony/issues/445
		// s.CatalogValidator = validation.NewCatalogValidator(s.CatalogLookup, s.CatalogContainerLookup, s.ChildCatalogLookup)
		s.CatalogValidator = validation.NewCatalogValidator(s.CatalogLookup, nil, s.ChildCatalogLookup)
	}
	return nil
}

func (s *CatalogsManager) GetState(ctx context.Context, name string, namespace string) (model.CatalogState, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	getRequest := states.GetRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FederationGroup,
			"resource":  "catalogs",
			"namespace": namespace,
			"kind":      "Catalog",
		},
	}
	var entry states.StateEntry
	entry, err = s.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CatalogState{}, err
	}
	var ret model.CatalogState
	ret, err = getCatalogState(entry.Body)
	if err != nil {
		return model.CatalogState{}, err
	}
	ret.ObjectMeta.UpdateEtag(entry.ETag)
	return ret, nil
}

func getCatalogState(body interface{}) (model.CatalogState, error) {
	var catalogState model.CatalogState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &catalogState)
	if err != nil {
		return model.CatalogState{}, err
	}
	if catalogState.Spec == nil {
		catalogState.Spec = &model.CatalogSpec{}
	}
	if catalogState.Status == nil {
		catalogState.Status = &model.CatalogStatus{}
	}
	return catalogState, nil
}

func (m *CatalogsManager) UpsertState(ctx context.Context, name string, state model.CatalogState) error {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	oldState, getStateErr := m.GetState(ctx, state.ObjectMeta.Name, state.ObjectMeta.Namespace)
	if getStateErr == nil {
		state.ObjectMeta.PreserveSystemMetadata(oldState.ObjectMeta)
	}

	if m.needValidate {
		if state.ObjectMeta.Labels == nil {
			state.ObjectMeta.Labels = make(map[string]string)
		}
		if state.Spec != nil {
			state.ObjectMeta.Labels[constants.RootResource] = state.Spec.RootResource
			if state.Spec.ParentName != "" {
				state.ObjectMeta.Labels[constants.ParentName] = validation.ConvertReferenceToObjectName(state.Spec.ParentName)
			}
		}
		if err = validation.ValidateCreateOrUpdateWrapper(ctx, &m.CatalogValidator, state, oldState, getStateErr); err != nil {
			return err
		}
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "Catalog",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
			ETag: state.ObjectMeta.ETag,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	m.Context.Publish("catalog", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": state.Spec.CatalogType,
		},
		Body: v1alpha2.JobData{
			Id:     state.ObjectMeta.Name,
			Action: v1alpha2.JobUpdate,
			Body:   state,
		},
		Context: ctx,
	})
	return nil
}

func (m *CatalogsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if m.needValidate {
		if err = m.ValidateDelete(ctx, name, namespace); err != nil {
			return err
		}
	}

	//TODO: publish DELETE event
	err = m.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogs",
			"kind":      "Catalog",
		},
	})
	return err
}

func (t *CatalogsManager) ListState(ctx context.Context, namespace string, filterType string, filterValue string) ([]model.CatalogState, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FederationGroup,
			"resource":  "catalogs",
			"namespace": namespace,
			"kind":      "Catalog",
		},
	}
	listRequest.FilterType = filterType
	listRequest.FilterValue = filterValue
	var catalogs []states.StateEntry
	catalogs, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CatalogState, 0)
	for _, t := range catalogs {
		var rt model.CatalogState
		rt, err = getCatalogState(t.Body)
		if err != nil {
			return nil, err
		}
		rt.ObjectMeta.UpdateEtag(t.ETag)
		ret = append(ret, rt)
	}
	return ret, nil
}
func (g *CatalogsManager) setProviderDataIfNecessary(ctx context.Context, namespace string) error {
	if !g.GraphProvider.IsPure() {
		catalogs, err := g.ListState(ctx, namespace, "", "")
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
func (g *CatalogsManager) GetChains(ctx context.Context, filter string, namespace string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetChains",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugCtx(ctx, " M (Graph): GetChains")
	err = g.setProviderDataIfNecessary(ctx, namespace)
	if err != nil {
		return nil, err
	}
	var ret graph.GetSetsResponse
	ret, err = g.GraphProvider.GetChains(ctx, graph.ListRequest{Filter: filter})
	if err != nil {
		return nil, err
	}
	res := make(map[string][]v1alpha2.INode)
	for key, set := range ret.Sets {
		res[key] = set.Nodes
	}
	return res, nil
}
func (g *CatalogsManager) GetTrees(ctx context.Context, filter string, namespace string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("Catalogs Manager", ctx, &map[string]string{
		"method": "GetTrees",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.DebugCtx(ctx, " M (Graph): GetTrees")
	err = g.setProviderDataIfNecessary(ctx, namespace)
	if err != nil {
		return nil, err
	}
	var ret graph.GetSetsResponse
	ret, err = g.GraphProvider.GetTrees(ctx, graph.ListRequest{Filter: filter})
	if err != nil {
		return nil, err
	}
	res := make(map[string][]v1alpha2.INode)
	for key, set := range ret.Sets {
		res[key] = set.Nodes
	}
	return res, nil
}

func (t *CatalogsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.CatalogValidator, state, err)
}

func (t *CatalogsManager) CatalogContainerLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.CatalogContainer, name, namespace)
}

func (t *CatalogsManager) CatalogLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.Catalog, name, namespace)
}

func (t *CatalogsManager) ChildCatalogLookup(ctx context.Context, name string, namespace string, uid string) (bool, error) {
	catalogList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.Catalog, namespace, map[string]string{constants.ParentName: name}, 1)
	if err != nil {
		return false, err
	}
	return len(catalogList) > 0, nil
}
