/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package graph

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type GetRequest struct {
	Name   string `json:"name,omitempty"`
	Filter string `json:"filter,omitempty"`
}

type ListRequest struct {
	Filter string `json:"filter,omitempty"`
}

type GetSetResponse struct {
	Nodes []v1alpha2.INode `json:"nodes,omitempty"`
}

type GetGraphResponse struct {
	Nodes []v1alpha2.INode `json:"nodes,omitempty"`
	Edges []v1alpha2.IEdge `json:"edges,omitempty"`
}

type GetSetsResponse struct {
	Sets map[string]GetSetResponse `json:"sets,omitempty"`
}

type GetGraphsResponse struct {
	Graphs map[string]GetGraphResponse `json:"graphs,omitempty"`
}

// / IGraphProvider is the interface that provides a graph query interface that is able to return:
// / * a set of nodes (of a given parent), or all sets of nodes (GetSet, GetSets)
// / * a tree of nodes (of a given root), or all trees of nodes (GetTree, GetTrees)
// / * a graph of nodes and edges (of a given collection of edges and nodes), or all graphs of nodes and edges (GetGraph, GetGraphs)
// / * a chain of nodes (of a given root), or all chains of nodes (GetChain, GetChains)
// / If a graph provider is pure (for instnace, backed by a graph database engine), it means that it does not need to be initialized with a set of nodes,
// / and can return a graph of nodes and edges without any input.
// / Otherwise, it needs to be initialized with a set of nodes, and can return a graph of nodes and edges only if the input contains a set of nodes.
// / At this point, the behavior of Filter hasn't been clearly defined. It is used to filter the nodes returned by the graph provider, but it is not clear
// / what happens if filter breaks tree/chain/graph relationships.
type IGraphProvider interface {
	GetSet(ctx context.Context, request GetRequest) (GetSetResponse, error)
	GetTree(ctx context.Context, request GetRequest) (GetSetResponse, error)
	GetGraph(ctx context.Context, request GetRequest) (GetGraphResponse, error)
	GetChain(ctx context.Context, request GetRequest) (GetSetResponse, error)

	GetSets(ctx context.Context, request ListRequest) (GetSetsResponse, error)
	GetTrees(ctx context.Context, request ListRequest) (GetSetsResponse, error)
	GetChains(ctx context.Context, request ListRequest) (GetSetsResponse, error)
	GetGraphs(ctx context.Context, request ListRequest) (GetGraphsResponse, error)

	IsPure() bool
	SetData(data []v1alpha2.INode) error
}
