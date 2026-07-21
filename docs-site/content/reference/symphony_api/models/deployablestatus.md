---
type: docs
title: "Deployable Status"
linkTitle: "DeployableStatus"
description: ""
weight: 60
---

## DeployableStatus

| Field | Type | Description |
|--------|--------|--------|
| Properties | map[string]string | Status properties. |
| ProvisioningStatus | ProvisioningStatus | Provisioning status. |
| LastModified | time.Time | Last modification time. |

## DeployableStatusV2

| Field | Type | Description |
|--------|--------|--------|
| ProvisioningStatus | ProvisioningStatus | Provisioning status. |
| LastModified | time.Time | Last modification time. |
| Deployed | int | Number deployed. |
| Targets | int | Number of targets. |
| Status | string | Overall status. |
| StatusDetails | string | Additional status details. |
| RunningJobId | int | Running job identifier. |
| ExpectedRunningJobId | int | Expected job identifier. |
| Generation | int | Generation marker. |
| TargetStatuses | []TargetDeployableStatus | Per-target status list. |
| Properties | map[string]string | Status properties. |

## TargetDeployableStatus

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Target name. |
| Status | string | Target status. |
| ComponentStatuses | []ComponentDeployableStatus | Component statuses. |

## ComponentDeployableStatus

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Component name. |
| Status | string | Component status. |
