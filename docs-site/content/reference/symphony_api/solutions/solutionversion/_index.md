---
type: docs
title: "SolutionVersion routes"
linkTitle: "/solutionversion"
description: ""
weight: 60
---

This page documents solution version deployment and reconciliation routes.

## Apply deployment through instances route

* **Route:** `/solutionversion/instances`
* **Method:** `GET | POST | DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | namespace | Namespace for request scope. | No |
    | instance | Instance name for queue/deployment operations. | No |
    | name | Instance object name in request path when provided by caller route. | No |
* **Body:** Deployment payload (for POST) or none (for GET/DELETE).
* **Response:** Deployment or queue operation result.

## Reconcile deployment state

* **Route:** `/solutionversion/reconcile`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | delete | Set to `true` to reconcile as removal. | No |
* **Body:** Deployment request payload consumed by the SolutionVersion manager routes.
* **Response:** Reconcile summary payload.

## Queue operations

* **Route:** `/solutionversion/queue`
* **Method:** `GET | POST | DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | namespace | Namespace for queue operations. | No |
    | instance | Instance identifier for queue item. | Yes |
    | objectType | `instance`, `target`, or `deployment`. | No |
    | delete | Set `true` to enqueue deletion. | No |
* **Body:** Queue job payload (for POST) or none.
* **Response:** Queue state or operation result.
