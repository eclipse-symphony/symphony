---
type: docs
title: "Target State"
linkTitle: "TargetState"
description: ""
weight: 60
---

## TargetState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the target object. |
| Status | TargetStatus | Current target deployable status. |
| Spec | TargetSpec | Desired target definition. |

## TargetSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Scope | string | Target scope. |
| SolutionVersionScope | string | SolutionVersion scope. |
| Metadata | map[string]string | Additional metadata. |
| Properties | map[string]string | Target properties. |
| Components | []ComponentSpec | Components to deploy. |
| Constraints | string | Placement constraints. |
| Topologies | []TopologySpec | Topology constraints. |
| ForceRedeploy | bool | Force redeploy behavior. |
| IsDryRun | bool | Dry-run flag. |
