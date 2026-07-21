---
type: docs
title: "Component and Routing Models"
linkTitle: "ComponentSpec"
description: ""
weight: 60
---

## ComponentSpec

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Component name. |
| Type | string | Optional component type. |
| Metadata | map[string]string | Component metadata. |
| Properties | map[string]interface{} | Component properties. |
| Parameters | map[string]string | Component parameters. |
| Routes | []RouteSpec | Component routes. |
| Constraints | string | Placement constraints. |
| Dependencies | []string | Component dependencies. |
| Skills | []string | Skills referenced by the component. |
| Sidecars | []SidecarSpec | Sidecar definitions. |

## BindingSpec

| Field | Type | Description |
|--------|--------|--------|
| Role | string | Binding role. |
| Provider | string | Binding provider. |
| Config | map[string]string | Provider config. |

## RouteSpec

| Field | Type | Description |
|--------|--------|--------|
| Route | string | Route pattern. |
| Type | string | Route type. |
| Properties | map[string]string | Route properties. |
| Filters | []FilterSpec | Route filters. |

## NodeSpec

| Field | Type | Description |
|--------|--------|--------|
| Id | string | Node identifier. |
| NodeType | string | Node type. |
| Name | string | Node name. |
| Configurations | map[string]string | Node configurations. |
| Inputs | []RouteSpec | Input routes. |
| Outputs | []RouteSpec | Output routes. |
| Model | string | Model name. |

## TopologySpec

| Field | Type | Description |
|--------|--------|--------|
| Device | string | Target device. |
| Selector | map[string]string | Selector used to match devices. |
| Bindings | []BindingSpec | Bindings associated with the topology. |
