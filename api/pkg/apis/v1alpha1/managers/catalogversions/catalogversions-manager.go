/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogversions

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

type CatalogVersionsManager struct {
	managers.Manager
	StateProvider    states.IStateProvider
	GraphProvider    graph.IGraphProvider
	needValidate     bool
	CatalogVersionValidator validation.CatalogVersionValidator
}

func (s *CatalogVersionsManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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
		// s.CatalogVersionValidator = validation.NewCatalogVersionValidator(s.CatalogVersionLookup, s.CatalogLookup, s.ChildCatalogVersionLookup)
		s.CatalogVersionValidator = validation.NewCatalogVersionValidator(s.CatalogVersionLookup, nil, s.ChildCatalogVersionLookup)
	}
	return nil
}

func (s *CatalogVersionsManager) GetState(ctx context.Context, name string, namespace string) (model.CatalogVersionState, error) {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
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
			"resource":  "catalogversions",
			"namespace": namespace,
			"kind":      "CatalogVersion",
		},
	}
	var entry states.StateEntry
	entry, err = s.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CatalogVersionState{}, err
	}
	var ret model.CatalogVersionState
	ret, err = getCatalogVersionState(entry.Body)
	if err != nil {
		return model.CatalogVersionState{}, err
	}
	ret.ObjectMeta.UpdateEtag(entry.ETag)
	return ret, nil
}

func getCatalogVersionState(body interface{}) (model.CatalogVersionState, error) {
	var catalogversionState model.CatalogVersionState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &catalogversionState)
	if err != nil {
		return model.CatalogVersionState{}, err
	}
	if catalogversionState.Spec == nil {
		catalogversionState.Spec = &model.CatalogVersionSpec{}
	}
	if catalogversionState.Status == nil {
		catalogversionState.Status = &model.CatalogVersionStatus{}
	}
	return catalogversionState, nil
}

func (m *CatalogVersionsManager) UpsertState(ctx context.Context, name string, state model.CatalogVersionState) error {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
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
		if err = validation.ValidateCreateOrUpdateWrapper(ctx, &m.CatalogVersionValidator, state, oldState, getStateErr); err != nil {
			return err
		}
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID: name,
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "CatalogVersion",
				"metadata":   state.ObjectMeta,
				"spec":       state.Spec,
			},
			ETag: state.ObjectMeta.ETag,
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogversions",
			"kind":      "CatalogVersion",
		},
	}
	_, err = m.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	m.Context.Publish("catalogversion", v1alpha2.Event{
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

func (m *CatalogVersionsManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
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
			"resource":  "catalogversions",
			"kind":      "CatalogVersion",
		},
	})
	return err
}

func (t *CatalogVersionsManager) ListState(ctx context.Context, namespace string, filterType string, filterValue string) ([]model.CatalogVersionState, error) {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FederationGroup,
			"resource":  "catalogversions",
			"namespace": namespace,
			"kind":      "CatalogVersion",
		},
	}
	listRequest.FilterType = filterType
	listRequest.FilterValue = filterValue
	var catalogversions []states.StateEntry
	catalogversions, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CatalogVersionState, 0)
	for _, t := range catalogversions {
		var rt model.CatalogVersionState
		rt, err = getCatalogVersionState(t.Body)
		if err != nil {
			return nil, err
		}
		rt.ObjectMeta.UpdateEtag(t.ETag)
		ret = append(ret, rt)
	}
	return ret, nil
}
func (g *CatalogVersionsManager) setProviderDataIfNecessary(ctx context.Context, namespace string) error {
	if !g.GraphProvider.IsPure() {
		catalogversions, err := g.ListState(ctx, namespace, "", "")
		if err != nil {
			return err
		}
		data := make([]v1alpha2.INode, 0)
		for _, catalogversion := range catalogversions {
			data = append(data, catalogversion)
		}
		err = g.GraphProvider.SetData(data)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *CatalogVersionsManager) GetChains(ctx context.Context, filter string, namespace string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
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
func (g *CatalogVersionsManager) GetTrees(ctx context.Context, filter string, namespace string) (map[string][]v1alpha2.INode, error) {
	ctx, span := observability.StartSpan("CatalogVersions Manager", ctx, &map[string]string{
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

func (t *CatalogVersionsManager) ValidateDelete(ctx context.Context, name string, namespace string) error {
	state, err := t.GetState(ctx, name, namespace)
	return validation.ValidateDeleteWrapper(ctx, &t.CatalogVersionValidator, state, err)
}

func (t *CatalogVersionsManager) CatalogLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.Catalog, name, namespace)
}

func (t *CatalogVersionsManager) CatalogVersionLookup(ctx context.Context, name string, namespace string) (interface{}, error) {
	return states.GetObjectState(ctx, t.StateProvider, validation.CatalogVersion, name, namespace)
}

func (t *CatalogVersionsManager) ChildCatalogVersionLookup(ctx context.Context, name string, namespace string, uid string) (bool, error) {
	catalogversionList, err := states.ListObjectStateWithLabels(ctx, t.StateProvider, validation.CatalogVersion, namespace, map[string]string{constants.ParentName: name}, 1)
	if err != nil {
		return false, err
	}
	return len(catalogversionList) > 0, nil
}
