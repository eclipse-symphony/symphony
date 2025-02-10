/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memoryqueue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.InitWithMap(map[string]string{
		"name": "test",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", queue.ID())
}
func TestPush(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a", context.TODO())
	queue.Enqueue("queue1", "b", context.TODO())
	queue.Enqueue("queue1", "c", context.TODO())
	element, err := queue.Peek("queue1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
}

func TestPop(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a", context.TODO())
	queue.Enqueue("queue1", "b", context.TODO())
	queue.Enqueue("queue1", "c", context.TODO())
	element, err := queue.Dequeue("queue1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
	element, err = queue.Peek("queue1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "b", element)
}
func TestPopEmpty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Dequeue("queue1", context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestPeekEmepty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Peek("queue1", context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestSize(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a", context.TODO())
	queue.Enqueue("queue1", "b", context.TODO())
	queue.Enqueue("queue1", "c", context.TODO())
	assert.Equal(t, 3, queue.Size("queue1", context.TODO()))
}
