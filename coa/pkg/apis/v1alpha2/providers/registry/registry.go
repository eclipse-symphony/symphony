/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package registry

import (
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IRegistryProvider interface {
	ID() string
	Init(config providers.IProviderConfig) error
}
