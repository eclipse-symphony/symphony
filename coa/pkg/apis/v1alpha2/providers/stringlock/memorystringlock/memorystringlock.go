/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorystringlock

import (
	"sync"
	"time"
)

type MemoryStringLockProvider struct {
	lm LockManager
}

// Init(config providers.IProviderConfig) error
func (mslp *MemoryStringLockProvider) Lock(string) {
}
func (mslp *MemoryStringLockProvider) UnLock(string) {

}

type Lock struct {
	// Lock implementation here
}

type LockNode struct {
	lastusedtime time.Time
	key          string
	lock         sync.Mutex
	prev         *LockNode
	next         *LockNode
}

type LockManager struct {
	lockMap map[string]*LockNode

	m    *sync.Map
	head *LockNode
	tail *LockNode
	mu   sync.Mutex
}

func NewLockManager() *LockManager {
	return &LockManager{
		lockMap: make(map[string]*LockNode),
		m:       &sync.Map{},
		head:    nil,
		tail:    nil,
	}
}

func (lm *LockManager) moveToHead(node *LockNode) {
	if lm.head == node {
		return
	}
	if lm.tail == node {
		lm.tail = node.prev
	}
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	node.prev = nil
	node.next = lm.head
	if lm.head != nil {
		lm.head.prev = node
	}
	lm.head = node
	if lm.tail == nil {
		lm.tail = node
	}
}

func (lm *LockManager) GetLockNode(key string) *sync.Mutex {
	mutex, _ := lm.m.LoadOrStore(key, &LockNode{prev: nil, next: nil})
	return mutex.(*sync.Mutex)
}

func (lm *LockManager) UpdateLockLRU(key string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	node, _ := lm.m.Load(key)

	node

	if node, exists := lm.lockMap[key]; exists {
		node.lock = lock
		lm.moveToHead(node)
	} else {
		node := &ListNode{key: key, lock: lock}
		lm.lockMap[key] = node
		lm.moveToHead(node)
	}
}

func (lm *LockManager) Clean() *Lock {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.tail == nil {
		return nil
	}

	node := lm.tail
	if node.prev != nil {
		node.prev.next = nil
	} else {
		lm.head = nil
	}
	lm.tail = node.prev

	delete(lm.lockMap, node.key)
	return node.lock
}
