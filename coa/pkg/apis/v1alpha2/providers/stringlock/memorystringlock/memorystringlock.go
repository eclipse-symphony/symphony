/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorystringlock

import (
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/stringlock"
)

type MemoryStringLockProvider struct {
	lm            *LockManager
	cleanInterval int //seconds
}

func (mslp *MemoryStringLockProvider) Init(config providers.IProviderConfig) error {
	mslp.lm = NewLockManager()
	mslp.lm.Clean()
	go func() {
		for {
			mslp.lm.Clean()
			time.Sleep(2 * time.Second)
		}
	}()
	return nil
}

func (mslp *MemoryStringLockProvider) Lock(key string) stringlock.UnLock {
	mslp.lm.GetLockNode(key).Lock()

	return func() {
		mslp.lm.GetLockNode(key).Unlock()
		go mslp.lm.UpdateLockLRU(key)
	}
}

type LockNode struct {
	lastUsedTime time.Time
	key          string
	lock         sync.Mutex
	prev         *LockNode
	next         *LockNode
}

type LockManager struct {
	purgeDuration time.Duration // 12 hours before

	m    *sync.Map
	head *LockNode
	tail *LockNode
	mu   sync.Mutex
}

func NewLockManager() *LockManager {
	return &LockManager{
		purgeDuration: -(time.Second * 30),
		m:             &sync.Map{},
		head:          nil,
		tail:          nil,
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
	locknode, _ := lm.m.LoadOrStore(key, &LockNode{key: key, prev: nil, next: nil})
	if ln, ok := locknode.(*LockNode); ok {
		return &ln.lock
	} else {
		print("unexpected behavior")
		panic("unexpected behavior")
	}
}

func (lm *LockManager) UpdateLockLRU(key string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	locknode, ok := lm.m.Load(key)

	if ok {
		locknode.(*LockNode).lastUsedTime = time.Now()

		lm.moveToHead(locknode.(*LockNode))
	}
}

func (lm *LockManager) cleanLast() bool {
	if lm.tail == nil || lm.tail.lastUsedTime.After(time.Now().Add(lm.purgeDuration)) {
		return false
	}

	node := lm.tail
	if node.prev != nil {
		node.prev.next = nil
	} else {
		lm.head = nil
	}
	lm.tail = node.prev

	lm.m.Delete(node.key)
	return true
}

func (lm *LockManager) Clean() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for lm.cleanLast() {
	}
}
