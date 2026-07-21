---
type: docs
title: "Device State"
linkTitle: "DeviceState"
description: ""
weight: 60
---

## DeviceState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the device object. |
| Spec | DeviceSpec | Desired device definition. |
| Status | DeviceStatus | Current device status. |

## DeviceSpec

| Field | Type | Description |
|--------|--------|--------|
| DisplayName | string | Human-readable name. |
| Properties | map[string]string | Device properties. |
| Bindings | []BindingSpec | Device bindings. |

## DeviceStatus

| Field | Type | Description |
|--------|--------|--------|
| Properties | map[string]string | Device status properties. |
