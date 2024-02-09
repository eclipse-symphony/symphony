package vendors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogs"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	memorygraph "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph/memory"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var catalogSpec = model.CatalogSpec{
	SiteId: "site1",
	Name:   "name1",
	Type:   "catalog",
	Properties: map[string]interface{}{
		"property1": "value1",
		"property2": "value2",
	},
	Metadata: map[string]string{
		"metadata1": "value1",
		"metadata2": "value2",
	},
	ParentName: "parent1",
	Generation: "1",
}

func CreateSimpleChain(root string, length int, CTManager *catalogs.CatalogsManager, catalog model.CatalogSpec) error {
	if length < 1 {
		return errors.New("Length can not be less than 1.")
	}

	catalog.Name = root
	catalog.ParentName = ""
	err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog, "default")
	if err != nil {
		return err
	}
	for i := 1; i < length; i++ {
		tmp := catalog.Name
		catalog.Name = fmt.Sprintf("%s-%d", root, i)
		catalog.ParentName = tmp
		err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog, "default")
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateSimpleBinaryTree(root string, depth int, CTManager *catalogs.CatalogsManager, catalog model.CatalogSpec) error {
	if depth < 1 {
		return errors.New("Depth can not be less than 1.")
	}
	catalog.Name = fmt.Sprintf("%s-%d", root, 0)
	catalog.ParentName = ""
	err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog, "default")
	if err != nil {
		return err
	}
	count := 1
	for i := 1; i < depth; i++ {
		levelSize := 1 << i
		for j := 0; j < levelSize; j++ {
			parentIndex := (count - 1) / 2
			catalog.Name = fmt.Sprintf("%s-%d", root, count)
			catalog.ParentName = fmt.Sprintf("%s-%d", root, parentIndex)
			err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog, "default")
			if err != nil {
				return err
			}
			count++
		}
	}
	return nil
}

func CatalogVendorInit() CatalogsVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	graphProvider := &memorygraph.MemoryGraphProvider{}
	graphProvider.Init(memorygraph.MemoryGraphProviderConfig{})

	catalogProviders := make(map[string]providers.IProvider)
	catalogProviders["StateProvider"] = stateProvider
	catalogProviders["GraphProvider"] = graphProvider
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := CatalogsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Route: "catalogs",
		Managers: []managers.ManagerConfig{
			{
				Name: "catalog-manager",
				Type: "managers.symphony.catalogs",
				Properties: map[string]string{
					"providers.state": "StateProvider",
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"catalog-manager": catalogProviders,
	}, &pubSubProvider)
	return vendor
}

func TestCatalogGetInfo(t *testing.T) {
	vendor := CatalogVendorInit()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestCatalogGetEndpoints(t *testing.T) {
	vendor := CatalogVendorInit()
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "catalogs/check", endpoints[len(endpoints)-1].Route)
}

func TestCatalogOnCheck(t *testing.T) {
	vendor := CatalogVendorInit()

	b, err := json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    b,
	}

	response := vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	var catalogSpec = model.CatalogSpec{
		SiteId: "site1",
		Type:   "catalog",
		Properties: map[string]interface{}{
			"property1": "value1",
			"property2": "value2",
		},
		Metadata: map[string]string{
			"metadata1": "value1",
			"metadata2": "value2",
		},
		ParentName: "parent1",
		Generation: "1",
	}
	b, err = json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	// The check should fail. This is a bug
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestPost.Body = []byte("Invalid input")
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.InternalError, response.State)

	catalogSpec.Name = "test1"
	catalogSpec.Metadata = map[string]string{
		"schema": "EmailCheckSchema",
	}
	b, err = json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.InternalError, response.State)

	schema := utils.Schema{
		Rules: map[string]utils.Rule{
			"email": {
				Pattern: "<email>",
			},
		},
	}
	catalogSpec.Properties = map[string]interface{}{
		"spec": schema,
	}
	catalogSpec.Name = "EmailCheckSchema"
	catalogSpec.ParentName = ""
	catalogSpec.Metadata = nil
	b, err = json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	requestPost.Parameters = map[string]string{
		"__name": catalogSpec.Name,
	}
	response = vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	catalogSpec.Name = "test1"
	catalogSpec.Metadata = map[string]string{
		"schema": "EmailCheckSchema",
	}
	b, err = json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)
}

