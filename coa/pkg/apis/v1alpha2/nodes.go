/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

type INode interface {
	GetId() string
	GetParent() string
	GetType() string
	GetProperties() map[string]interface{}
}

type IEdge interface {
	GetFrom() string
	GetTo() string
	GetProperties() map[string]interface{}
}
