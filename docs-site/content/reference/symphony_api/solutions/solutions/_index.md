---
type: docs
title: "Solution containers route"
linkTitle: "/solutions"
description: ""
weight: 60
---

This page documents solution container routes.

## List solutions

* **Route:** `/solutions`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Solution object name. If omitted, returns all solutions in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Solution object or array of solution objects.

## Create or update solution

* **Route:** `/solutions`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Solution object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [SolutionState]({{< relref "../../models/solutionstate.md" >}})
* **Response:** `200 OK` on success.

## Delete solution

* **Route:** `/solutions`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Solution object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
