/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memoryqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a")
	queue.Enqueue("queue1", "b")
	queue.Enqueue("queue1", "c")
	element, err := queue.Peek("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
}
func TestPop(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a")
	queue.Enqueue("queue1", "b")
	queue.Enqueue("queue1", "c")
	element, err := queue.Dequeue("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
	element, err = queue.Peek("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "b", element)
}
func TestPopEmpty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Dequeue("queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestPeekEmepty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Peek("queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestSize(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue("queue1", "a")
	queue.Enqueue("queue1", "b")
	queue.Enqueue("queue1", "c")
	assert.Equal(t, 3, queue.Size("queue1"))
}
