/*
Copyright 2022 The COA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
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
	TargetCount    int                         `json:"targetCount"`
	SuccessCount   int                         `json:"successCount"`
	TargetResults  map[string]TargetResultSpec `json:"targets,omitempty"`
	SummaryMessage string                      `json:"message,omitempty"`
	Skipped        bool                        `json:"skipped"`
}
type SummaryResult struct {
	Summary    SummarySpec `json:"summary"`
	Generation string      `json:"generation"`
	Time       time.Time   `json:"time"`
}

func (s *SummarySpec) UpdateTargetResult(target string, spec TargetResultSpec) {
	s.TargetResults[target] = spec
	count := 0
	for _, r := range s.TargetResults {
		if r.Status == "OK" {
			count++
		}
	}
	s.SuccessCount = count
}
