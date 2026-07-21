---
type: docs
title: "Activations routes"
linkTitle: "/activations"
description: ""
weight: 60
---

This page documents activation routes used by campaigns.

## List activations

* **Route:** `/activations/registry`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Activation object name. If omitted, returns all activations in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Activation object or array of activation objects.

## Create or update activation

* **Route:** `/activations/registry`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Activation object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [ActivationState]({{< relref "../../models/activationstate.md" >}})
* **Response:** `200 OK` on success.

## Delete activation

* **Route:** `/activations/registry`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Activation object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.

## Report activation status

* **Route:** `/activations/status`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Activation object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** [ActivationState]({{< relref "../../models/activationstate.md" >}}) section `ActivationStatus`
* **Response:** `200 OK` on success.
