/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "time"

type DeployableStatus struct {
	Properties         map[string]string  `json:"properties,omitempty"`
	ProvisioningStatus ProvisioningStatus `json:"provisioningStatus"`
	LastModified       time.Time          `json:"lastModified,omitempty"`
}
