---
type: docs
title: "Skill State"
linkTitle: "SkillState"
description: ""
weight: 60
---

## SkillState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the skill object. |
| Spec | SkillSpec | Desired skill definition. |

## SkillSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Parameters | map[string]string | Skill parameters. |
| Nodes | []NodeSpec | Skill nodes. |
| Properties | map[string]string | Skill properties. |
| Bindings | []BindingSpec | Skill bindings. |
| Edges | []EdgeSpec | Skill edges. |

## SkillPackageSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Skill | string | Skill name. |
| Properties | map[string]string | Package properties. |
| Constraints | string | Package constraints. |
| Routes | []RouteSpec | Package routes. |
