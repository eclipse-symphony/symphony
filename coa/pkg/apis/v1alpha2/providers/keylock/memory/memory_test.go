/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"
)

type MockProviderConfig struct {
	CleanInterval int `json:"cleanInterval"`
	PurgeDuration int `json:"purgeDuration"`
}

func (m MockProviderConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CleanInterval int `json:"cleanInterval"`
		PurgeDuration int `json:"purgeDuration"`
	}{
		CleanInterval: m.CleanInterval,
		PurgeDuration: m.PurgeDuration,
	})
}

func TestToMemoryKeyLockProviderConfig(t *testing.T) {
	mockConfig := MockProviderConfig{
		CleanInterval: 45,
		PurgeDuration: 3600,
	}
	expectedConfig := MemoryKeyLockProviderConfig{
		CleanInterval: 45,
		PurgeDuration: 3600,
	}

	data, err := json.Marshal(mockConfig)
	if err != nil {
		t.Fatalf("Failed to marshal mock config: %v", err)
	}

	var providerConfig MemoryKeyLockProviderConfig
	err = json.Unmarshal(data, &providerConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal provider config: %v", err)
	}

	if providerConfig != expectedConfig {
		t.Errorf("Expected %v, got %v", expectedConfig, providerConfig)
	}
}

func TestInit(t *testing.T) {
	mockConfig := MockProviderConfig{
		CleanInterval: 45,
		PurgeDuration: 3600,
	}
	provider := GlobalMemoryKeyLock{}
	err := provider.Init(mockConfig)
	if err != nil {
		t.Fatalf("Initialization failed: %v", err)
	}

	if provider.cleanInterval != 45 {
		t.Errorf("Expected cleanInterval to be 45, got %d", provider.cleanInterval)
	}
	if provider.purgeDuration != 3600 {
		t.Errorf("Expected purgeDuration to be 3600, got %d", provider.purgeDuration)
	}
	if provider.lm == nil {
		t.Errorf("Expected LockManager to be initialized")
	}
}

func TestInitDefaultValues(t *testing.T) {
	mockConfig := MockProviderConfig{}
	provider := GlobalMemoryKeyLock{}
	err := provider.Init(mockConfig)
	if err != nil {
		t.Fatalf("Initialization failed: %v", err)
	}

	// Check default values
	if provider.cleanInterval != 30 {
		t.Errorf("Expected default cleanInterval to be 30, got %d", provider.cleanInterval)
	}
	if provider.purgeDuration != 60*60*12 {
		t.Errorf("Expected default purgeDuration to be 43200, got %d", provider.purgeDuration)
	}
}

func TestLockAndUnlock(t *testing.T) {
	provider := GlobalMemoryKeyLock{}
	provider.lm = NewLockMap()

	provider.Lock("testKey")
	if provider.lm.getLockNode("testKey").TryLock() {
		t.Errorf("Lock should be acquired, but it's not")
	}

	// Unlock the key
	provider.UnLock("testKey")

	if !provider.lm.getLockNode("testKey").TryLock() {
		t.Errorf("Lock should be released, but it's not")
	}
}

func TestUpdateLockLRU(t *testing.T) {
	lockManager := NewLockMap()
	key1 := "key1"
	key2 := "key2"

	// Create two nodes
	lockManager.getLockNode(key1).Lock()
	lockManager.getLockNode(key2).Lock()

	// Update LRU for key1
	lockManager.updateLockLRU(key2)
	lockManager.updateLockLRU(key1)

	// Check if key1 is moved to the head
	if lockManager.head.next.key != key1 {
		t.Errorf("Expected head to be %s, got %s", key1, lockManager.head.key)
	}

	// Unlock the nodes
	lockManager.getLockNode(key1).Unlock()
	lockManager.getLockNode(key2).Unlock()
}

func TestLockManagerClean(t *testing.T) {
	lockManager := NewLockMap()

	// Create a node
	node := &LockNode{
		lastUsedTime: time.Now().Add(-time.Hour), // 1 hour ago
		key:          "testKey",
		prev:         nil,
		next:         nil,
	}

	// Add to the manager
	lockManager.m.Store(node.key, node)
	lockManager.moveToHead(node)

	// Clean the node
	lockManager.clean(-3600) // Clean nodes older than 1 hour

	// Check if node is deleted
	if _, ok := lockManager.m.Load(node.key); ok {
		t.Errorf("Node should be cleaned but it's not")
	}
}

func TestConcurrentLockUnlock(t *testing.T) {
	provider := GlobalMemoryKeyLock{}
	provider.lm = NewLockMap()

	var wg sync.WaitGroup
	numGoroutines := 100
	key := "testKey"

	wg.Add(numGoroutines)

	// Concurrently lock and unlock the same key
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			provider.Lock(key)
			time.Sleep(10 * time.Millisecond) // Simulate some work with the lock
			provider.UnLock(key)
		}()
	}

	wg.Wait()

	// After all operations, the lock should be available
	mutex := provider.lm.getLockNode(key)
	if !mutex.TryLock() {
		t.Errorf("Expected lock to be released, but it was not")
	} else {
		mutex.Unlock()
	}
}

func TestConcurrentAccessDifferentKeys(t *testing.T) {
	provider := GlobalMemoryKeyLock{}
	provider.lm = NewLockMap()

	var wg sync.WaitGroup
	numGoroutines := 100
	numKeys := 10

	wg.Add(numGoroutines * numKeys)

	// Concurrently lock and unlock different keys
	for i := 0; i < numKeys; i++ {
		key := "testKey" + strconv.Itoa(i)
		for j := 0; j < numGoroutines; j++ {
			go func(key string) {
				defer wg.Done()
				provider.Lock(key)
				time.Sleep(10 * time.Millisecond) // Simulate some work with the lock
				provider.UnLock(key)
			}(key)
		}
	}

	wg.Wait()

	// Check that all locks are released
	for i := 0; i < numKeys; i++ {
		key := "testKey" + strconv.Itoa(i)
		mutex := provider.lm.getLockNode(key)
		if !mutex.TryLock() {
			t.Errorf("Expected lock to be released for key %s, but it was not", key)
		} else {
			mutex.Unlock()
		}
	}
}

func lockRecurisive(Total int, cur int, wg *sync.WaitGroup, lockManager *LockMap) {
	key := "key" + strconv.Itoa(cur)
	mutex := lockManager.getLockNode(key)
	mutex.Lock()

	// Simulate some work
	time.Sleep(time.Duration(cur%10) * time.Millisecond)

	// Update LRU
	lockManager.updateLockLRU(key)
	mutex.Unlock()

	if cur < Total {
		go lockRecurisive(Total, cur+1, wg, lockManager)
	}
	wg.Done()
}

func TestLockManagerLRUConcurrent(t *testing.T) {
	lockManager := NewLockMap()
	numKeys := 10

	var wg sync.WaitGroup
	wg.Add(numKeys)

	// Concurrently access different keys to test LRU
	go lockRecurisive(numKeys, 1, &wg, lockManager)

	wg.Wait()

	// Check if the most recently accessed key is at the head of the LRU list
	for i := 1; i <= numKeys; i++ {
		expectedLeastRecentKey := "key" + strconv.Itoa(i)
		if lockManager.tail.prev.key != expectedLeastRecentKey {
			t.Errorf("Expected most recent key to be %s, but got %s", expectedLeastRecentKey, lockManager.tail.prev.key)
		}
		lockManager.cleanLast(0)
	}
}
