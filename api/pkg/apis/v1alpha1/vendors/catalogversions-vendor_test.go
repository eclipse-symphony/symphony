package vendors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/catalogversions"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	memorygraph "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph/memory"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var catalogversionState = model.CatalogVersionState{
	ObjectMeta: model.ObjectMeta{
		Name: "name1-v-version1",
	},
	Spec: &model.CatalogVersionSpec{
		CatalogType: "catalogVersion",
		Properties: map[string]interface{}{
			"property1": "value1",
			"property2": "value2",
		},
		//ParentName: "parent1",
		Metadata: map[string]string{
			"metadata1": "value1",
			"metadata2": "value2",
		},
		RootResource: "name1",
	},
}

func CreateSimpleChain(root string, length int, CTManager catalogversions.CatalogVersionsManager, catalogversion model.CatalogVersionState) error {
	if length < 1 {
		return errors.New("Length can not be less than 1.")
	}

	var newCatalogVersion model.CatalogVersionState
	jData, _ := json.Marshal(catalogversion)
	json.Unmarshal(jData, &newCatalogVersion)

	newCatalogVersion.ObjectMeta.Name = root
	newCatalogVersion.Spec.ParentName = ""
	newCatalogVersion.Spec.RootResource = validation.GetRootResourceFromName(root)
	err := CTManager.UpsertState(context.Background(), newCatalogVersion.ObjectMeta.Name, newCatalogVersion)
	if err != nil {
		return err
	}
	for i := 1; i < length; i++ {
		tmp := newCatalogVersion.ObjectMeta.Name
		var childCatalogVersion model.CatalogVersionState
		jData, _ := json.Marshal(newCatalogVersion)
		json.Unmarshal(jData, &childCatalogVersion)
		childCatalogVersion.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, i)
		childCatalogVersion.Spec.ParentName = tmp
		err := CTManager.UpsertState(context.Background(), childCatalogVersion.ObjectMeta.Name, childCatalogVersion)
		if err != nil {
			return err
		}
		newCatalogVersion = childCatalogVersion
	}
	return nil
}

func CreateSimpleBinaryTree(root string, depth int, CTManager catalogversions.CatalogVersionsManager, catalogversion model.CatalogVersionState) error {
	if depth < 1 {
		return errors.New("Depth can not be less than 1.")
	}

	var newCatalogVersion model.CatalogVersionState
	jData, _ := json.Marshal(catalogversion)
	json.Unmarshal(jData, &newCatalogVersion)

	newCatalogVersion.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, 0)
	newCatalogVersion.Spec.ParentName = ""
	newCatalogVersion.Spec.RootResource = validation.GetRootResourceFromName(root)
	err := CTManager.UpsertState(context.Background(), newCatalogVersion.ObjectMeta.Name, newCatalogVersion)
	if err != nil {
		return err
	}
	count := 1
	for i := 1; i < depth; i++ {
		levelSize := 1 << i
		for j := 0; j < levelSize; j++ {
			parentIndex := (count - 1) / 2
			var childCatalogVersion model.CatalogVersionState
			jData, _ := json.Marshal(newCatalogVersion)
			json.Unmarshal(jData, &childCatalogVersion)
			childCatalogVersion.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, count)
			childCatalogVersion.Spec.ParentName = fmt.Sprintf("%s-%d", root, parentIndex)
			err := CTManager.UpsertState(context.Background(), childCatalogVersion.ObjectMeta.Name, childCatalogVersion)
			if err != nil {
				return err
			}
			count++
		}
	}
	return nil
}

func CatalogVersionVendorInit() CatalogVersionsVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	graphProvider := &memorygraph.MemoryGraphProvider{}
	graphProvider.Init(memorygraph.MemoryGraphProviderConfig{})

	catalogversionProviders := make(map[string]providers.IProvider)
	catalogversionProviders["StateProvider"] = stateProvider
	catalogversionProviders["GraphProvider"] = graphProvider
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := CatalogVersionsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Route: "catalogversions",
		Managers: []managers.ManagerConfig{
			{
				Name: "catalogversion-manager",
				Type: "managers.symphony.catalogversions",
				Properties: map[string]string{
					"providers.persistentstate": "StateProvider",
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"catalogversion-manager": catalogversionProviders,
	}, &pubSubProvider)
	return vendor
}

