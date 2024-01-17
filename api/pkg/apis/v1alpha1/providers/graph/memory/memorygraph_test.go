/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorygraph

import (
	"context"
	"fmt"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

type TestNode struct {
	Id         string                 `json:"id,omitempty"`
	Parent     string                 `json:"parent,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	From       string                 `json:"from,omitempty"`
	To         string                 `json:"to,omitempty"`
}

func (n *TestNode) GetId() string {
	return n.Id
}
func (n *TestNode) GetParent() string {
	return n.Parent
}
func (n *TestNode) GetType() string {
	return "mock"
}
func (n *TestNode) GetProperties() map[string]interface{} {
	return n.Properties
}
func (e *TestNode) GetFrom() string {
	return e.From
}
func (e *TestNode) GetTo() string {
	return e.To
}
func createFullGraph(nodes []string) []v1alpha2.INode {
	ret := make([]v1alpha2.INode, 0)
	for _, node := range nodes {
		ret = append(ret, &TestNode{
			Id: node,
		})
	}

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			ret = append(ret, &TestNode{
				From: nodes[i],
				To:   nodes[j],
			})
		}
	}

	return ret
}
func createSimpleChain(root string, length int) []v1alpha2.INode {
	nodes := make([]v1alpha2.INode, length)
	nodes[0] = &TestNode{
		Id:     root,
		Parent: "",
	}
	for i := 1; i < length; i++ {
		nodes[i] = &TestNode{
			Id:     fmt.Sprintf("%s-%d", root, i),
			Parent: nodes[i-1].GetId(),
		}
	}
	return nodes
}
func createSimpleBinaryTree(root string, depth int) []v1alpha2.INode {
	nodes := make([]v1alpha2.INode, 0)
	nodes = append(nodes, &TestNode{
		Id:     root,
		Parent: "",
	})
	for i := 1; i < depth; i++ {
		levelSize := 1 << i
		for j := 0; j < levelSize; j++ {
			parentIndex := (len(nodes) - 1) / 2
			parent := nodes[parentIndex].GetId()
			node := &TestNode{
				Id:     fmt.Sprintf("%s-%d", root, len(nodes)),
				Parent: parent,
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}
func createSimpleSet(parent string, count int) []v1alpha2.INode {
	nodes := make([]v1alpha2.INode, count+1)
	nodes[0] = &TestNode{
		Id:     parent,
		Parent: "",
	}
	for i := 0; i < count; i++ {
		nodes[i+1] = &TestNode{
			Id:     fmt.Sprintf("%s-%d", parent, i),
			Parent: parent,
		}
	}
	return nodes
}

func TestGetSet(t *testing.T) {
	provider := MemoryGraphProvider{}
	err := provider.Init(MemoryGraphProviderConfig{})
	assert.Nil(t, err)

	testNodes := []v1alpha2.INode{}
	testNodes = append(testNodes, createSimpleSet("parent", 5)...)
	provider.SetData(testNodes)

	res, err := provider.GetSet(context.Background(), graph.GetRequest{Name: "parent"})
	assert.Nil(t, err)
	assert.Equal(t, 5, len(res.Nodes))
}

func TestGetTree(t *testing.T) {
	provider := MemoryGraphProvider{}
	err := provider.Init(MemoryGraphProviderConfig{})
	assert.Nil(t, err)

	testNodes := []v1alpha2.INode{}
	testNodes = append(testNodes, createSimpleBinaryTree("root", 3)...)
	provider.SetData(testNodes)

	res, err := provider.GetTree(context.Background(), graph.GetRequest{Name: "root"})
	assert.Nil(t, err)
	assert.Equal(t, 7, len(res.Nodes))
}

func TestGetChain(t *testing.T) {
	provider := MemoryGraphProvider{}
	err := provider.Init(MemoryGraphProviderConfig{})
	assert.Nil(t, err)

	testNodes := []v1alpha2.INode{}
	testNodes = append(testNodes, createSimpleChain("root", 3)...)
	provider.SetData(testNodes)

	res, err := provider.GetChain(context.Background(), graph.GetRequest{Name: "root"})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res.Nodes))
}

func TestGetChainSingleNode(t *testing.T) {
	provider := MemoryGraphProvider{}
	err := provider.Init(MemoryGraphProviderConfig{})
	assert.Nil(t, err)

	testNodes := []v1alpha2.INode{}
	testNodes = append(testNodes, createSimpleChain("root", 1)...)
	provider.SetData(testNodes)

	res, err := provider.GetChain(context.Background(), graph.GetRequest{Name: "root"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res.Nodes))
}

func TestGetChains(t *testing.T) {
	provider := MemoryGraphProvider{}
	err := provider.Init(MemoryGraphProviderConfig{})
	assert.Nil(t, err)

	testNodes := []v1alpha2.INode{}
	testNodes = append(testNodes, createSimpleChain("root", 3)...)
	testNodes = append(testNodes, createSimpleChain("root2", 5)...)
	provider.SetData(testNodes)

	res, err := provider.GetChains(context.Background(), graph.ListRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res.Sets))
	assert.Equal(t, 3, len(res.Sets["root"].Nodes))
	assert.Equal(t, 5, len(res.Sets["root2"].Nodes))

}
