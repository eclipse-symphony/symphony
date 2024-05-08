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
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var vcLog = logger.NewLogger("coa.runtime")

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
	vcLog.Info("V (Models): onVisPacket")

	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onVisPacket-POST", pCtx, nil)
		var packet model.Packet
		err := json.Unmarshal(request.Body, &packet)
		if err != nil {
			vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}

		if !packet.IsValid() {
			vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", "invalid visualization packet", span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("invalid visualization packet"),
			})
		}

		catalog, err := convertVisualizationPacketToCatalog(c.Context.SiteInfo.SiteId, packet)
		if err != nil {
			vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		if packet.Solution != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Solution), catalog)
			if err != nil {
				vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Instance != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Instance), catalog)
			if err != nil {
				vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.InternalError,
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Target != "" {
			err = c.updateSolutionTopologyCatalog(ctx, fmt.Sprintf("%s-topology", packet.Target), catalog)
			if err != nil {
				vcLog.Errorf("V (Visualization): onVisPacket failed - %s, traceId: %s", err.Error(), span.SpanContext().TraceID().String())
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

func (c *VisualizationVendor) updateSolutionTopologyCatalog(ctx context.Context, name string, catalog model.CatalogState) error {
	catalog.ObjectMeta.Name = name
	existingCatalog, err := c.CatalogsManager.GetState(ctx, name, catalog.ObjectMeta.Namespace)
	if err != nil {
		if !v1alpha2.IsNotFound(err) {
			return err
		}
		return c.CatalogsManager.UpsertState(ctx, name, catalog)
	} else {
		catalog, err = mergeCatalogs(existingCatalog, catalog)
		if err != nil {
			return err
		}
		return c.CatalogsManager.UpsertState(ctx, name, catalog)
	}
}
func mergeCatalogs(existingCatalog, newCatalog model.CatalogState) (model.CatalogState, error) {
	mergedCatalog := existingCatalog
	for k, v := range newCatalog.Spec.Properties {
		if ev, ok := existingCatalog.Spec.Properties[k]; ok {
			if vd, ok := v.(map[string]model.Packet); ok {
				if ed, ok := ev.(map[string]interface{}); ok {
					for ik, iv := range vd {
						ed[ik] = iv
					}
				} else if ed, ok := ev.(map[string]model.Packet); ok {
					for ik, iv := range vd {
						ed[ik] = iv
					}
				} else {
					return model.CatalogState{}, fmt.Errorf("cannot merge catalogs, existing property %s is not a map[string]interface{}", k)
				}
			} else {
				return model.CatalogState{}, fmt.Errorf("cannot merge catalogs, new property %s is not a map[string]model.Packet", k)
			}
		} else {
			mergedCatalog.Spec.Properties[k] = v
		}
	}
	return mergedCatalog, nil
}

func convertVisualizationPacketToCatalog(site string, packet model.Packet) (model.CatalogState, error) {
	catalog := model.CatalogState{
		Spec: &model.CatalogSpec{
			Type: "topology",
			Properties: map[string]interface{}{
				packet.From: map[string]model.Packet{
					packet.To: packet,
				},
			},
		},
	}
	return catalog, nil
}
