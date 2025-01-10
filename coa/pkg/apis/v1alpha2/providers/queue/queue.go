/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package queue

type IQueueProvider interface {
	Enqueue(queue string, element interface{}) (string, error)
	Dequeue(queue string) (interface{}, error)
	Peek(queue string) (interface{}, error)
	Size(queue string) int
}
