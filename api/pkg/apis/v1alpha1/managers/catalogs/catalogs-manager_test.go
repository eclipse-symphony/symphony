/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

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
var catalogState = model.CatalogState{
	ObjectMeta: model.ObjectMeta{
		Name: "name1",
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
			"name":      "name1",
		},
	},
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

func CreateSimpleChain(root string, length int, CTManager CatalogsManager, catalog model.CatalogState) error {
	if length < 1 {
		return errors.New("Length can not be less than 1.")
	}

	var newCatalog model.CatalogState
	jData, _ := json.Marshal(catalog)
	json.Unmarshal(jData, &newCatalog)

	newCatalog.ObjectMeta.Name = root
	newCatalog.Spec.ParentName = ""
	err := CTManager.UpsertState(context.Background(), newCatalog.ObjectMeta.Name, newCatalog)
	if err != nil {
		return err
	}
	for i := 1; i < length; i++ {
		tmp := newCatalog.ObjectMeta.Name
		var childCatalog model.CatalogState
		jData, _ := json.Marshal(newCatalog)
		json.Unmarshal(jData, &childCatalog)
		childCatalog.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, i)
		childCatalog.Spec.ParentName = tmp
		err := CTManager.UpsertState(context.Background(), childCatalog.ObjectMeta.Name, childCatalog)
		if err != nil {
			return err
		}
		newCatalog = childCatalog
	}
	return nil
}

func CreateSimpleBinaryTree(root string, depth int, CTManager CatalogsManager, catalog model.CatalogState) error {
	if depth < 1 {
		return errors.New("Depth can not be less than 1.")
	}

	var newCatalog model.CatalogState
	jData, _ := json.Marshal(catalog)
	json.Unmarshal(jData, &newCatalog)

	newCatalog.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, 0)
	newCatalog.Spec.ParentName = ""
	err := CTManager.UpsertState(context.Background(), newCatalog.ObjectMeta.Name, newCatalog)
	if err != nil {
		return err
	}
	count := 1
	for i := 1; i < depth; i++ {
		levelSize := 1 << i
		for j := 0; j < levelSize; j++ {
			parentIndex := (count - 1) / 2
			var childCatalog model.CatalogState
			jData, _ := json.Marshal(newCatalog)
			json.Unmarshal(jData, &childCatalog)
			childCatalog.ObjectMeta.Name = fmt.Sprintf("%s-%d", root, count)
			childCatalog.Spec.ParentName = fmt.Sprintf("%s-%d", root, parentIndex)
			err := CTManager.UpsertState(context.Background(), childCatalog.ObjectMeta.Name, childCatalog)
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

	err = manager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
		return nil
	})
	val, err := manager.GetState(context.Background(), catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	equal, err := catalogState.DeepEquals(val)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestList(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
		return nil
	})
	val, err := manager.ListState(context.Background(), catalogState.ObjectMeta.Namespace, "", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(val))
	equal, err := catalogState.DeepEquals(val[0])
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestDelete(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalog", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "catalog", event.Metadata["objectType"])
		assert.Equal(t, "name1", job.Id)
		assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
		return nil
	})
	val, err := manager.GetState(context.Background(), catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	equal, err := catalogState.DeepEquals(val)
	assert.Nil(t, err)
	assert.True(t, equal)

	err = manager.DeleteState(context.Background(), catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	val, err = manager.GetState(context.Background(), catalogState.ObjectMeta.Name, catalogState.ObjectMeta.Namespace)
	assert.NotNil(t, err)
	assert.Empty(t, val)
}

func TestGetChains(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = CreateSimpleChain("root", 4, manager, catalogState)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background(), catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	tk, err := manager.ListState(context.Background(), catalogState.ObjectMeta.Namespace, "", "")
	assert.Nil(t, err)
	for _, v := range tk {
		fmt.Println(v.ObjectMeta.Name)
	}

	val, err := manager.GetChains(context.Background(), catalogState.Spec.Type, catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(val["root"]))
}

func TestGetTrees(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)

	err = CreateSimpleBinaryTree("root", 3, manager, catalogState)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background(), catalogState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	val, err := manager.GetTrees(context.Background(), catalogState.Spec.Type, catalogState.ObjectMeta.Namespace)
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
	catalogState.Spec.Properties = map[string]interface{}{
		"spec": schema,
	}
	catalogState.Spec.ParentName = ""
	catalogState.ObjectMeta = model.ObjectMeta{
		Name: "EmailCheckSchema",
	}
	err = manager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.Nil(t, err)

	catalogState.Spec.Metadata = map[string]string{
		"schema": "EmailCheckSchema",
	}
	catalogState.ObjectMeta = model.ObjectMeta{
		Name: "Email",
	}
	catalogState.Spec.Properties = map[string]interface{}{
		"email": "This is an invalid email",
	}

	err = manager.UpsertState(context.Background(), catalogState.ObjectMeta.Name, catalogState)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "schema validation error"))
}
