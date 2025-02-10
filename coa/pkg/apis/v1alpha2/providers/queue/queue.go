/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package queue

import "context"

type IQueueProvider interface {
	Enqueue(queue string, element interface{}, context context.Context) (string, error)
	Dequeue(queue string, context context.Context) (interface{}, error)
	Peek(queue string, context context.Context) (interface{}, error)
	Size(queue string, context context.Context) int
	RemoveFromQueue(queue string, messageID string, context context.Context) error
	QueryByPaging(queueName string, start string, size int, context context.Context) ([][]byte, string, error)
}
