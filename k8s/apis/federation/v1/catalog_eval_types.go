/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DefaultStopActionTimeout is default value of 1200 sec timeout period for stop action.
	DefaultStopActionTimeout int64 = 1200
)

// VirtualMachineStopActionSpec defines the desired state of VirtualMachineStopActionSpec.
type CatalogEvalExpressionSpec struct {
	ActionSpecBase `json:",inline"`
}

// VirtualMachineStopActionStatus defines the observed state of VirtualMachineStopAction.
type CatalogEvalExpressionStatus struct {
	ActionStatusBase `json:",inline"`
}

// +kubebuilder:object:root=true
// VirtualMachineStopAction is the Schema for the virtualmachinestopactions API.
type CatalogEvalExpression struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEvalExpressionSpec   `json:"spec,omitempty"`
	Status CatalogEvalExpressionStatus `json:"status,omitempty"`
}

// GetResourceNamespacedName returns the resource reference targeted by this action.
func (a *CatalogEvalExpression) GetResourceNamespacedName() *types.NamespacedName {
	return a.Spec.ParentRef.GetNamespacedName()
}

// +kubebuilder:object:root=true
// VirtualMachineStopActionList contains a list of VirtualMachineStopAction.
type CatalogEvalExpressionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEvalExpression `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEvalExpression{}, &CatalogEvalExpressionList{})
}
