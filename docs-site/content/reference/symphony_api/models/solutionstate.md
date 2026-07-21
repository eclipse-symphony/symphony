---
type: docs
title: "Solution State"
linkTitle: "SolutionState"
description: ""
weight: 60
---

## SolutionState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the solution object. |
| Spec | SolutionSpec | Desired solution definition. |
| Status | SolutionStatus | Current solution status. |

## SolutionSpec

The solution spec is currently empty in the source tree.

## SolutionStatus

| Field | Type | Description |
|--------|--------|--------|
| Properties | map[string]string | Solution status properties. |
