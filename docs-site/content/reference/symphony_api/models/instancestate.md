---
type: docs
title: "Instance State"
linkTitle: "InstanceState"
description: ""
weight: 60
---

## InstanceState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the instance object. |
| Spec | InstanceSpec | Desired instance definition. |
| Status | InstanceStatus | Current deployable status. |

## InstanceSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Scope | string | Instance scope. |
| Parameters | map[string]string | Optional parameters. |
| Metadata | map[string]string | Additional metadata. |
| SolutionVersion | string | Bound solution version. |
| Target | TargetSelector | Target selection. |
| Topologies | []TopologySpec | Topology constraints. |
| Pipelines | []PipelineSpec | Skill pipelines. |
| IsDryRun | bool | Dry-run flag. |
| ActiveState | ActiveState | Active/inactive state. |

## TargetSelector

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Direct target name. |
| Selector | map[string]string | Label selector for target matching. |

## PipelineSpec

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Pipeline name. |
| Skill | string | Skill name. |
| Parameters | map[string]string | Pipeline parameters. |
