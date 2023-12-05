/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package providers

type IProviderConfig interface {
}

type IProvider interface {
	Init(config IProviderConfig) error
}
