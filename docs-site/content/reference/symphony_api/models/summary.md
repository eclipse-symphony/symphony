---
type: docs
title: "Summary Models"
linkTitle: "Summary"
description: ""
weight: 60
---

## ComponentResultSpec

| Field | Type | Description |
|--------|--------|--------|
| Status | State | Component result status. |
| Message | string | Component result message. |

## TargetResultSpec

| Field | Type | Description |
|--------|--------|--------|
| Status | string | Target status. |
| Message | string | Target message. |
| ComponentResults | map[string]ComponentResultSpec | Per-component results. |

## SummarySpec

| Field | Type | Description |
|--------|--------|--------|
| TargetCount | int | Number of targets. |
| SuccessCount | int | Number of successful targets. |
| PlannedDeployment | int | Planned deployment count. |
| CurrentDeployed | int | Current deployed count. |
| TargetResults | map[string]TargetResultSpec | Per-target results. |
| SummaryMessage | string | Summary message. |
| JobID | string | Job identifier. |
| Skipped | bool | Whether deployment was skipped. |
| IsRemoval | bool | Whether this was a removal. |
| AllAssignedDeployed | bool | Whether all assigned targets were deployed. |
| Removed | bool | Whether removal completed. |

## SummaryResult

| Field | Type | Description |
|--------|--------|--------|
| Summary | SummarySpec | Embedded summary data. |
| SummaryId | string | Summary identifier. |
| Generation | string | Generation marker. |
| Time | time.Time | Timestamp. |
| State | SummaryState | Summary state. |
| DeploymentHash | string | Deployment hash. |
