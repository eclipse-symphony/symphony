---
type: docs
title: "Model State"
linkTitle: "ModelState"
description: ""
weight: 60
---

## ModelState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the model object. |
| Spec | ModelSpec | Desired model definition. |

## ModelSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Properties | map[string]string | Model properties. |
| Constraints | string | Model constraints. |
| Bindings | []BindingSpec | Model bindings. |
