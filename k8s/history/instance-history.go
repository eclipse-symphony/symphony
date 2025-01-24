/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package history

import (
	"context"
)

type (
	// Prototype for instance snapshot.
	SaveObjectSnapshotFunc func(ctx context.Context, objectName string, namespace string, instance interface{}) error
)

type InstanceHistory struct {
	SaveInstanceHistoryFunc SaveObjectSnapshotFunc
}

func NewInstanceHistory(saveInstanceHistoryFunc SaveObjectSnapshotFunc) InstanceHistory {
	return InstanceHistory{
		SaveInstanceHistoryFunc: saveInstanceHistoryFunc,
	}
}

// Validate Instance creation or update
// 1. DisplayName is unique
// 2. Solution exists
// 3. Target exists if provided by name rather than selector
// 4. Target is valid, i.e. either name or selector is provided

func (i *InstanceHistory) SaveInstanceHistory(ctx context.Context, objectName string, namespace string, instance interface{}) error {
	return i.SaveInstanceHistoryFunc(ctx, objectName, namespace, instance)
}
