/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package providerfactory

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
)

type IProviderFactory interface {
	CreateProviders(config vendors.VendorConfig) (map[string]map[string]providers.IProvider, error)
	CreateProvider(providerType string, config providers.IProviderConfig) (providers.IProvider, error)
}
