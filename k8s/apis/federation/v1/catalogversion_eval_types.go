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

// CatalogVersionEvalExpressionActionSpec defines the desired state of CatalogVersionEvalExpressionActionSpec.
type CatalogVersionEvalExpressionSpec struct {
	// ParentRef is a reference to Resource on which this action is to be performed.
	ResourceRef ParentReference `json:"resourceRef"`
}

// CatalogVersionEvalExpressionActionStatus defines the observed state of CatalogVersionEvalExpressionAction.
type CatalogVersionEvalExpressionStatus struct {
	ActionStatusBase `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CatalogVersionEvalExpressionAction is the Schema for the CatalogVersionEvalExpressionactions API.
type CatalogVersionEvalExpression struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogVersionEvalExpressionSpec   `json:"spec,omitempty"`
	Status CatalogVersionEvalExpressionStatus `json:"status,omitempty"`
}

// GetResourceNamespacedName returns the resource reference targeted by this action.
func (a *CatalogVersionEvalExpression) GetResourceNamespacedName() *types.NamespacedName {
	return a.Spec.ResourceRef.GetNamespacedName()
}

func (a *CatalogVersionEvalExpression) GetOperationID() string {
	annotations := a.GetAnnotations()
	return annotations[constants.AzureOperationIdKey]
}

// +kubebuilder:object:root=true
// CatalogVersionEvalExpressionActionList contains a list of CatalogVersionEvalExpressionAction.
type CatalogVersionEvalExpressionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogVersionEvalExpression `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogVersionEvalExpression{}, &CatalogVersionEvalExpressionList{})
}
