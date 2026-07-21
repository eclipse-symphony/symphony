---
type: docs
title: "Target registry route"
linkTitle: "/targets/registry"
description: ""
weight: 60
---

This page documents target registry routes.

## List targets

* **Route:** `/targets/registry`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. If omitted, returns all targets in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Target object or array of target objects.

## Create or update target

* **Route:** `/targets/registry`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
    | with-binding | Optional binding shortcut. Supported value: `staging`. | No |
* **Body:** [TargetState]({{< relref "../../models/targetstate.md" >}})
* **Response:** `200 OK` on success.

## Delete target

* **Route:** `/targets/registry`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
    | direct | If `true`, bypasses job-manager deletion path when enabled. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
