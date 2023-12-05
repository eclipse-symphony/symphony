/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

type SiteInfo struct {
	SiteId      string            `json:"siteId"`
	Properties  map[string]string `json:"properties,omitempty"`
	ParentSite  SiteConnection    `json:"parentSite,omitempty"`
	CurrentSite SiteConnection    `json:"currentSite"`
}
type SiteConnection struct {
	BaseUrl  string `json:"baseUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
}
