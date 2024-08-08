/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stringlock

import (
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	//"encoding/json"
)

type UnLock func()

type IStringLockProvider interface {
	Init(config providers.IProviderConfig) error
	Lock(string) UnLock
}
