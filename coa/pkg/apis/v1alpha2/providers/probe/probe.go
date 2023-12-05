/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package probe

import (
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IProbeProvider interface {
	Init(config providers.IProviderConfig) error
	Probe(user string, password string, ip string, name string) (map[string]string, error)
}
