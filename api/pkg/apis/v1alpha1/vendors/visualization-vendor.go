/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogversions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var vcLog = logger.NewLogger("coa.runtime")

type VisualizationVendor struct {
	vendors.Vendor
	CatalogVersionsManager *catalogversions.CatalogVersionsManager
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
		if c, ok := m.(*catalogversions.CatalogVersionsManager); ok {
			e.CatalogVersionsManager = c
		}
	}
	if e.CatalogVersionsManager == nil {
		return v1alpha2.NewCOAError(nil, "catalogversions manager is not supplied", v1alpha2.MissingConfig)
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
	vcLog.InfoCtx(pCtx, "V (Models): onVisPacket")

	switch request.Method {
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onVisPacket-POST", pCtx, nil)
		var packet model.Packet
		err := utils2.UnmarshalJson(request.Body, &packet)
		if err != nil {
			vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte(err.Error()),
			})
		}

		if !packet.IsValid() {
			vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", "invalid visualization packet")
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.BadRequest,
				Body:  []byte("invalid visualization packet"),
			})
		}

		catalogversion, err := convertVisualizationPacketToCatalogVersion(packet)
		if err != nil {
			vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", err.Error())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		if packet.SolutionVersion != "" {
			err = c.updateSolutionVersionTopologyCatalogVersion(ctx, fmt.Sprintf("%s-topology", packet.SolutionVersion), catalogversion)
			if err != nil {
				vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Instance != "" {
			err = c.updateSolutionVersionTopologyCatalogVersion(ctx, fmt.Sprintf("%s-topology", packet.Instance), catalogversion)
			if err != nil {
				vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
					Body:  []byte(err.Error()),
				})
			}
		}

		if packet.Target != "" {
			err = c.updateSolutionVersionTopologyCatalogVersion(ctx, fmt.Sprintf("%s-topology", packet.Target), catalogversion)
			if err != nil {
				vcLog.ErrorfCtx(pCtx, "V (Visualization): onVisPacket failed - %s", err.Error())
				return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
					State: v1alpha2.GetErrorState(err),
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

func (c *VisualizationVendor) updateSolutionVersionTopologyCatalogVersion(ctx context.Context, name string, catalogversion model.CatalogVersionState) error {
	catalogversion.ObjectMeta.Name = name + "-v-version1"
	catalogversion.Spec.RootResource = validation.GetRootResourceFromName(catalogversion.ObjectMeta.Name)
	existingCatalogVersion, err := c.CatalogVersionsManager.GetState(ctx, name, catalogversion.ObjectMeta.Namespace)
	if err != nil {
		if !utils.IsNotFound(err) {
			return err
		}
		return c.CatalogVersionsManager.UpsertState(ctx, catalogversion.ObjectMeta.Name, catalogversion)
	} else {
		catalogversion, err = mergeCatalogVersions(existingCatalogVersion, catalogversion)
		if err != nil {
			return err
		}
		return c.CatalogVersionsManager.UpsertState(ctx, catalogversion.ObjectMeta.Name, catalogversion)
	}
}
func mergeCatalogVersions(existingCatalogVersion, newCatalogVersion model.CatalogVersionState) (model.CatalogVersionState, error) {
	mergedCatalogVersion := existingCatalogVersion
	for k, v := range newCatalogVersion.Spec.Properties {
		if ev, ok := existingCatalogVersion.Spec.Properties[k]; ok {
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
					return model.CatalogVersionState{}, fmt.Errorf("cannot merge catalogversions, existing property %s is not a map[string]interface{}", k)
				}
			} else {
				return model.CatalogVersionState{}, fmt.Errorf("cannot merge catalogversions, new property %s is not a map[string]model.Packet", k)
			}
		} else {
			mergedCatalogVersion.Spec.Properties[k] = v
		}
	}
	return mergedCatalogVersion, nil
}

func convertVisualizationPacketToCatalogVersion(packet model.Packet) (model.CatalogVersionState, error) {
	catalogversion := model.CatalogVersionState{
		Spec: &model.CatalogVersionSpec{
			CatalogType: "topology",
			Properties: map[string]interface{}{
				packet.From: map[string]model.Packet{
					packet.To: packet,
				},
			},
		},
	}
	return catalogversion, nil
}
