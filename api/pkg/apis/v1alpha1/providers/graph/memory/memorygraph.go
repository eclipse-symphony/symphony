/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorygraph

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/graph"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type MemoryGraphProviderConfig struct {
}

type MemoryGraphProvider struct {
	Config  MemoryGraphProviderConfig
	Context *contexts.ManagerContext
	Data    []v1alpha2.INode
}

func (g *MemoryGraphProvider) Init(config providers.IProviderConfig) error {
	mockConfig, err := toMemoryGraphProviderConfig(config)
	if err != nil {
		return err
	}
	g.Config = mockConfig
	return nil
}
func (s *MemoryGraphProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func toMemoryGraphProviderConfig(config providers.IProviderConfig) (MemoryGraphProviderConfig, error) {
	ret := MemoryGraphProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MemoryGraphProvider) InitWithMap(properties map[string]string) error {
	config, err := MemoryhGraphProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MemoryhGraphProviderConfigFromMap(properties map[string]string) (MemoryGraphProviderConfig, error) {
	ret := MemoryGraphProviderConfig{}
	return ret, nil
}
func (i *MemoryGraphProvider) GetSet(ctx context.Context, request graph.GetRequest) (graph.GetSetResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetSet",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	ret := graph.GetSetResponse{
		Nodes: make([]v1alpha2.INode, 0),
	}
	_, err = i.getNode(request.Name, request.Filter)
	if err != nil {
		return ret, err
	}
	for _, node := range i.Data {
		if request.Filter != "" && node.GetType() != request.Filter {
			continue
		}
		if node.GetParent() == request.Name {
			ret.Nodes = append(ret.Nodes, node)
		}
	}
	return ret, nil
}
func (i *MemoryGraphProvider) GetTree(ctx context.Context, request graph.GetRequest) (graph.GetSetResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetTree",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	ret := graph.GetSetResponse{
		Nodes: make([]v1alpha2.INode, 0),
	}
	root, err := i.getNode(request.Name, request.Filter)
	if err != nil {
		return ret, err
	}
	ret.Nodes = append(ret.Nodes, root)
	i.collectChildren(root, request.Filter, &ret)
	return ret, nil
}
func (i *MemoryGraphProvider) collectChildren(root v1alpha2.INode, filter string, ret *graph.GetSetResponse) {
	queue := []v1alpha2.INode{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for _, child := range i.Data {
			if filter != "" && child.GetType() != filter {
				continue
			}
			if child.GetParent() == node.GetId() {
				ret.Nodes = append(ret.Nodes, child)
				queue = append(queue, child)
			}
		}
	}
}
func (i *MemoryGraphProvider) GetGraph(ctx context.Context, request graph.GetRequest) (graph.GetGraphResponse, error) {
	return graph.GetGraphResponse{}, v1alpha2.NewCOAError(nil, "not implemented", v1alpha2.NotImplemented)
}
func (i *MemoryGraphProvider) getNode(name string, filter string) (v1alpha2.INode, error) {
	var root v1alpha2.INode
	for _, node := range i.Data {
		if filter != "" && node.GetType() != filter {
			continue
		}
		if node.GetId() == name {
			root = node
			break
		}
	}
	if root == nil {
		return nil, v1alpha2.NewCOAError(nil, "root node not found", v1alpha2.NotFound)
	}
	return root, nil
}
func (i *MemoryGraphProvider) GetChain(ctx context.Context, request graph.GetRequest) (graph.GetSetResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetChain",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	rep, err := i.GetTree(ctx, request)
	return rep, err
}
func (i *MemoryGraphProvider) GetSets(ctx context.Context, request graph.ListRequest) (graph.GetSetsResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetSets",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	seenSets := make(map[string]bool)
	ret := graph.GetSetsResponse{
		Sets: make(map[string]graph.GetSetResponse),
	}
	for _, node := range i.Data {
		if request.Filter != "" && node.GetType() != request.Filter {
			continue
		}
		if node.GetParent() == "" && !seenSets[node.GetId()] {
			seenSets[node.GetId()] = true
			var set graph.GetSetResponse
			set, err = i.GetSet(ctx, graph.GetRequest{
				Name: node.GetId(),
			})
			if err != nil {
				return ret, err
			}
			ret.Sets[node.GetId()] = set
		}
	}
	return ret, nil

}
func (i *MemoryGraphProvider) GetTrees(ctx context.Context, request graph.ListRequest) (graph.GetSetsResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetTrees",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	seenSets := make(map[string]bool)
	ret := graph.GetSetsResponse{
		Sets: make(map[string]graph.GetSetResponse),
	}
	for _, node := range i.Data {
		if request.Filter != "" && node.GetType() != request.Filter {
			continue
		}
		if node.GetParent() == "" && !seenSets[node.GetId()] {
			seenSets[node.GetId()] = true
			var set graph.GetSetResponse
			set, err = i.GetTree(ctx, graph.GetRequest{
				Name: node.GetId(),
			})
			if err != nil {
				return ret, err
			}
			ret.Sets[node.GetId()] = set
		}
	}
	return ret, nil
}
func (i *MemoryGraphProvider) GetChains(ctx context.Context, request graph.ListRequest) (graph.GetSetsResponse, error) {
	ctx, span := observability.StartSpan("Memory Graph Provider", ctx, &map[string]string{
		"method": "GetChains",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	seenSets := make(map[string]bool)
	ret := graph.GetSetsResponse{
		Sets: make(map[string]graph.GetSetResponse),
	}
	for _, node := range i.Data {
		if request.Filter != "" && node.GetType() != request.Filter {
			continue
		}
		if node.GetParent() == "" && !seenSets[node.GetId()] {
			seenSets[node.GetId()] = true
			var set graph.GetSetResponse
			set, err = i.GetChain(ctx, graph.GetRequest{
				Name: node.GetId(),
			})
			if err != nil {
				return ret, err
			}
			ret.Sets[node.GetId()] = set
		}
	}
	return ret, nil
}
func (i *MemoryGraphProvider) GetGraphs(ctx context.Context, request graph.ListRequest) (graph.GetGraphsResponse, error) {
	return graph.GetGraphsResponse{}, v1alpha2.NewCOAError(nil, "not implemented", v1alpha2.NotImplemented)
}

func (i *MemoryGraphProvider) IsPure() bool {
	return false
}
func (i *MemoryGraphProvider) SetData(data []v1alpha2.INode) error {
	i.Data = data
	return nil
}