func TestCatalogVersionGetInfo(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestCatalogVersionGetEndpoints(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "catalogversions/status", endpoints[len(endpoints)-1].Route)
}

func TestCatalogVersionOnCheck(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	b, err := json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    b,
	}

	response := vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	var catalogversionState = model.CatalogVersionState{
		Spec: &model.CatalogVersionSpec{
			CatalogType: "catalogVersion",
			Properties: map[string]interface{}{
				"property1": "value1",
				"property2": "value2",
			},
			Metadata: map[string]string{
				"metadata1": "value1",
				"metadata2": "value2",
			},
		},
	}
	b, err = json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost.Body = b

	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.BadRequest, response.State)
	assert.Contains(t, string(response.Body), "rootResource must be a non-empty string")

	requestPost.Body = []byte("Invalid input")
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.InternalError, response.State)

	catalogversionState = model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test1-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			CatalogType: "config",
			Metadata: map[string]string{
				"schema": "emailcheckschema:version1",
			},
			RootResource: "test1",
		},
	}
	b, err = json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost.Body = b
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.BadRequest, response.State)
	assert.Contains(t, string(response.Body), "could not find the required schema")

	schema := utils.Schema{
		Rules: map[string]utils.Rule{
			"email": {
				Pattern: "<email>",
			},
		},
	}
	schemaCatalogVersion := model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "emailcheckschema-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			CatalogType: "schema",
			Properties: map[string]interface{}{
				"spec": schema,
			},
			RootResource: "emailcheckschema",
		},
	}
	b, err = json.Marshal(schemaCatalogVersion)
	assert.Nil(t, err)
	requestPost.Body = b
	requestPost.Parameters = map[string]string{
		"__name": schemaCatalogVersion.ObjectMeta.Name,
	}
	response = vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	b, err = json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost.Body = b
	response = vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)
}

func TestCatalogVersionOnCheckNotSupport(t *testing.T) {
	vendor := CatalogVersionVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogversionState.ObjectMeta.Name,
		},
	}

	response := vendor.onCheck(*requestPost)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogVersionOnCatalogVersionsGet(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)
	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "name1-v-version1",
		},
	}

	response := vendor.onCatalogVersions(*requestGet)
	assert.Equal(t, v1alpha2.NotFound, response.State)

	catalogversionState.ObjectMeta.Name = "name1-v-version1"
	b, err := json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    b,
		Parameters: map[string]string{
			"__name": catalogversionState.ObjectMeta.Name,
		},
	}

	// The check should fail. This is a bug
	response = vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	response = vendor.onCatalogVersions(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summary model.CatalogVersionState
	err = json.Unmarshal(response.Body, &summary)
	assert.Nil(t, err)
	assert.Equal(t, catalogversionState.ObjectMeta.Name, summary.ObjectMeta.Name)

	requestGet.Parameters = nil
	response = vendor.onCatalogVersions(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summarys []model.CatalogVersionState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(summarys))
	assert.Equal(t, catalogversionState.ObjectMeta.Name, summarys[0].ObjectMeta.Name)
}

func TestCatalogVersionOnCatalogVersionsPost(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    []byte("wrongObject"),
		Parameters: map[string]string{
			"__name": catalogversionState.ObjectMeta.Name,
		},
	}

	response := vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.InternalError, response.State)

	b, err := json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost.Body = b
	requestPost.Parameters = nil
	response = vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.BadRequest, response.State)

	requestPost.Parameters = map[string]string{
		"__name": catalogversionState.ObjectMeta.Name,
	}
	response = vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "name1-v-version1",
		},
	}
	response = vendor.onCatalogVersions(*requestGet)
	assert.Equal(t, v1alpha2.OK, response.State)
	var summarys model.CatalogVersionState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, catalogversionState.ObjectMeta.Name, summarys.ObjectMeta.Name)
}

