---
type: docs
title: "Campaign containers route"
linkTitle: "/campaigns"
description: ""
weight: 60
---

This page documents campaign container routes.

## List campaigns

* **Route:** `/campaigns`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign object name. If omitted, returns all campaigns in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Campaign object or array of campaign objects.

## Create or update campaign

* **Route:** `/campaigns`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [CampaignState]({{< relref "../../models/campaignstate.md" >}})
* **Response:** `200 OK` on success.

## Delete campaign

* **Route:** `/campaigns`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.

Request and response payloads use [CampaignState](../../models/campaignstate/).
