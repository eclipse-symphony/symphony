/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

type Packet struct {
	SolutionVersion string      `json:"solutionversion,omitempty"`
	From     string      `json:"from"`
	To       string      `json:"to"`
	Instance string      `json:"instance,omitempty"`
	Target   string      `json:"target,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	DataType string      `json:"dataType,omitempty"`
}

func (p *Packet) IsValid() bool {
	return p.From != "" && p.To != "" && (p.SolutionVersion != "" || p.Instance != "" || p.Target != "")
}
