---
type: docs
title: "Solution Version State"
linkTitle: "SolutionVersion"
description: ""
weight: 60
---

## SolutionVersionState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the solution version object. |
| Spec | SolutionVersionSpec | Desired solution version definition. |

## SolutionVersionSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Metadata | map[string]string | Additional metadata. |
| Components | []ComponentSpec | Components included in the solution version. |
| Version | string | Version string. |
| RootResource | string | Root resource name. |
