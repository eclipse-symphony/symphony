---
type: docs
title: "Target operational routes"
linkTitle: "Operations"
description: ""
weight: 60
---

This page documents target operational routes.

## Bootstrap target auth

* **Route:** `/targets/bootstrap`
* **Method:** `POST`
* **Parameters:** None.
* **Body:** `AuthRequest` payload (`userName`, `password`)
* **Response:** Access token payload when credentials are accepted.

## Report target heartbeat

* **Route:** `/targets/ping`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` with empty JSON object.

## Report target status

* **Route:** `/targets/status`
* **Method:** `PUT`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
    | component | Optional component name context. | No |
* **Body:** Status payload. The handler reads `status.properties` and combines it with non-internal request parameters.
* **Response:** Updated target state payload.

## Download target document

* **Route:** `/targets/download`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Target object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
    | doc-type | Output document type (`yaml` for text output, otherwise JSON). | Yes |
    | path | Optional field path selection. | No |
* **Body:** None.
* **Response:** Target document rendered in requested format.
