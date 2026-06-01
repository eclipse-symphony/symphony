/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogversions

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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

var manager CatalogVersionsManager
var catalogversionState = model.CatalogVersionState{
	ObjectMeta: model.ObjectMeta{
		Name: "name1-v-version1",
	},
	Spec: &model.CatalogVersionSpec{
		CatalogType: "catalogversion",
		Properties: map[string]interface{}{
			"property1": "value1",
			"property2": "value2",
		},
		// ParentName: "parent1",
		Metadata: map[string]string{
			"metadata1": "value1",
			"metadata2": "value2",
			"name":      "name1",
		},
		RootResource: "name1",
	},
}

func initalizeManager() error {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	graphProvider := &memorygraph.MemoryGraphProvider{}
	graphProvider.Init(memorygraph.MemoryGraphProviderConfig{})

	manager = CatalogVersionsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "StateProvider",
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

func CreateSimpleChain(root string, length int, CTManager CatalogVersionsManager, catalogversion model.CatalogVersionState) error {
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

func CreateSimpleBinaryTree(root string, depth int, CTManager CatalogVersionsManager, catalogversion model.CatalogVersionState) error {
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

func TestInit(t *testing.T) {

	err := initalizeManager()
	assert.Nil(t, err)
}

func TestUpsertAndGet(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil

	err = manager.UpsertState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalogversion", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			var job v1alpha2.JobData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &job)
			assert.Nil(t, err)
			assert.Equal(t, "catalogversion", event.Metadata["objectType"])
			assert.Equal(t, "name1", job.Id)
			assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
			return nil
		},
	})
	val, err := manager.GetState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	// Upsert state will set rootResource label on the object. Reset it before comparison
	val.ObjectMeta.Labels = nil
	equal, err := catalogversionState.DeepEquals(val)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestList(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil

	err = manager.UpsertState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalogversion", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			var job v1alpha2.JobData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &job)
			assert.Nil(t, err)
			assert.Equal(t, "catalogversion", event.Metadata["objectType"])
			assert.Equal(t, "name1", job.Id)
			assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
			return nil
		},
	})
	val, err := manager.ListState(context.Background(), catalogversionState.ObjectMeta.Namespace, "", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(val))
	// Upsert state will set rootResource label on the object. Reset it before comparison
	list := val[0]
	list.ObjectMeta.Labels = nil
	equal, err := catalogversionState.DeepEquals(list)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestDelete(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil

	err = manager.UpsertState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState)
	assert.Nil(t, err)
	manager.Context.Subscribe("catalogversion", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			var job v1alpha2.JobData
			jData, _ := json.Marshal(event.Body)
			err := json.Unmarshal(jData, &job)
			assert.Nil(t, err)
			assert.Equal(t, "catalogversion", event.Metadata["objectType"])
			assert.Equal(t, "name1", job.Id)
			assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
			return nil
		},
	})
	val, err := manager.GetState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	// Upsert state will set rootResource label on the object. Reset it before comparison
	val.ObjectMeta.Labels = nil
	equal, err := catalogversionState.DeepEquals(val)
	assert.Nil(t, err)
	assert.True(t, equal)

	err = manager.DeleteState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	val, err = manager.GetState(context.Background(), catalogversionState.ObjectMeta.Name, catalogversionState.ObjectMeta.Namespace)
	assert.NotNil(t, err)
	assert.Empty(t, val)
}

func TestGetChains(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil
	err = CreateSimpleChain("root-v-version1", 4, manager, catalogversionState)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background(), catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	tk, err := manager.ListState(context.Background(), catalogversionState.ObjectMeta.Namespace, "", "")
	assert.Nil(t, err)
	for _, v := range tk {
		fmt.Println(v.ObjectMeta.Name)
	}

	val, err := manager.GetChains(context.Background(), catalogversionState.Spec.CatalogType, catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(val["root-v-version1"]))
}

func TestGetTrees(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil
	err = CreateSimpleBinaryTree("root-v-version1", 3, manager, catalogversionState)
	assert.Nil(t, err)
	err = manager.setProviderDataIfNecessary(context.Background(), catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)

	val, err := manager.GetTrees(context.Background(), catalogversionState.Spec.CatalogType, catalogversionState.ObjectMeta.Namespace)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(val["root-v-version1-0"]))
}

func TestSchemaCheck(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil
	schema := utils.Schema{
		Rules: map[string]utils.Rule{
			"email": {
				Pattern: "<email>",
			},
		},
	}
	schemaCatalogVersion := model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "EmailCheckSchema-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "EmailCheckSchema",
			CatalogType:  "schema",
			Properties: map[string]interface{}{
				"spec": schema,
			},
		},
	}

	err = manager.UpsertState(context.Background(), schemaCatalogVersion.ObjectMeta.Name, schemaCatalogVersion)
	assert.Nil(t, err)

	catalogversion := model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "Email-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "Email",
			CatalogType:  "catalogVersion",
			Metadata: map[string]string{
				"schema": "EmailCheckSchema:version1",
			},
			Properties: map[string]interface{}{
				"email": "This is an invalid email",
			},
		},
	}

	err = manager.UpsertState(context.Background(), catalogversion.ObjectMeta.Name, catalogversion)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "email: property does not match pattern"))
}

func TestParentCatalogVersion(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	manager.CatalogVersionValidator.CatalogLookupFunc = nil
	childCatalogVersion := model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "EmailCheckSchema-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "EmailCheckSchema",
			CatalogType:  "schema",
			ParentName:   "parent:version1",
		},
	}

	err = manager.UpsertState(context.Background(), childCatalogVersion.ObjectMeta.Name, childCatalogVersion)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parent catalogversion not found")

	parentCatalogVersion := model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "parent-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "parent",
			CatalogType:  "catalogVersion",
		},
	}

	err = manager.UpsertState(context.Background(), parentCatalogVersion.ObjectMeta.Name, parentCatalogVersion)
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), childCatalogVersion.ObjectMeta.Name, childCatalogVersion)
	assert.Nil(t, err)

	parentCatalogVersion = model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "parent-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "parent",
			CatalogType:  "catalogVersion",
			ParentName:   "EmailCheckSchema:version1",
		},
	}
	err = manager.UpsertState(context.Background(), parentCatalogVersion.ObjectMeta.Name, parentCatalogVersion)
	assert.Contains(t, err.Error(), "parent catalogversion has circular dependency")

	err = manager.DeleteState(context.Background(), parentCatalogVersion.ObjectMeta.Name, parentCatalogVersion.ObjectMeta.Namespace)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CatalogVersion has one or more child catalogversions. Update or Deletion is not allowed")
}

/*
func TestCatalogVersion(t *testing.T) {
	err := initalizeManager()
	assert.Nil(t, err)
	err = manager.UpsertState(context.Background(), "test-v-version1", model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rootResource must be a valid container")
	manager.StateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "test",
			Body: map[string]interface{}{
				"apiVersion": model.FederationGroup + "/v1",
				"kind":       "CatalogVersion",
				"metadata": model.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				"spec": model.CatalogVersionState{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FederationGroup,
			"version":   "v1",
			"resource":  "catalogversions",
			"kind":      "CatalogVersion",
		},
	})

	err = manager.UpsertState(context.Background(), "test-v-version1", model.CatalogVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CatalogVersionSpec{
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
}
*/
