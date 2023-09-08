/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package graph

import (
	"context"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
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