func TestCatalogOnCheckNotSupport(t *testing.T) {
	vendor := CatalogVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogSpec.Name,
		},
	}

	response := vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogOnCatalogsGet(t *testing.T) {
	vendor := CatalogVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test1",
		},
	}

	response := vendor.onCatalogs(*requestGet)
	assert.Equal(t, v1alpha2.NotFound, response.State)

	catalogSpec.Name = "test1"
	b, err := json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    b,
		Parameters: map[string]string{
			"__name": catalogSpec.Name,
		},
	}

	// The check should fail. This is a bug
	response = vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	response = vendor.onCatalogs(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summary model.CatalogState
	err = json.Unmarshal(response.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, catalogSpec.Name, summary.Spec.Name)

	requestGet.Parameters = nil
	response = vendor.onCatalogs(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summarys []model.CatalogState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(summarys))
	assert.Equal(t, catalogSpec.Name, summarys[0].Spec.Name)
}

func TestCatalogOnCatalogsPost(t *testing.T) {
	vendor := CatalogVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    []byte("wrongObject"),
		Parameters: map[string]string{
			"__name": catalogSpec.Name,
		},
	}

	response := vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.InternalError, response.State)

	catalogSpec.Name = "test1"
	b, err := json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	requestPost.Parameters = nil
	response = vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.BadRequest, response.State)

	requestPost.Parameters = map[string]string{
		"__name": catalogSpec.Name,
	}
	response = vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test1",
		},
	}
	response = vendor.onCatalogs(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summarys model.CatalogState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, catalogSpec.Name, summarys.Spec.Name)
}

func TestCatalogOnCatalogsDelete(t *testing.T) {
	vendor := CatalogVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogSpec.Name,
		},
	}

	catalogSpec.Name = "test1"
	b, err := json.Marshal(catalogSpec)
	assert.Nil(t, err)
	requestPost.Body = b
	response := vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestPost.Parameters = map[string]string{
		"__name": catalogSpec.Name,
	}
	response = vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestDelete := &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test1",
		},
	}
	response = vendor.onCatalogs(*requestDelete)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test1",
		},
	}
	response = vendor.onCatalogs(*requestGet)
	assert.Equal(t, v1alpha2.NotFound, response.State)
}

func TestCatalogOnCatalogsNotSupport(t *testing.T) {
	vendor := CatalogVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogSpec.Name,
		},
	}

	response := vendor.onCatalogs(*requestPost)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogOnCatalogsGraphGetChains(t *testing.T) {
	vendor := CatalogVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "config-chains",
		},
	}

	catalogSpec.Type = "config"
	err := CreateSimpleChain("root", 4, vendor.CatalogsManager, catalogSpec)
	assert.Nil(t, err)

	response := vendor.onCatalogsGraph(*requestGet)
	var summarys map[string][]model.CatalogState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(summarys["root"]))
}

func TestCatalogOnCatalogsGraphGetTrees(t *testing.T) {
	vendor := CatalogVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "asset-trees",
		},
	}

	catalogSpec.Type = "asset"
	err := CreateSimpleBinaryTree("root", 3, vendor.CatalogsManager, catalogSpec)
	assert.Nil(t, err)

	response := vendor.onCatalogsGraph(*requestGet)
	var summarys map[string][]model.CatalogState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(summarys["root-0"]))
}

func TestCatalogOnCatalogsGraphGetUnknownTemplate(t *testing.T) {
	vendor := CatalogVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "unkown-template",
		},
	}

	response := vendor.onCatalogsGraph(*requestGet)
	assert.Equal(t, v1alpha2.BadRequest, response.State)
}

func TestCatalogOnCatalogsGraphMethodNotAllowed(t *testing.T) {
	vendor := CatalogVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "unkown-template",
		},
	}

	response := vendor.onCatalogsGraph(*requestGet)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogSubscribe(t *testing.T) {
	vendor := CatalogVendorInit()

	vendor.Context.Publish("catalog-sync", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": catalogSpec.Type,
		},
		Body: v1alpha2.JobData{
			Id:     catalogSpec.Name,
			Action: "UPDATE",
			Body:   catalogSpec,
		},
	})

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": fmt.Sprintf("%s-%s", catalogSpec.SiteId, catalogSpec.Name),
		},
	}
	response := vendor.onCatalogs(*requestGet)

	for i := 0; i < 10; i++ {
		response = vendor.onCatalogs(*requestGet)

		if response.State != v1alpha2.OK {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	assert.Equal(t, v1alpha2.OK, response.State)
}
