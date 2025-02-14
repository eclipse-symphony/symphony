/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package queue

import "context"

type IQueueProvider interface {
	Enqueue(context context.Context, queue string, element interface{}) (string, error)
	Dequeue(context context.Context, queue string) (interface{}, error)
	Peek(context context.Context, queue string) (interface{}, error)
	Size(context context.Context, queue string) int
	RemoveFromQueue(context context.Context, queue string, messageID string) error
	QueryByPaging(context context.Context, queueName string, start string, size int) ([][]byte, string, error)
}
