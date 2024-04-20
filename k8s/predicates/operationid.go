/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package predicates

import (
	"gopls-workspace/constants"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// indicates/validates that this type is a predicate
var _ predicate.Predicate = &OperationIdPredicate{}

type OperationIdPredicate struct {
	predicate.Funcs // fills the defaults
}

// Update implements predicate.Predicate.
func (OperationIdPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		return false
	}
	if e.ObjectNew == nil {
		return false
	}
	oldAnnotations := e.ObjectOld.GetAnnotations()
	newAnnotations := e.ObjectNew.GetAnnotations()

	var oldOperationId, newOperationId string
	if oldAnnotations != nil {
		oldOperationId = oldAnnotations[constants.AzureOperationIdKey]
	}
	if newAnnotations != nil {
		newOperationId = newAnnotations[constants.AzureOperationIdKey]
	}
	return oldOperationId != newOperationId
}
