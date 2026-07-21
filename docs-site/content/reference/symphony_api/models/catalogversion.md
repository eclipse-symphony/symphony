---
type: docs
title: "Catalog Version Models"
linkTitle: "CatalogVersion"
description: ""
weight: 60
---

## CatalogVersionState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the catalog version object. |
| Spec | CatalogVersionSpec | Catalog version definition. |
| Status | CatalogVersionStatus | Current catalog version status. |

## CatalogVersionSpec

| Field | Type | Description |
|--------|--------|--------|
| CatalogType | string | Catalog type. |
| Metadata | map[string]string | Additional metadata. |
| Properties | map[string]interface{} | Catalog properties. |
| ParentName | string | Parent catalog name. |
| ObjectRef | ObjectRef | Reference to the backing object. |
| Version | string | Version string. |
| RootResource | string | Root resource name. |

## ObjectRef

| Field | Type | Description |
|--------|--------|--------|
| SiteId | string | Site identifier. |
| Name | string | Object name. |
| Group | string | API group. |
| Version | string | API version. |
| Kind | string | Kubernetes kind. |
| Namespace | string | Namespace. |
| Address | string | Optional address. |
| Generation | string | Generation marker. |
| Metadata | map[string]string | Additional metadata. |

## CatalogVersionStatus

| Field | Type | Description |
|--------|--------|--------|
| Properties | map[string]string | Reported status properties. |
