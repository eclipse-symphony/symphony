---
type: docs
title: "Campaign Version Models"
linkTitle: "CampaignVersion"
description: ""
weight: 60
---

## CampaignVersionState

| Field | Type | Description |
|--------|--------|--------|
| Metadata | ObjectMeta | Metadata for the campaign version object. |
| Spec | CampaignVersionSpec | Campaign version definition. |

## CampaignVersionSpec

| Field | Type | Description |
|--------|--------|--------|
| FirstStage | string | First stage to run. |
| Stages | map[string]StageSpec | Stage definitions keyed by name. |
| SelfDriving | bool | Enables self-driving campaign behavior. |
| Version | string | Version string. |
| RootResource | string | Root resource name. |

## StageSpec

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Stage name. |
| Contexts | string | Context names. |
| Provider | string | Stage provider. |
| Config | interface{} | Provider-specific config. |
| StageSelector | string | Selector used to pick the next stage. |
| Inputs | map[string]interface{} | Stage inputs. |
| HandleErrors | bool | Marks the stage as an error handler. |
| Schedule | string | RFC3339 schedule timestamp. |
| Proxy | ProxySpec | Optional proxy settings. |
| Target | string | Target name. |
| Tasks | []TaskSpec | Tasks executed by the stage. |
| TaskOption | TaskOption | Task execution options. |

## TaskSpec

| Field | Type | Description |
|--------|--------|--------|
| Name | string | Task name. |
| Provider | string | Task provider. |
| Config | interface{} | Task config. |
| Inputs | map[string]interface{} | Task inputs. |
| Target | string | Target name. |

## TaskOption

| Field | Type | Description |
|--------|--------|--------|
| Concurrency | int | Parallel task limit. |
| ErrorAction | ErrorAction | Error handling behavior. |

## ErrorAction

| Field | Type | Description |
|--------|--------|--------|
| Mode | ErrorActionMode | Failure handling mode. |
| MaxToleratedFailures | int | Allowed failures before stopping. |
