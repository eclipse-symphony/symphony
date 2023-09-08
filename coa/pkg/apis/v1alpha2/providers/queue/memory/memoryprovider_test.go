/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
