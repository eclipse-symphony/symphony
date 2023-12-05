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
	stack := MemoryQueueProvider{}
	err := stack.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	stack.Enqueue("queue1", "a")
	stack.Enqueue("queue1", "b")
	stack.Enqueue("queue1", "c")
	element, err := stack.Peek("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
}
func TestPop(t *testing.T) {
	stack := MemoryQueueProvider{}
	err := stack.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	stack.Enqueue("queue1", "a")
	stack.Enqueue("queue1", "b")
	stack.Enqueue("queue1", "c")
	element, err := stack.Dequeue("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
	element, err = stack.Peek("queue1")
	assert.Nil(t, err)
	assert.Equal(t, "b", element)
}
func TestPopEmpty(t *testing.T) {
	stack := MemoryQueueProvider{}
	err := stack.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := stack.Dequeue("queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestPeekEmepty(t *testing.T) {
	stack := MemoryQueueProvider{}
	err := stack.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := stack.Peek("queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestSize(t *testing.T) {
	stack := MemoryQueueProvider{}
	err := stack.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	stack.Enqueue("queue1", "a")
	stack.Enqueue("queue1", "b")
	stack.Enqueue("queue1", "c")
	assert.Equal(t, 3, stack.Size("queue1"))
}
