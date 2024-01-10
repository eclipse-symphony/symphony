/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package certs

import (
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type ICertProvider interface {
	ID() string
	Init(config providers.IProviderConfig) error
	GetCert(host string) ([]byte, []byte, error)
}
