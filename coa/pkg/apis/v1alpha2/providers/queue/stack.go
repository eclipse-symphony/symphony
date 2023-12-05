/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package queue

type IQueueProvider interface {
	Enqueue(stack string, element interface{}) error
	Dequeue(stack string) (interface{}, error)
	Peek(stack string) (interface{}, error)
	Size(stack string) int
}
