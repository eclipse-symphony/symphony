/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"bytes"
	"sync"
)

type SafeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (sb *SafeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.buffer.Write(p)
}

func (sb *SafeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.buffer.String()
}

func (sb *SafeBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.buffer.Reset()
}
