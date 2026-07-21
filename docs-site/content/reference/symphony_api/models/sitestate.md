---
type: docs
title: "Site State"
linkTitle: "SiteState"
description: ""
weight: 60
---

## SiteState

| Field | Type | Description |
|--------|--------|--------|
| Id | string | Site identifier. |
| Metadata | ObjectMeta | Metadata for the site object. |
| Spec | SiteSpec | Desired site definition. |
| Status | SiteStatus | Current site status. |

## SiteSpec

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Site name. |
| IsSelf | bool | Marks the current site. |
| PublicKey | string | Site secret hash / public key field. |
| Properties | map[string]string | Site properties. |

## SiteStatus

| Field | Type | Description |
|--------|--------|--------|
| IsOnline | bool | Whether the site is online. |
| TargetStatuses | map[string]SiteTargetStatus | Per-target status map. |
| InstanceStatuses | map[string]SiteInstanceStatus | Per-instance status map. |
| LastReported | string | Last report timestamp. |
