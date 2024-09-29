/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

// ActionSpecBase contains the common properties contained in all Action spec types.
type ActionSpecBase struct {
	// ParentRef is a reference to Resource on which this action is to be performed.
	ResourceRef ParentReference `json:"resourceRef,omitempty"`
}

// ActionStatusBase contains the common properties contained in all Action status types.
type ActionStatusBase struct {
	// ActionStatus contains information about result of performing the action.
	ActionStatus ActionResult `json:"actionStatus,omitempty"`
}
