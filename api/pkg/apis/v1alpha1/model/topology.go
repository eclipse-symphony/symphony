/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

// TopologySpec defines the desired device topology the instance
// +kubebuilder:object:generate=true
type TopologySpec struct {
	Device   string            `json:"device,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
	Bindings []BindingSpec     `json:"bindings,omitempty"`
}
