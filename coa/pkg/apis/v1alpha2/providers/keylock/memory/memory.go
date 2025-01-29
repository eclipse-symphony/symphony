/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type MemoryKeyLock struct {
	lm            *LockMap
	cleanInterval int //seconds
	purgeDuration int // 12 hours before
}

var globalMemoryKeyLock *MemoryKeyLock
var initLock = sync.Mutex{}

func (gml *MemoryKeyLock) Lock(key string) {
	gml.lm.getLockNode(key).Lock()
}

func (gml *MemoryKeyLock) UnLock(key string) {
	gml.lm.getLockNode(key).Unlock()
	go gml.lm.updateLockLRU(key)
}

type MemoryKeyLockProvider struct {
	memKeyLockInstance *MemoryKeyLock
}

func toMemoryKeyLockProviderConfig(config providers.IProviderConfig) (MemoryKeyLockProviderConfig, error) {
	ret := MemoryKeyLockProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (gml *MemoryKeyLock) Init(KeyLockConfig MemoryKeyLockProviderConfig) error {
	sLog.Info("Init MemoryKeyLock")
	if KeyLockConfig.CleanInterval > 0 {
		gml.cleanInterval = KeyLockConfig.CleanInterval
	} else {
		gml.cleanInterval = 30 // default: 30 seconds
	}
	if KeyLockConfig.PurgeDuration > 0 {
		gml.purgeDuration = KeyLockConfig.PurgeDuration
	} else {
		gml.purgeDuration = 60 * 60 * 12 // default: 12 hours
	}
	gml.lm = NewLockMap()
	go func() {
		for {
			gml.lm.clean(-gml.purgeDuration)
			time.Sleep(time.Duration(gml.cleanInterval) * time.Second)
		}
	}()
	return nil
}

func (mslp *MemoryKeyLockProvider) Init(config providers.IProviderConfig) error {
	KeyLockConfig, err := toMemoryKeyLockProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (String Lock): failed to parse provider config %+v", err)
		return errors.New("expected MemoryKeyLockProviderConfig")
	}

	if KeyLockConfig.Mode == Global {
		sLog.Info("Trying to init global memoryKeyLock")
		initLock.Lock()
		defer initLock.Unlock()
		if globalMemoryKeyLock == nil {
			globalMemoryKeyLock = &MemoryKeyLock{}
			err = globalMemoryKeyLock.Init(KeyLockConfig)
			mslp.memKeyLockInstance = globalMemoryKeyLock
		}
	} else if KeyLockConfig.Mode == Shared {
		sLog.Info("Trying to init shared memoryKeyLock")
		initLock.Lock()
		defer initLock.Unlock()
		if globalMemoryKeyLock == nil {
			err = errors.New("A global MemoryKeyLock instance should be initialized before using a shared mode MemoryKeyLock")
		} else {
			mslp.memKeyLockInstance = globalMemoryKeyLock
		}
	} else if KeyLockConfig.Mode == Dedicated {
		sLog.Info("Trying to init dedicated memoryKeyLock")
		mslp.memKeyLockInstance = &MemoryKeyLock{}
		err = mslp.memKeyLockInstance.Init(KeyLockConfig)
	} else {
		err = errors.New("MemoryKeyLockProvider: unknown init mode")
	}

	return err
}

func (mslp *MemoryKeyLockProvider) Lock(key string) {
	mslp.memKeyLockInstance.lm.getLockNode(key).Lock()
}

func (mslp *MemoryKeyLockProvider) UnLock(key string) {
	mslp.memKeyLockInstance.lm.getLockNode(key).Unlock()
	go mslp.memKeyLockInstance.lm.updateLockLRU(key)
}

func (mslp *MemoryKeyLockProvider) TryLock(key string) bool {
	return mslp.memKeyLockInstance.lm.getLockNode(key).TryLock()
}

func (mslp *MemoryKeyLockProvider) TryLockWithTimeout(key string, duration time.Duration) bool {
	start := time.Now()
	for start.Add(duration).After(time.Now()) {
		if mslp.TryLock(key) {
			return true
		} else {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}
	return false
}

type LockNode struct {
	lastUsedTime time.Time
	key          string
	lock         sync.Mutex
	prev         *LockNode
	next         *LockNode
}

type LockMap struct {
	m    *sync.Map
	head *LockNode
	tail *LockNode
	mu   sync.Mutex
}

func NewLockMap() *LockMap {
	lm := LockMap{
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

func (lm *LockMap) moveToHead(node *LockNode) {
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

func (lm *LockMap) getLockNode(key string) *sync.Mutex {
	locknode, _ := lm.m.LoadOrStore(key, &LockNode{key: key, prev: nil, next: nil})
	if ln, ok := locknode.(*LockNode); ok {
		return &ln.lock
	} else {
		print("unexpected behavior")
		panic("unexpected behavior")
	}
}

func (lm *LockMap) updateLockLRU(key string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	locknode, ok := lm.m.Load(key)

	if ok {
		locknode.(*LockNode).lastUsedTime = time.Now()

		lm.moveToHead(locknode.(*LockNode))
	}
}

func (lm *LockMap) cleanLast(purgeDuration int) bool {
	if lm.tail.prev == lm.head || lm.tail.prev.lastUsedTime.After(time.Now().Add(time.Duration(purgeDuration)*time.Second)) {
		return false
	}

	node := lm.tail.prev
	node.prev.next = node.next
	node.next.prev = node.prev

	lm.m.Delete(node.key)
	return true
}

func (lm *LockMap) clean(purgeDuration int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for lm.cleanLast(purgeDuration) {
	}
}
