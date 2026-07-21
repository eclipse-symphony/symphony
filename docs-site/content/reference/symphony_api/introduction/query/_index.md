---
type: docs
title: "Object query"
linkTitle: "Object query"
description: ""
weight: 60
---

When retrieving objects through a GET request, Symphony API provides several query capabilities to help you filter and format results more efficiently.

## Select Between JSON and YAML

The Symphony REST API returns objects in JSON format by default. To receive responses in YAML format, include the `doc-type=yaml` query parameter in your request.

## JSONPath Queries

Use the path query parameter to filter the response with a JSONPath expression. For example, path=$.spec returns only the specification object contained within a state object.

## Query by Namespace and Name

You can use the `namespace` parameter to limit results to a specific namespace. If no namespace is provided, the query searches across objects in all namespaces. You can also specify a `name` parameter to retrieve a particular object directly. When a matching object is found, the API returns that object rather than a collection of results.