func TestCatalogVersionOnCatalogVersionsDelete(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	catalogversionState.ObjectMeta.Name = "name1-v-version1"
	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogversionState.ObjectMeta.Name,
		},
	}

	b, err := json.Marshal(catalogversionState)
	assert.Nil(t, err)
	requestPost.Body = b
	response := vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestPost.Parameters = map[string]string{
		"__name": catalogversionState.ObjectMeta.Name,
	}
	response = vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestDelete := &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "name1-v-version1",
		},
	}
	response = vendor.onCatalogVersions(*requestDelete)
	assert.Equal(t, v1alpha2.OK, response.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "name1-v-version1",
		},
	}
	response = vendor.onCatalogVersions(*requestGet)
	assert.Equal(t, v1alpha2.NotFound, response.State)
}

func TestCatalogVersionOnCatalogVersionsNotSupport(t *testing.T) {
	vendor := CatalogVersionVendorInit()

	requestPost := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": catalogversionState.ObjectMeta.Name,
		},
	}

	response := vendor.onCatalogVersions(*requestPost)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogVersionOnCatalogVersionsGraphGetChains(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "config-chains",
		},
	}

	catalogversionState.Spec.CatalogType = "config"
	err := CreateSimpleChain("root-v-version1", 4, *vendor.CatalogVersionsManager, catalogversionState)
	assert.Nil(t, err)

	response := vendor.onCatalogVersionsGraph(*requestGet)
	var summarys map[string][]model.CatalogVersionState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(summarys["root-v-version1"]))
}

func TestCatalogVersionOnCatalogVersionsGraphGetTrees(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "asset-trees",
		},
	}

	catalogversionState.Spec.CatalogType = "asset"
	err := CreateSimpleBinaryTree("root-v-version1", 3, *vendor.CatalogVersionsManager, catalogversionState)
	assert.Nil(t, err)

	response := vendor.onCatalogVersionsGraph(*requestGet)
	var summarys map[string][]model.CatalogVersionState
	err = json.Unmarshal(response.Body, &summarys)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(summarys["root-v-version1-0"]))
}

func TestCatalogVersionOnCatalogVersionsGraphGetUnknownTemplate(t *testing.T) {
	vendor := CatalogVersionVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "unkown-template",
		},
	}

	response := vendor.onCatalogVersionsGraph(*requestGet)
	assert.Equal(t, v1alpha2.BadRequest, response.State)
}

func TestCatalogVersionOnCatalogVersionsGraphMethodNotAllowed(t *testing.T) {
	vendor := CatalogVersionVendorInit()

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"template": "unkown-template",
		},
	}

	response := vendor.onCatalogVersionsGraph(*requestGet)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}

func TestCatalogVersionSubscribe(t *testing.T) {
	vendor := CatalogVersionVendorInit()
	vendor.CatalogVersionsManager.CatalogVersionValidator = validation.NewCatalogVersionValidator(vendor.CatalogVersionsManager.CatalogVersionLookup, nil, vendor.CatalogVersionsManager.ChildCatalogVersionLookup)

	origin := "parent"
	vendor.Context.Publish("catalogversion-sync", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": catalogversionState.Spec.CatalogType,
			"origin":     origin,
		},
		Body: v1alpha2.JobData{
			Id:     catalogversionState.ObjectMeta.Name,
			Action: v1alpha2.JobUpdate,
			Body:   catalogversionState,
		},
	})

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": fmt.Sprintf("%s-%s", origin, catalogversionState.ObjectMeta.Name),
		},
	}
	response := vendor.onCatalogVersions(*requestGet)

	for i := 0; i < 10; i++ {
		response = vendor.onCatalogVersions(*requestGet)

		if response.State != v1alpha2.OK {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	assert.Equal(t, v1alpha2.OK, response.State)
}
