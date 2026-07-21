---
type: docs
title: "Activation Models"
linkTitle: "ActivationState"
description: ""
weight: 60
---

## ActivationState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the activation object. |
| Spec | ActivationSpec | Activation request definition. |
| Status | ActivationStatus | Current activation status. |

## ActivationSpec

| Field | Type | Description |
|--------|--------|--------|
| CampaignVersion | string | Campaign version to activate. |
| Stage | string | Stage to start from. |
| Inputs | map[string]interface{} | Activation inputs. |

## ActivationStatus

| Field | Type | Description |
|--------|--------|--------|
| ActivationGeneration | string | Activation generation marker. |
| UpdateTime | string | Last update time. |
| Status | State | Current state. |
| StatusMessage | string | Human-readable status. |
| StageHistory | []StageStatus | Execution history for stages. |

## StageStatus

| Field | Type | Description |
|--------|--------|--------|
| Stage | string | Current stage name. |
| NextStage | string | Next selected stage name. |
| Inputs | map[string]interface{} | Stage inputs. |
| Outputs | map[string]interface{} | Stage outputs. |
| Status | State | Stage execution state. |
| IsActive | bool | Whether this stage is currently active. |
| StatusMessage | string | Human-readable stage status. |
| ErrorMessage | string | Stage error details, when present. |
