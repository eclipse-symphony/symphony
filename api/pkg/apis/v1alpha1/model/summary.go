/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
	IsRemoval      bool                        `json:"isRemoval"`
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
