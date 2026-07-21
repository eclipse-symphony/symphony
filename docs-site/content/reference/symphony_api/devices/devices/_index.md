---
type: docs
title: "Device containers route"
linkTitle: "/devices"
description: ""
weight: 60
---

This page documents device container routes.

## List devices

* **Route:** `/devices`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Device object name. If omitted, returns all devices in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Device object or array of device objects.

## Create or update device

* **Route:** `/devices`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Device object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [DeviceState]({{< relref "../../models/devicestate.md" >}})
* **Response:** `200 OK` on success.

## Delete device

* **Route:** `/devices`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Device object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
