/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"time"

	"golang.org/x/exp/maps"

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
	if v, ok := s.TargetResults[target]; !ok {
		s.TargetResults[target] = spec
	} else {
		status := v.Status
		if spec.Status != "OK" {
			status = spec.Status
		}
		message := v.Message
		if spec.Message != "" {
			if message != "" {
				message += "; "
			}
			message += spec.Message
		}
		v.Status = status
		v.Message = message
		maps.Copy(v.ComponentResults, spec.ComponentResults)
	}
}
