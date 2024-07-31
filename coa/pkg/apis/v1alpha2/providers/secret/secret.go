/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package secret

import (
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ISecretProvider interface {
	Init(config providers.IProviderConfig) error
	Read(name string, field string, localContext interface{}) (string, error)
}

type IExtSecretProvider interface {
	Get(name string, field string, localContext interface{}) (string, error)
}
