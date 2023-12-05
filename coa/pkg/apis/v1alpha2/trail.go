/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

type Trail struct {
	Origin     string                 `json:"origin"`
	Catalog    string                 `json:"catalog"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}
