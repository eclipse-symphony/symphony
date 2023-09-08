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
	"errors"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type SiteState struct {
	Id       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Spec     *SiteSpec         `json:"spec,omitempty"`
	Status   *SiteStatus       `json:"status,omitempty"`
}
type TargetStatus struct {
	State  v1alpha2.State `json:"state,omitempty"`
	Reason string         `json:"reason,omitempty"`
}
type InstanceStatus struct {
	State  v1alpha2.State `json:"state,omitempty"`
	Reason string         `json:"reason,omitempty"`
}

// +kubebuilder:object:generate=true
type SiteStatus struct {
	IsOnline         bool                      `json:"isOnline,omitempty"`
	TargetStatuses   map[string]TargetStatus   `json:"targetStatuses,omitempty"`
	InstanceStatuses map[string]InstanceStatus `json:"instanceStatuses,omitempty"`
	LastReported     string                    `json:"lastReported,omitempty"`
}

// +kubebuilder:object:generate=true
type SiteSpec struct {
	Name       string            `json:"name,omitempty"`
	IsSelf     bool              `json:"isSelf,omitempty"`
	PublicKey  string            `json:"secretHash,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func (s SiteSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherS, ok := other.(SiteSpec)
	if !ok {
		return false, errors.New("parameter is not a SiteSpec type")
	}

	if s.Name != otherS.Name {
		return false, nil
	}

	if s.PublicKey != otherS.PublicKey {
		return false, nil
	}

	return true, nil
}
