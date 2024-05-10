package vendors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	memorygraph "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	mockledger "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	memoryqueue "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var SiteSpec = model.SiteSpec{
	Name:      "child1",
	IsSelf:    false,
	PublicKey: "examplePublicKey",
	Properties: map[string]string{
		"property1": "value1",
		"property2": "value2",
	},
}

func federationVendorInit() FederationVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	graphProvider := &memorygraph.MemoryGraphProvider{}
	graphProvider.Init(memorygraph.MemoryGraphProviderConfig{})

	catalogProviders := make(map[string]providers.IProvider)
	catalogProviders["StateProvider"] = stateProvider
	catalogProviders["GraphProvider"] = graphProvider

	queueProvider := &memoryqueue.MemoryQueueProvider{}
	queueProvider.Init(memoryqueue.MemoryQueueProvider{})
	stagingProviders := make(map[string]providers.IProvider)
	stagingProviders["StateProvider"] = stateProvider
	stagingProviders["QueueProvider"] = queueProvider

	siteProviders := make(map[string]providers.IProvider)
	siteProviders["StateProvider"] = stateProvider

	mProvider := &mockledger.MockLedgerProvider{}
	mProvider.Init(mockledger.MockLedgerProviderConfig{})
	trailsProviders := make(map[string]providers.IProvider)
	trailsProviders["LedgerProvider"] = mProvider

	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := FederationVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "exampleSiteId",
			Properties: map[string]string{
				"property1": "value1",
				"property2": "value2",
			},
		},
		Route: "federation",
		Managers: []managers.ManagerConfig{
			{
				Name: "catalog-manager",
				Type: "managers.symphony.catalogs",
				Properties: map[string]string{
					"providers.state": "StateProvider",
				},
			},
			{
				Name: "staging-manager",
				Type: "managers.symphony.staging",
				Properties: map[string]string{
					"providers.state": "StateProvider",
					"providers.queue": "QueueProvider",
				},
			},
			{
				Name: "sites-manager",
				Type: "managers.symphony.sites",
				Properties: map[string]string{
					"providers.state": "StateProvider",
				},
			},
			{
				Name: "sync-manager",
				Type: "managers.symphony.sync",
			},
			{
				Name: "trails-manager",
				Type: "managers.symphony.trails",
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"catalog-manager": catalogProviders,
		"trails-manager":  trailsProviders,
		"sites-manager":   siteProviders,
		"staging-manager": stagingProviders,
	}, &pubSubProvider)
	return vendor
}

func TestFederationGetEndpoint(t *testing.T) {
	vendor := federationVendorInit()
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "federation/k8shook", endpoints[len(endpoints)-1].Route)
}

func TestFederationGetInfo(t *testing.T) {
	vendor := federationVendorInit()
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "Federation", info.Name)
}

func TestFederationOnRegister(t *testing.T) {
	vendor := federationVendorInit()

	SiteSpec.Name = "test1"
	b, err := json.Marshal(SiteSpec)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
		Body: b,
	}

	response := vendor.onRegistry(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	response = vendor.onRegistry(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summarys []model.SiteState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(summarys))
	assert.Equal(t, true, summarys[0].Id == SiteSpec.Name || summarys[1].Id == SiteSpec.Name)

	requestGet.Parameters = map[string]string{
		"__name": SiteSpec.Name,
	}
	response = vendor.onRegistry(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summary model.SiteState
	err = json.Unmarshal(response.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, SiteSpec.Name, summary.Id)

	requestDelete := &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
	}
	response = vendor.onRegistry(*requestDelete)
	assert.Equal(t, v1alpha2.OK, response.State)

	response = vendor.onRegistry(*requestGet)
	assert.Equal(t, v1alpha2.NotFound, response.State)

	requestPatch := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
	}
	response = vendor.onRegistry(*requestPatch)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)

}

