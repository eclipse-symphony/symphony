/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package secret

import (
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ISecretProvider interface {
	Init(config providers.IProviderConfig) error
	Get(object string, field string) (string, error)
}
