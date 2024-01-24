/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/valyala/fasthttp"
)

type VisualizationVendor struct {
	vendors.Vendor
	CatalogsManager *catalogs.CatalogsManager
}

func (s *VisualizationVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  s.Vendor.Version,
		Name:     "Visualization",
		Producer: "Microsoft",
	}
}

func (e *VisualizationVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*catalogs.CatalogsManager); ok {
			e.CatalogsManager = c
		}
	}
	if e.CatalogsManager == nil {
		return v1alpha2.NewCOAError(nil, "catalogs manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *VisualizationVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "visualization"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodPost},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onVisPacket,
			Parameters: []string{},
		},
	}
}

func (c *VisualizationVendor) onVisPacket(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Visualization Vendor", request.Context, &map[string]string{
		"method": "onVisPacket",
	})
	defer span.End()

	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onVisPacket-POST", pCtx, nil)
		var packet model.Packet
		err := json.Unmarshal(request.Body, &packet)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}

		if !packet.IsValid() {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("invalid visualization packet"),
			})
		}

		catalog, err := convertVisualizationPacketToCatalog(packet)
		if err != nil {
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		if packet.Solution != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Solution), catalog)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Instance != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Instance), catalog)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Target != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Target), catalog)
			if err != nil {
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}

		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}

func (c *VisualizationVendor) updateSolutionTopologyCatalog(ctx context.Context, name string, catalog model.CatalogSpec) error {
	catalog.Name = name
	existingCatalog, err := c.CatalogsManager.GetSpec(ctx, name)
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			return err
		}
		return c.CatalogsManager.UpsertSpec(ctx, name, catalog)
	} else {
		catalog = mergeCatalogs(*existingCatalog.Spec, catalog)
		return c.CatalogsManager.UpsertSpec(ctx, name, catalog)
	}
}
func mergeCatalogs(existingCatalog, newCatalog model.CatalogSpec) model.CatalogSpec {
	mergedCatalog := existingCatalog
	for k, v := range newCatalog.Properties {
		mergedCatalog.Properties[k] = v
	}
	return mergedCatalog
}

func convertVisualizationPacketToCatalog(packet model.Packet) (model.CatalogSpec, error) {
	catalog := model.CatalogSpec{
		Properties: map[string]interface{}{
			packet.From: packet.To,
		},
	}
	return catalog, nil
}
