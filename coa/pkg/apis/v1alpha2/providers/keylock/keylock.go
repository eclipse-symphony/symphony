/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package keylock

import (
	"time"

	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type IKeyLockProvider interface {
	Init(config providers.IProviderConfig) error
	Lock(string)
	UnLock(string)
	TryLock(string) bool
	TryLockWithTimeout(string, time.Duration) bool
}