func TestFederationOnStatus(t *testing.T) {
	vendor := federationVendorInit()

	SiteSpec.Name = "test1"
	var state model.SiteState
	state.Id = SiteSpec.Name
	state.Spec = &model.SiteSpec{}
	var status model.SiteStatus
	status.IsOnline = true
	state.Status = &status
	b, err := json.Marshal(state)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
		Body: b,
	}

	response := vendor.onStatus(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestPatch := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
		Body: b,
	}
	response = vendor.onStatus(*requestPatch)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestFederationOnSyncPost(t *testing.T) {
	vendor := federationVendorInit()

	activationStatus := model.ActivationStatus{
		Stage:     "exampleStage",
		NextStage: "exampleNextStage",
		Inputs: map[string]interface{}{
			"input1": "value1",
			"input2": "value2",
		},
		Outputs: map[string]interface{}{
			"output1": "value1",
			"output2": "value2",
		},
		Status:               v1alpha2.OK,
		IsActive:             true,
		ActivationGeneration: "1",
		UpdateTime:           "exampleUpdateTime",
	}
	// vendor.Context.PubsubProvider.Publish("report", v1alpha2.Event{
	// 	Metadata: map[string]string{
	// 		"objectType": "instance",
	// 	},
	// 	Body: v1alpha2.JobData{
	// 		Id:     "testInstanceId",
	// 		Action: "UPDATE",
	// 	},
	// })

	// vendor.Context.PubsubProvider.Publish("report", v1alpha2.Event{
	// 	Metadata: map[string]string{
	// 		"objectType": "instance",
	// 	},
	// 	Body: activationStatus,
	// })

	b, err := json.Marshal(activationStatus)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    b,
	}
	response := vendor.onSync(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)
	vendor.Context.PubsubProvider.Subscribe("job-report", func(topic string, event v1alpha2.Event) error {
		jData, _ := json.Marshal(event.Body)
		var status model.ActivationStatus
		err := json.Unmarshal(jData, &status)
		assert.Nil(t, err)
		assert.Equal(t, activationStatus.Stage, status.Stage)
		return nil
	})

	requestPatch := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
		Body: b,
	}
	response = vendor.onSync(*requestPatch)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestFederationOnSyncGet(t *testing.T) {
	vendor := federationVendorInit()

	SiteSpec.Name = "test1"
	b, err := json.Marshal(SiteSpec)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": SiteSpec.Name,
		},
		Body: b,
	}

	response := vendor.onRegistry(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	vendor.Context.PubsubProvider.Publish("remote", v1alpha2.Event{
		Metadata: map[string]string{
			"site": SiteSpec.Name,
		},
		Body: v1alpha2.JobData{
			Id:     "job1",
			Action: v1alpha2.JobRun,
		},
	})
	for i := 0; i < 30; i++ {
		requestGet := &v1alpha2.COARequest{
			Method:  fasthttp.MethodGet,
			Context: context.Background(),
			Parameters: map[string]string{
				"__site": SiteSpec.Name,
				"count":  "1",
			},
		}
		response = vendor.onSync(*requestGet)
		assert.Equal(t, v1alpha2.OK, response.State)
		var summary model.SyncPackage
		err = json.Unmarshal(response.Body, &summary)
		assert.Nil(t, err)
		if len(summary.Jobs) == 1 {
			assert.Equal(t, "job1", summary.Jobs[0].Id)
			break
		} else {
			time.Sleep(time.Second)
		}
	}

	vendor.Context.PubsubProvider.Publish("catalog", v1alpha2.Event{
		Metadata: map[string]string{
			"site": SiteSpec.Name,
		},
		Body: v1alpha2.JobData{
			Id:     "catalog1",
			Action: v1alpha2.JobUpdate,
		},
	})
	for i := 0; i < 30; i++ {
		requestGet := &v1alpha2.COARequest{
			Method:  fasthttp.MethodGet,
			Context: context.Background(),
			Parameters: map[string]string{
				"__site": SiteSpec.Name,
				"count":  "1",
			},
		}
		response = vendor.onSync(*requestGet)
		if response.State == v1alpha2.InternalError {
			break
		} else {
			time.Sleep(time.Second)
		}
	}

	var catalogState = model.CatalogState{
		ObjectMeta: model.ObjectMeta{
			Name: "catalog1",
		},
		Spec: &model.CatalogSpec{
			Type: "catalog",
			Properties: map[string]interface{}{
				"property1": "value1",
				"property2": "value2",
			},
			ParentName: "parent1",
			Generation: "1",
			Metadata: map[string]string{
				"metadata1": "value1",
				"metadata2": "value2",
			},
		},
	}
	err = vendor.CatalogsManager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.Nil(t, err)
	vendor.Context.PubsubProvider.Publish("catalog", v1alpha2.Event{
		Metadata: map[string]string{
			"site": SiteSpec.Name,
		},
		Body: v1alpha2.JobData{
			Id:     "catalog1",
			Action: v1alpha2.JobUpdate,
		},
	})
	for i := 0; i < 30; i++ {
		requestGet := &v1alpha2.COARequest{
			Method:  fasthttp.MethodGet,
			Context: context.Background(),
			Parameters: map[string]string{
				"__site": SiteSpec.Name,
				"count":  "1",
			},
		}
		response = vendor.onSync(*requestGet)
		assert.Equal(t, v1alpha2.OK, response.State)
		var summary model.SyncPackage
		err = json.Unmarshal(response.Body, &summary)
		assert.Nil(t, err)
		if len(summary.Catalogs) == 1 {
			assert.Equal(t, catalogState.ObjectMeta.Name, summary.Catalogs[0].ObjectMeta.Name)
			break
		} else {
			time.Sleep(time.Second)
		}
	}

}

