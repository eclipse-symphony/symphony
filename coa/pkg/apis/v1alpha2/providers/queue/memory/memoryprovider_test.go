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
	queue.Enqueue(context.TODO(), "queue1", "a")
	queue.Enqueue(context.TODO(), "queue1", "b")
	queue.Enqueue(context.TODO(), "queue1", "c")
	element, err := queue.Peek(context.TODO(), "queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
}

func TestPop(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue(context.TODO(), "queue1", "a")
	queue.Enqueue(context.TODO(), "queue1", "b")
	queue.Enqueue(context.TODO(), "queue1", "c")
	element, err := queue.Dequeue(context.TODO(), "queue1")
	assert.Nil(t, err)
	assert.Equal(t, "a", element)
	element, err = queue.Peek(context.TODO(), "queue1")
	assert.Nil(t, err)
	assert.Equal(t, "b", element)
}
func TestPopEmpty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Dequeue(context.TODO(), "queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestPeekEmepty(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	element, err := queue.Peek(context.TODO(), "queue1")
	assert.NotNil(t, err)
	assert.Equal(t, nil, element)
}
func TestSize(t *testing.T) {
	queue := MemoryQueueProvider{}
	err := queue.Init(MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	queue.Enqueue(context.TODO(), "queue1", "a")
	queue.Enqueue(context.TODO(), "queue1", "b")
	queue.Enqueue(context.TODO(), "queue1", "c")
	assert.Equal(t, 3, queue.Size(context.TODO(), "queue1"))
}
