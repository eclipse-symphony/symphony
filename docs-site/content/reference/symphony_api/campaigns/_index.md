---
type: docs
title: "Campaigns API"
linkTitle: "Campaigns API"
description: ""
weight: 60
---

Campaigns API covers management and execution of resilient and distributed workflows.

## Pages

- [Campaign containers route](/reference/symphony_api/campaigns/campaigns/)
- [Campaign versions routes](/reference/symphony_api/campaigns/campaignversions/)
- [Activations routes](/reference/symphony_api/campaigns/activations/)

Campaigns are defined by a **Campaign** object, which consists of one or more **Stages**. Once a workflow is defined, it can be activated by creating one or more **Activation** objects. **Activation** objects also serve as execution records of the campaign, and they are periodically purged based on a **retention policy**.

Each **Stage** is handled by a **Stage Provider**. And once a provider finishes handling a stage, it runs a **Stage Selector** to select the next stage. When no next stages are selected, the activation terminates.

A **Campaign** can have multiple versions, described as **CampaignVersion** objects. A **Campaign** needs to have at least one version.

