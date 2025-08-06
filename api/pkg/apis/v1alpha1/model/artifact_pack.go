/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

type ArtifactPack struct {
	SolutionContainers []SolutionContainerState `json:"solutionContainers,omitempty"`
	Solutions          []SolutionState          `json:"solutions,omitempty"`
	Campaigns          []CampaignState          `json:"campaigns,omitempty"`
	Targets            []TargetState            `json:"targets,omitempty"`
	Instances          []InstanceState          `json:"instances,omitempty"`
	Activations        []ActivationState        `json:"activations,omitempty"`
	Catalogs           []CatalogState           `json:"catalogs,omitempty"`
}
