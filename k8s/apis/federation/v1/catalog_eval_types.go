/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"gopls-workspace/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DefaultStopActionTimeout is default value of 1200 sec timeout period for stop action.
	DefaultStopActionTimeout int64 = 1200
)

// CatalogEvalExpressionActionSpec defines the desired state of CatalogEvalExpressionActionSpec.
type CatalogEvalExpressionSpec struct {
	// ParentRef is a reference to Resource on which this action is to be performed.
	ResourceRef ParentReference `json:"resourceRef"`
}

// CatalogEvalExpressionActionStatus defines the observed state of CatalogEvalExpressionAction.
type CatalogEvalExpressionStatus struct {
	ActionStatusBase `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CatalogEvalExpressionAction is the Schema for the CatalogEvalExpressionactions API.
type CatalogEvalExpression struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogEvalExpressionSpec   `json:"spec,omitempty"`
	Status CatalogEvalExpressionStatus `json:"status,omitempty"`
}

// GetResourceNamespacedName returns the resource reference targeted by this action.
func (a *CatalogEvalExpression) GetResourceNamespacedName() *types.NamespacedName {
	return a.Spec.ResourceRef.GetNamespacedName()
}

func (a *CatalogEvalExpression) GetOperationID() string {
	annotations := a.GetAnnotations()
	return annotations[constants.AzureOperationIdKey]
}

// +kubebuilder:object:root=true
// CatalogEvalExpressionActionList contains a list of CatalogEvalExpressionAction.
type CatalogEvalExpressionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogEvalExpression `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogEvalExpression{}, &CatalogEvalExpressionList{})
}
