/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package reporter

import (
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IReporter interface {
	Init(config providers.IProviderConfig) error
	Report(id string, namespace string, group string, kind string, version string, properties map[string]string, overwrite bool) error
}
