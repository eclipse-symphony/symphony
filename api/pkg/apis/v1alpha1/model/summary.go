/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type ComponentResultSpec struct {
	Status  v1alpha2.State `json:"status"`
	Message string         `json:"message"`
}
type TargetResultSpec struct {
	Status           string                         `json:"status"`
	Message          string                         `json:"message,omitempty"`
	ComponentResults map[string]ComponentResultSpec `json:"components,omitempty"`
}
type SummarySpec struct {
	TargetCount         int                         `json:"targetCount"`
	SuccessCount        int                         `json:"successCount"`
	TargetResults       map[string]TargetResultSpec `json:"targets,omitempty"`
	SummaryMessage      string                      `json:"message,omitempty"`
	Skipped             bool                        `json:"skipped"`
	IsRemoval           bool                        `json:"isRemoval"`
	AllAssignedDeployed bool                        `json:"allAssignedDeployed"`
}
type SummaryResult struct {
	Summary    SummarySpec `json:"summary"`
	Generation string      `json:"generation"`
	Time       time.Time   `json:"time"`
}

func (s *SummarySpec) UpdateTargetResult(target string, spec TargetResultSpec) {
	s.TargetResults[target] = spec
}
