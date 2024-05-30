/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestProjectINode(t *testing.T) {
	catalog := &CatalogState{}
	var iNode v1alpha2.INode = catalog
	assert.NotNil(t, iNode)
}
func TestProjectIEdge(t *testing.T) {
	catalog := &CatalogState{}
	var iEdge v1alpha2.IEdge = catalog
	assert.NotNil(t, iEdge)
}
func TestIntefaceConvertion(t *testing.T) {
	var val interface{} = CatalogState{}
	var iNode v1alpha2.INode = val.(v1alpha2.INode)
	assert.NotNil(t, iNode)
}
func TestCatalogMatch(t *testing.T) {
	catalog1 := CatalogSpec{
		ParentName: "parentName",
		Generation: "1",
		Properties: map[string]interface{}{
			"key": "value",
		},
	}
	catalog2 := CatalogSpec{
		ParentName: "parentName",
		Generation: "1",
		Properties: map[string]interface{}{
			"key": "value",
		},
	}
	equal, err := catalog1.DeepEquals(catalog2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestCatalogMatchOneEmpty(t *testing.T) {
	catalog1 := CatalogSpec{
		Type: "type",
		Properties: map[string]interface{}{
			"key": "value",
		},
	}
	res, err := catalog1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a CatalogSpec type")
	assert.False(t, res)
}

func TestCatalogNotMatch(t *testing.T) {
	catalog1 := CatalogSpec{}
	catalog2 := CatalogSpec{}

	// parentName not match
	catalog1.ParentName = "parentName"
	catalog2.ParentName = "parentName2"
	equal, err := catalog1.DeepEquals(catalog2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// generation not match
	catalog2.ParentName = "parentName"
	catalog1.Generation = "1"
	catalog2.Generation = "2"
	equal, err = catalog1.DeepEquals(catalog2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// properties not match
	catalog2.Generation = "1"
	catalog1.Properties = map[string]interface{}{
		"key": "value",
	}
	catalog2.Properties = map[string]interface{}{
		"key": "value2",
	}
	equal, err = catalog1.DeepEquals(catalog2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestGetId(t *testing.T) {
	catalog := CatalogState{
		ObjectMeta: ObjectMeta{
			Name: "id",
		},
	}
	assert.Equal(t, catalog.GetId(), "id")
}

func TestGetParent(t *testing.T) {
	catalog := CatalogState{
		Spec: &CatalogSpec{
			ParentName: "parent",
		},
	}
	assert.Equal(t, catalog.GetParent(), "parent")

	catalog.Spec = nil
	assert.Equal(t, catalog.GetParent(), "")
}

func TestGetProperties(t *testing.T) {
	catalog := CatalogState{
		Spec: &CatalogSpec{
			Properties: map[string]interface{}{
				"key": "value",
			},
		},
	}
	assert.Equal(t, catalog.GetProperties(), map[string]interface{}{
		"key": "value",
	})

	catalog.Spec = nil
	assert.Equal(t, catalog.GetProperties(), map[string]interface{}(nil))
}

func TestGetType(t *testing.T) {
	catalog := CatalogState{
		Spec: &CatalogSpec{
			Type: "type",
		},
	}
	assert.Equal(t, catalog.GetType(), "type")

	catalog.Spec = nil
	assert.Equal(t, catalog.GetType(), "")
}

func TestGetFrom(t *testing.T) {
	catalog := CatalogState{
		Spec: &CatalogSpec{
			Type: "edge",
			Metadata: map[string]string{
				"from": "from",
			},
		},
	}
	assert.Equal(t, catalog.GetFrom(), "from")

	catalog.Spec = nil
	assert.Equal(t, catalog.GetFrom(), "")
}

func TestGetTo(t *testing.T) {
	catalog := CatalogState{
		Spec: &CatalogSpec{
			Type: "edge",
			Metadata: map[string]string{
				"to": "to",
			},
		},
	}
	assert.Equal(t, catalog.GetTo(), "to")

	catalog.Spec = nil
	assert.Equal(t, catalog.GetTo(), "")
}
