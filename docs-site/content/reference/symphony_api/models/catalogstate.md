---
type: docs
title: "Catalog State"
linkTitle: "CatalogState"
description: ""
weight: 60
---

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the catalog object. |
| Spec | CatalogSpec | Desired catalog definition. |
| Status | CatalogStatus | Current catalog status. |

## CatalogSpec

The catalog spec is currently empty in the source tree.

## CatalogStatus

| Field | Type | Description |
|--------|--------|--------|
| Properties | map[string]string | Catalog status properties. |
