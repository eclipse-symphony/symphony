---
type: docs
title: "Instance containers route"
linkTitle: "/instances"
description: ""
weight: 60
---

This page documents instance container routes.

## List instances

* **Route:** `/instances`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Instance object name. If omitted, returns all instances in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Instance object or array of instance objects.

## Create or update instance

* **Route:** `/instances`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Instance object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
    | solutionversion | Shortcut parameter for route-based instance creation. | No |
    | target | Target name shortcut parameter for route-based instance creation. | No |
    | target-selector | Selector shortcut parameter in `<key>=<value>` format. | No |
* **Body:** [InstanceState]({{< relref "../../models/instancestate.md" >}})
* **Response:** `200 OK` on success.

## Delete instance

* **Route:** `/instances`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Instance object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
    | direct | If `true`, bypasses job-manager deletion path when enabled. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
