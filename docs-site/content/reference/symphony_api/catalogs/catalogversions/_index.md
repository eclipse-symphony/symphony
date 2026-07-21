---
type: docs
title: "Catalog versions routes"
linkTitle: "/catalogversions"
description: ""
weight: 60
---

This page documents catalog version routes.

## List catalog versions

* **Route:** `/catalogversions/registry`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog version object name. If omitted, returns all catalog versions in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Catalog version object or array of catalog version objects.

## Create or update catalog version

* **Route:** `/catalogversions/registry`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog version object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** [CatalogVersionState]({{< relref "../../models/catalogversion.md" >}})
* **Response:** `200 OK` on success.

## Delete catalog version

* **Route:** `/catalogversions/registry`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog version object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.

## Query catalog version graph

* **Route:** `/catalogversions/graph`
* **Method:** `GET`
* **Parameters:** None.
* **Body:** None.
* **Response:** Catalog version dependency graph payload.

## Validate catalog version payload

* **Route:** `/catalogversions/check`
* **Method:** `POST`
* **Parameters:** None.
* **Body:** [CatalogVersionState]({{< relref "../../models/catalogversion.md" >}})
* **Response:** `200 OK` when valid, `400 Bad Request` with validation details when invalid.

## Report catalog version status

* **Route:** `/catalogversions/status`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog version object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** `[]ComponentSpec` payload for reported component state.
* **Response:** `200 OK` on success.
