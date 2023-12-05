/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
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
