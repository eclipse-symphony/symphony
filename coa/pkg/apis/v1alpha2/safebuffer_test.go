/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeBuffer(t *testing.T) {
	sb := &SafeBuffer{}
	sb.Write([]byte("test"))
	assert.Equal(t, "test", sb.String())
}

func TestSafeBufferConcurrent(t *testing.T) {
	sb := &SafeBuffer{}
	sig1 := make(chan bool)
	sig2 := make(chan bool)
	go func() {
		sb.Write([]byte("test1"))
		sig1 <- true
	}()
	go func() {
		sb.Write([]byte("test2"))
		sig2 <- true
	}()

	<-sig1
	<-sig2
	assert.True(t, sb.String() == "test1test2" || sb.String() == "test2test1")
}

func TestSafeBufferReset(t *testing.T) {
	sb := &SafeBuffer{}
	sb.Write([]byte("test"))
	sb.Reset()
	assert.Equal(t, "", sb.String())
}
