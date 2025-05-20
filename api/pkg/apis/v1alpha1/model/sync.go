/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"

type SyncPackage struct {
	Origin   string             `json:"origin,omitempty"`
	Catalogs []CatalogState     `json:"catalogs,omitempty"`
	Jobs     []v1alpha2.JobData `json:"jobs,omitempty"`
}
