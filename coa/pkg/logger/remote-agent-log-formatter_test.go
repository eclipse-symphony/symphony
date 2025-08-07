/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"bytes"
	"os"
	"testing"
)

func TestNewRemoteAgentLogFormatter(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	if formatter == nil {
		t.Error("Expected formatter to be created, got nil")
	}
}

func TestParseStructuredLog(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	// Test JSON format
	jsonLog := `{"time":"2023-01-01T12:00:00Z","level":"info","scope":"test","msg":"test message","caller":"test:123"}`
	entry := formatter.parseStructuredLog(jsonLog)

	if entry == nil {
		t.Error("Expected to parse JSON log entry")
		return
	}

	if entry.Timestamp != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected timestamp '2023-01-01T12:00:00Z', got '%s'", entry.Timestamp)
	}

	if entry.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", entry.Level)
	}

	if entry.Scope != "test" {
		t.Errorf("Expected scope 'test', got '%s'", entry.Scope)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}
}

func TestParseTextLog(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	// Test text format
	textLog := "2023-01-01T12:00:00Z INFO [test] test message"
	entry := formatter.parseStructuredLog(textLog)

	if entry == nil {
		t.Error("Expected to parse text log entry")
		return
	}

	if entry.Timestamp != "2023-01-01T12:00:00Z" {
		t.Errorf("Expected timestamp '2023-01-01T12:00:00Z', got '%s'", entry.Timestamp)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got '%s'", entry.Level)
	}

	if entry.Scope != "test" {
		t.Errorf("Expected scope 'test', got '%s'", entry.Scope)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}
}

func TestParseUnstructuredLog(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	// Test unstructured log
	unstructuredLog := "This is just a plain log message"
	entry := formatter.parseStructuredLog(unstructuredLog)

	if entry != nil {
		t.Error("Expected nil for unstructured log, but got an entry")
	}
}

func TestLogRemoteAgentLogs(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	logs := []string{
		"{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch.func1:123\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"response status: 200 OK\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:15.063634418+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch.func1:123\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"response status: 200 OK\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:15.103985338+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch:130\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"All starter requests processed. Starting polling agent.\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:15.104078942+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch.func2:138\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"Found jobs: 1. \\n\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:25.124111086+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch.func2.1:146\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"warning\",\"msg\":\"Warning: correlationId not found or not a string. Using a mock one.\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:25.124210691+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/providers.(*RemoteAgentProvider).Apply:266\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"  P (Remote Agent Provider): applying artifacts: default - target-runtime-remote-new\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:25.12457661+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/providers.(*RemoteAgentProvider).Apply:273\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"  P (Remote Agent Provider): There is no action. Report status back.\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:25.124632612+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}", "{\"caller\":\"github.com/eclipse-symphony/symphony/remote-agent/bindings/http.(*HttpBinding).Launch.func2.1:159\",\"instance\":\"cn-zhengyuhuang\",\"level\":\"info\",\"msg\":\"Agent response: {a93b3746-2b31-4455-ab28-1a7fa99677db default \\u003cnil\\u003e [123 34 114 101 109 111 116 101 45 97 103 101 110 116 34 58 123 34 115 116 97 116 117 115 34 58 50 48 48 44 34 109 101 115 115 97 103 101 34 58 34 123 92 34 99 101 114 116 105 102 105 99 97 116 101 69 120 112 105 114 97 116 105 111 110 92 34 58 92 34 50 48 50 53 45 49 49 45 48 50 84 48 54 58 48 52 58 52 54 90 92 34 44 92 34 108 97 115 116 67 111 110 110 101 99 116 101 100 92 34 58 92 34 50 48 50 53 45 48 56 45 48 54 84 48 53 58 52 54 58 50 53 90 92 34 44 92 34 115 116 97 116 101 92 34 58 92 34 97 99 116 105 118 101 92 34 44 92 34 118 101 114 115 105 111 110 92 34 58 92 34 48 46 48 46 48 46 49 92 34 125 34 125 125] []}\",\"scope\":\"remote-agent\",\"time\":\"2025-08-06T13:46:25.12477922+08:00\",\"type\":\"log\",\"ver\":\"unknown\"}",
	}

	// This should not panic and should handle all log types
	formatter.LogRemoteAgentLogs("test-operation-123", logs)

	// Test with empty logs
	formatter.LogRemoteAgentLogs("test-operation-456", []string{})

	// Test with nil logs
	formatter.LogRemoteAgentLogs("test-operation-789", nil)
}

func TestRemoteAgentLogFormatter_HookAndStdout_JSON(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	formatter.logger.SetOutput(w) // Ensure logrus output goes to the pipe

	jsonLog := `{"instance":"test-agent","time":"2025-08-04T16:59:11.152835401+08:00","level":"info","msg":"Found jobs: 1","scope":"remote-agent","type":"log"}`
	formatter.ParseAndLogRemoteAgentMessage("test-op", jsonLog)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	// Check stdout contains clean JSON
	if !bytes.Contains(buf.Bytes(), []byte(`"instance":"test-agent"`)) || !bytes.Contains(buf.Bytes(), []byte(`"msg":"Found jobs: 1"`)) {
		t.Errorf("Stdout does not contain expected JSON log: %s", buf.String())
	}
}

func TestRemoteAgentLogFormatter_HookAndStdout_NonJSON(t *testing.T) {
	formatter := NewRemoteAgentLogFormatter()

	// Capture stdout
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	formatter.logger.SetOutput(w) // Ensure logrus output goes to the pipe

	textLog := "2025-08-04T16:59:11.152835401+08:00 INFO [remote-agent] Found jobs: 2"
	formatter.ParseAndLogRemoteAgentMessage("test-op", textLog)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	// Check stdout contains expected fallback log
	if !bytes.Contains(buf.Bytes(), []byte(`"instance":"[REMOTE-AGENT:test-op]"`)) {
		t.Errorf("Stdout does not contain expected fallback log: %s", buf.String())
	}
}
