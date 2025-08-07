/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

type TargetProviderGetReference struct {
	TargetNamespace string          `json:"targetNamespace,omitempty"`
	TargetName      string          `json:"targetName,omitempty"`
	Deployment      DeploymentSpec  `json:"deployment,omitempty"`
	References      []ComponentStep `json:"references,omitempty"`
}

type TargetProviderApplyReference struct {
	TargetNamespace string         `json:"targetNamespace,omitempty"`
	TargetName      string         `json:"targetName,omitempty"`
	Deployment      DeploymentSpec `json:"deployment,omitempty"`
	Step            DeploymentStep `json:"step,omitempty"`
	IsDryRun        bool           `json:"isDryRun,omitempty"`
}
