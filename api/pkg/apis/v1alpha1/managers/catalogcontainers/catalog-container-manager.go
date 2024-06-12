/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogcontainers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

type CatalogContainersManager struct {
	managers.Manager
	StateProvider states.IStateProvider
}

func (s *CatalogContainersManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
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

func (t *CatalogContainersManager) DeleteState(ctx context.Context, name string, namespace string) error {
	ctx, span := observability.StartSpan("CatalogContainersManager", ctx, &map[string]string{
		"method": "DeleteState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	err = t.StateProvider.Delete(ctx, states.DeleteRequest{
		ID: name,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogcontainers",
			"kind":      "CatalogContainer",
		},
	})
	return err
}

func (t *CatalogContainersManager) UpsertState(ctx context.Context, name string, state model.CatalogContainerState) error {
	ctx, span := observability.StartSpan("CatalogContainersManager", ctx, &map[string]string{
		"method": "UpsertState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	if state.ObjectMeta.Name != "" && state.ObjectMeta.Name != name {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("Name in metadata (%s) does not match name in request (%s)", state.ObjectMeta.Name, name), v1alpha2.BadRequest)
	}
	state.ObjectMeta.FixNames(name)

	body := map[string]interface{}{
		"apiVersion": model.FederationGroup + "/v1",
		"kind":       "CatalogContainer",
		"metadata":   state.ObjectMeta,
		"spec":       state.Spec,
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   name,
			Body: body,
			ETag: "",
		},
		Metadata: map[string]interface{}{
			"namespace": state.ObjectMeta.Namespace,
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogcontainers",
			"kind":      "CatalogContainer",
		},
	}
	_, err = t.StateProvider.Upsert(ctx, upsertRequest)
	if err != nil {
		return err
	}
	return nil
}

func (t *CatalogContainersManager) ListState(ctx context.Context, namespace string) ([]model.CatalogContainerState, error) {
	ctx, span := observability.StartSpan("CatalogContainersManager", ctx, &map[string]string{
		"method": "ListState",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FederationGroup,
			"resource":  "catalogcontainers",
			"namespace": namespace,
			"kind":      "CatalogContainer",
		},
	}
	var catalogcontainers []states.StateEntry
	catalogcontainers, _, err = t.StateProvider.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	ret := make([]model.CatalogContainerState, 0)
	for _, t := range catalogcontainers {
		var rt model.CatalogContainerState
		rt, err = getCatalogContainerState(t.Body, t.ETag)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}
	return ret, nil
}

func getCatalogContainerState(body interface{}, etag string) (model.CatalogContainerState, error) {
	var CatalogContainerState model.CatalogContainerState
	bytes, _ := json.Marshal(body)
	err := json.Unmarshal(bytes, &CatalogContainerState)
	if err != nil {
		return model.CatalogContainerState{}, err
	}
	if CatalogContainerState.Spec == nil {
		CatalogContainerState.Spec = &model.CatalogContainerSpec{}
	}
	return CatalogContainerState, nil
}

func (t *CatalogContainersManager) GetState(ctx context.Context, id string, namespace string) (model.CatalogContainerState, error) {
	ctx, span := observability.StartSpan("CatalogContainersManager", ctx, &map[string]string{
		"method": "GetSpec",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	getRequest := states.GetRequest{
		ID: id,
		Metadata: map[string]interface{}{
			"version":   "v1",
			"group":     model.FederationGroup,
			"resource":  "catalogcontainers",
			"namespace": namespace,
			"kind":      "CatalogContainer",
		},
	}
	var Campaign states.StateEntry
	Campaign, err = t.StateProvider.Get(ctx, getRequest)
	if err != nil {
		return model.CatalogContainerState{}, err
	}
	var ret model.CatalogContainerState
	ret, err = getCatalogContainerState(Campaign.Body, Campaign.ETag)
	if err != nil {
		return model.CatalogContainerState{}, err
	}
	return ret, nil
}
