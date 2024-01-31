package catalogs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	memorygraph "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph/memory"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

var manager CatalogsManager
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

func initalizeManager() error {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	graphProvider := &memorygraph.MemoryGraphProvider{}
	graphProvider.Init(memorygraph.MemoryGraphProviderConfig{})

	manager = CatalogsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	providers["GraphProvider"] = graphProvider
	vendorContext := &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendorContext.Init(&pubSubProvider)
	err := manager.Init(vendorContext, config, providers)
	return err
}

func CreateSimpleChain(root string, length int, CTManager CatalogsManager, catalog model.CatalogSpec) error {
	if length < 1 {
		return errors.New("Length can not be less than 1.")
	}

	catalog.Name = root
	catalog.ParentName = ""
	err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog)
	if err != nil {
		return err
	}
	for i := 1; i < length; i++ {
		tmp := catalog.Name
		catalog.Name = fmt.Sprintf("%s-%d", root, i)
		catalog.ParentName = tmp
		err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateSimpleBinaryTree(root string, depth int, CTManager CatalogsManager, catalog model.CatalogSpec) error {
	if depth < 1 {
		return errors.New("Depth can not be less than 1.")
	}
	catalog.Name = fmt.Sprintf("%s-%d", root, 0)
	catalog.ParentName = ""
	err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog)
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
			err := CTManager.UpsertSpec(context.Background(), catalog.Name, catalog)
			if err != nil {
				return err
			}
			count++
		}
	}
	return nil
}

func TestInit(t *testing.T) {

	err := initalizeManager()
	assert.Nil(t, err)
}

func TestUpsertAndGet(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = manager.UpsertSpec(context.Background(), catalogSpec.Name, catalogSpec)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == "UPDATE" || job.Action == "DELETE")
		return nil
	})
	val, err := manager.GetSpec(context.Background(), catalogSpec.Name)
	assert.Nil(t, err)
	assert.Equal(t, catalogSpec, *val.Spec)
}

func TestList(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = manager.UpsertSpec(context.Background(), catalogSpec.Name, catalogSpec)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == "UPDATE" || job.Action == "DELETE")
		return nil
	})
	val, err := manager.ListSpec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(val))
	assert.Equal(t, catalogSpec, *val[0].Spec)
}

func TestDelete(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = manager.UpsertSpec(context.Background(), catalogSpec.Name, catalogSpec)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == "UPDATE" || job.Action == "DELETE")
		return nil
	})
	val, err := manager.GetSpec(context.Background(), catalogSpec.Name)
	assert.Nil(t, err)
	assert.Equal(t, catalogSpec, *val.Spec)

	err = manager.DeleteSpec(context.Background(), catalogSpec.Name)
	assert.Nil(t, err)

	val, err = manager.GetSpec(context.Background(), catalogSpec.Name)
	assert.NotNil(t, err)
	assert.Empty(t, val)
}

func TestGetChains(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = CreateSimpleChain("root", 4, manager, catalogSpec)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background())
	assert.Nil(t, err)

	val, err := manager.GetChains(context.Background(), catalogSpec.Type)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(val["root"]))
}

func TestGetTrees(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = CreateSimpleBinaryTree("root", 3, manager, catalogSpec)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background())
	assert.Nil(t, err)

	val, err := manager.GetTrees(context.Background(), catalogSpec.Type)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(val["root-0"]))
}

func TestSchemaCheck(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

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
	err = manager.UpsertSpec(context.Background(), catalogSpec.Name, catalogSpec)
	assert.Nil(t, err)

	catalogSpec.Name = "Email"
	catalogSpec.Metadata = map[string]string{
		"schema": "EmailCheckSchema",
	}
	catalogSpec.Properties = map[string]interface{}{
		"email": "This is an invalid email",
	}

	err = manager.UpsertSpec(context.Background(), catalogSpec.Name, catalogSpec)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "schema validation error"))
}
