---
type: docs
title: "Catalog containers route"
linkTitle: "/catalogs"
description: ""
weight: 60
---

This page documents catalog container routes.

## List catalogs

* **Route:** `/catalogs`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog object name. If omitted, returns all catalogs in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Catalog object or array of catalog objects.

## Create or update catalog

* **Route:** `/catalogs`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [CatalogState]({{< relref "../../models/catalogstate.md" >}})
* **Response:** `200 OK` on success.

## Delete catalog

* **Route:** `/catalogs`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Catalog object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
