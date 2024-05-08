/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
)

func createVisualizationVendor(t *testing.T) VisualizationVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := VisualizationVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name: "catalogs-manager",
				Type: "managers.symphony.catalogs",
				Properties: map[string]string{
					"providers.state": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"catalogs-manager": {
			"mem-state": &stateProvider,
		},
	}, &pubSubProvider)
	assert.Nil(t, err)
	return vendor
}

func TestVisualizationEndpoints(t *testing.T) {
	vendor := createVisualizationVendor(t)
	vendor.Route = "visualization"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestVisualizationInfo(t *testing.T) {
	vendor := createVisualizationVendor(t)
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestHandleVisPacket(t *testing.T) {
	vendor := createVisualizationVendor(t)
	vendor.Context.EvaluationContext = &utils.EvaluationContext{}
	packet := model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-1",
		Data:     []byte("data-1"),
		DataType: "bytes",
	}
	jData, _ := json.Marshal(packet)
	response := vendor.onVisPacket(v1alpha2.COARequest{
		Method:  "POST",
		Body:    jData,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, response.State)
	state, err := vendor.CatalogsManager.GetState(context.Background(), "solution-1-topology", "default")
	assert.Nil(t, err)
	assert.Equal(t, "solution-1-topology", state.ObjectMeta.Name)
	state, err = vendor.CatalogsManager.GetState(context.Background(), "target-1-topology", "default")
	assert.Nil(t, err)
	assert.Equal(t, "target-1-topology", state.ObjectMeta.Name)
	state, err = vendor.CatalogsManager.GetState(context.Background(), "instance-1-topology", "default")
	assert.Nil(t, err)
	assert.Equal(t, "instance-1-topology", state.ObjectMeta.Name)
}

func TestConvertVisualizationPacketToCatalog(t *testing.T) {
	catalog, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-1",
		Data:     []byte("data-1"),
		DataType: "bytes",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog.Spec.Type)

	v, ok := catalog.Spec.Properties["from-1"].(map[string]model.Packet)
	assert.True(t, ok)
	assert.Equal(t, "from-1", v["to-1"].From)
	assert.Equal(t, "to-1", v["to-1"].To)
	assert.Equal(t, "data-1", string(v["to-1"].Data.([]byte)))
	assert.Equal(t, "bytes", v["to-1"].DataType)
}
func TestConvertVisualizationPacketToCatalogNoData(t *testing.T) {
	catalog, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog.Spec.Type)

	v, ok := catalog.Spec.Properties["from-1"].(map[string]model.Packet)
	assert.True(t, ok)
	assert.Equal(t, "from-1", v["to-1"].From)
	assert.Equal(t, "to-1", v["to-1"].To)
}
func TestMergeCatalogsSameKey(t *testing.T) {
	catalog1, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-1",
		Data:     []byte("data-1"),
		DataType: "bytes",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog1.Spec.Type)
	catalog2, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-2",
		Data:     []byte("data-2"),
		DataType: "bytes",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog2.Spec.Type)

	mergedCatalog, err := mergeCatalogs(catalog1, catalog2)
	assert.Nil(t, err)
	v, ok := mergedCatalog.Spec.Properties["from-1"].(map[string]model.Packet)
	assert.True(t, ok)
	assert.Equal(t, "from-1", v["to-1"].From)
	assert.Equal(t, "to-1", v["to-1"].To)
	assert.Equal(t, "data-1", string(v["to-1"].Data.([]byte)))
	assert.Equal(t, "bytes", v["to-1"].DataType)
	assert.Equal(t, "from-1", v["to-2"].From)
	assert.Equal(t, "to-2", v["to-2"].To)
	assert.Equal(t, "data-2", string(v["to-2"].Data.([]byte)))
	assert.Equal(t, "bytes", v["to-2"].DataType)
}

func TestMergeCatalogsDifferentKey(t *testing.T) {
	catalog1, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-1",
		To:       "to-1",
		Data:     []byte("data-1"),
		DataType: "bytes",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog1.Spec.Type)
	catalog2, err := convertVisualizationPacketToCatalog("fake-site", model.Packet{
		Solution: "solution-1",
		Target:   "target-1",
		Instance: "instance-1",
		From:     "from-2",
		To:       "to-1",
		Data:     []byte("data-1"),
		DataType: "bytes",
	})
	assert.Nil(t, err)
	assert.Equal(t, "topology", catalog2.Spec.Type)

	mergedCatalog, err := mergeCatalogs(catalog1, catalog2)
	assert.Nil(t, err)
	v, ok := mergedCatalog.Spec.Properties["from-1"].(map[string]model.Packet)
	assert.True(t, ok)
	assert.Equal(t, "from-1", v["to-1"].From)
	assert.Equal(t, "to-1", v["to-1"].To)
	assert.Equal(t, "data-1", string(v["to-1"].Data.([]byte)))
	assert.Equal(t, "bytes", v["to-1"].DataType)
	v, ok = mergedCatalog.Spec.Properties["from-2"].(map[string]model.Packet)
	assert.True(t, ok)
	assert.Equal(t, "from-2", v["to-1"].From)
	assert.Equal(t, "to-1", v["to-1"].To)
	assert.Equal(t, "data-1", string(v["to-1"].Data.([]byte)))
	assert.Equal(t, "bytes", v["to-1"].DataType)
}