// Commented due to race
// func TestFederationOnTrail(t *testing.T) {
// 	vendor := federationVendorInit()

// 	var trails = make([]v1alpha2.Trail, 0)
// 	trails = append(trails, v1alpha2.Trail{
// 		Origin: vendor.Config.SiteInfo.SiteId,
// 	})
// 	vendor.Context.PubsubProvider.Publish("trail", v1alpha2.Event{
// 		Metadata: map[string]string{
// 			"site": SiteSpec.Name,
// 		},
// 		Body: trails,
// 	})
// 	for i := 0; i < 3; i++ {
// 		for _, p := range vendor.TrailsManager.LedgerProviders {
// 			if mc, ok := p.(*mockledger.MockLedgerProvider); ok {
// 				if len(mc.LedgerData) == 1 {
// 					assert.Equal(t, vendor.Config.SiteInfo.SiteId, mc.LedgerData[0].Origin)
// 					break
// 				} else {
// 					time.Sleep(time.Second)
// 				}
// 			}
// 		}
// 	}

// 	requestPost := &v1alpha2.COARequest{
// 		Method:  fasthttp.MethodPost,
// 		Context: context.Background(),
// 		Parameters: map[string]string{
// 			"__name": SiteSpec.Name,
// 		},
// 	}

// 	response := vendor.onTrail(*requestPost)
// 	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
// }

func TestFederationOnK8SHook(t *testing.T) {
	vendor := federationVendorInit()

	var catalogState = model.CatalogState{
		ObjectMeta: model.ObjectMeta{
			Name: "catalog1",
		},
		Spec: &model.CatalogSpec{
			Type: "catalog",
			Properties: map[string]interface{}{
				"property1": "value1",
				"property2": "value2",
			},
			ParentName: "parent1",
			Generation: "1",
			Metadata: map[string]string{
				"metadata1": "value1",
				"metadata2": "value2",
			},
		},
	}

	b, err := json.Marshal(catalogState)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"objectType": "catalog",
		},
		Body: b,
	}
	response := vendor.onK8sHook(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestPatch := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
	}
	response = vendor.onK8sHook(*requestPatch)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}
