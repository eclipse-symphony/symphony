---
type: docs
title: "Campaign versions routes"
linkTitle: "/campaignversions"
description: ""
weight: 60
---

This page documents campaign version routes.

## List campaign versions

* **Route:** `/campaignversions`
* **Method:** `GET`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign version object name. If omitted, returns all campaign versions in the selected namespace(s). | No |
    | namespace | Namespace to query. If omitted, the API lists across namespaces. | No |
* **Body:** None.
* **Response:** Campaign version object or array of campaign version objects.

## Create or update campaign version

* **Route:** `/campaignversions`
* **Method:** `POST`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign version object name. | Yes |
    | namespace | Namespace the object belongs to. Defaults to `default` if omitted by caller route handling. | No |
* **Body:** [CampaignVersionState]({{< relref "../../models/campaignversion.md" >}})
* **Response:** `200 OK` on success.

## Delete campaign version

* **Route:** `/campaignversions`
* **Method:** `DELETE`
* **Parameters:**
    | Parameter | Description | Required |
    |--------|--------|--------|
    | name | Campaign version object name. | Yes |
    | namespace | Namespace the object belongs to. | No |
* **Body:** None.
* **Response:** `200 OK` on success.
