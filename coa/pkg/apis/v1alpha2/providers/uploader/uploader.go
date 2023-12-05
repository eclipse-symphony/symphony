/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package uploader

import (
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IUploader interface {
	Init(config providers.IProviderConfig) error
	Upload(name string, data []byte) (string, error)
}
