/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package exporters

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsoleExporter(t *testing.T) {
	writer := &bytes.Buffer{}
	exporter, err := NewTraceConsoleExporter(writer)
	assert.NoError(t, err)
	assert.NotNil(t, exporter)

	// Test writer is nil case
	exporter, err = NewTraceConsoleExporter(nil)
	assert.NoError(t, err)
	assert.NotNil(t, exporter)
}

func TestNewZipkinExporter(t *testing.T) {
	url := "http://localhost:9411/api/v2/spans"
	sampleRate := "0.5"
	exporter, err := NewZipkinExporter(url, sampleRate)
	assert.NoError(t, err)
	assert.NotNil(t, exporter)
}
