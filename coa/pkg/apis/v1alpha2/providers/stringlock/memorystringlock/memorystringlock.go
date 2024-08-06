/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorystringlock

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/stringlock"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type MemoryStringLockProvider struct {
	lm            *LockManager
	cleanInterval int //seconds
	purgeDuration int // 12 hours before
}

type MemoryStringLockProviderConfig struct {
	CleanInterval int `json:"cleanInterval"`
	PurgeDuration int `json:"purgeDuration"`
}

func toMemoryStringLockProviderConfig(config providers.IProviderConfig) (MemoryStringLockProviderConfig, error) {
	ret := MemoryStringLockProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (mslp *MemoryStringLockProvider) Init(config providers.IProviderConfig) error {
	stringLockConfig, err := toMemoryStringLockProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (String Lock): failed to parse provider config %+v", err)
		return errors.New("expected MemoryStringLockProviderConfig")
	}
	if stringLockConfig.CleanInterval > 0 {
		mslp.cleanInterval = stringLockConfig.CleanInterval
	} else {
		mslp.cleanInterval = 30 // default: 30 seconds
	}
	if stringLockConfig.PurgeDuration > 0 {
		mslp.purgeDuration = stringLockConfig.PurgeDuration
	} else {
		mslp.purgeDuration = 60 * 60 * 12 // default: 12 hours
	}
	mslp.lm = NewLockManager()
	go func() {
		for {
			mslp.lm.clean(-mslp.purgeDuration)
			time.Sleep(time.Duration(mslp.cleanInterval) * time.Second)
		}
	}()
	return nil
}

func (mslp *MemoryStringLockProvider) Lock(key string) stringlock.UnLock {
	mslp.lm.getLockNode(key).Lock()

	return func() {
		mslp.lm.getLockNode(key).Unlock()
		go mslp.lm.updateLockLRU(key)
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
	m    *sync.Map
	head *LockNode
	tail *LockNode
	mu   sync.Mutex
}

func NewLockManager() *LockManager {
	lm := LockManager{
		m:    &sync.Map{},
		head: nil,
		tail: nil,
	}
	lm.head = &LockNode{prev: nil, next: nil} // dummyhead
	lm.tail = &LockNode{prev: nil, next: nil} // dummytail
	lm.head.next = lm.tail
	lm.tail.prev = lm.head
	return &lm
}

func (lm *LockManager) moveToHead(node *LockNode) {
	if lm.head.next == node {
		return
	}
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}

	node.prev = lm.head
	node.next = lm.head.next
	node.next.prev = node
	lm.head.next = node
}

func (lm *LockManager) getLockNode(key string) *sync.Mutex {
	locknode, _ := lm.m.LoadOrStore(key, &LockNode{key: key, prev: nil, next: nil})
	if ln, ok := locknode.(*LockNode); ok {
		return &ln.lock
	} else {
		print("unexpected behavior")
		panic("unexpected behavior")
	}
}

func (lm *LockManager) updateLockLRU(key string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	locknode, ok := lm.m.Load(key)

	if ok {
		locknode.(*LockNode).lastUsedTime = time.Now()

		lm.moveToHead(locknode.(*LockNode))
	}
}

func (lm *LockManager) cleanLast(purgeDuration int) bool {
	if lm.tail.prev == lm.head || lm.tail.prev.lastUsedTime.After(time.Now().Add(time.Duration(purgeDuration)*time.Second)) {
		return false
	}

	node := lm.tail.prev
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}

	lm.m.Delete(node.key)
	return true
}

func (lm *LockManager) clean(purgeDuration int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for lm.cleanLast(purgeDuration) {
	}
}
