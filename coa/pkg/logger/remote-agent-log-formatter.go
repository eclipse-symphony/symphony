/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type RemoteAgentLogFormatter struct {
	logger *logrus.Logger
}

func NewRemoteAgentLogFormatter() *RemoteAgentLogFormatter {
	newLogger := logrus.New()
	newLogger.SetOutput(os.Stdout)
	newLogger.SetFormatter(&logrus.JSONFormatter{})
	return &RemoteAgentLogFormatter{
		logger: newLogger,
	}
}

func (r *RemoteAgentLogFormatter) ParseAndLogRemoteAgentMessage(operationId, message string) {
	if logEntry := r.parseStructuredLog(message); logEntry != nil {
		r.logStructuredEntry(operationId, logEntry)
		return
	}

	// Fallback to plain text logging with remote agent context
	r.logger.WithFields(logrus.Fields{
		"instance": "[REMOTE-AGENT:" + operationId + "]",
		"msg":      message,
	}).Info("")
}

type LogEntry struct {
	Instance      string
	Timestamp     string
	Level         string
	Scope         string
	Message       string
	Caller        string
	Type          string
	CorrelationId string
}

func (r *RemoteAgentLogFormatter) parseStructuredLog(message string) *LogEntry {
	// Try JSON format first
	if strings.HasPrefix(message, "{") {
		var jsonLog map[string]interface{}
		if err := json.Unmarshal([]byte(message), &jsonLog); err == nil {
			entry := &LogEntry{}
			if ts, ok := jsonLog["time"].(string); ok {
				entry.Timestamp = ts
			}
			if level, ok := jsonLog["level"].(string); ok {
				entry.Level = level
			}
			if scope, ok := jsonLog["scope"].(string); ok {
				entry.Scope = scope
			}
			if msg, ok := jsonLog["msg"].(string); ok {
				entry.Message = msg
			}
			if caller, ok := jsonLog["caller"].(string); ok {
				entry.Caller = caller
			}
			if typ, ok := jsonLog["type"].(string); ok {
				entry.Type = typ
			}
			if inst, ok := jsonLog["instance"].(string); ok {
				entry.Instance = inst
			}
			return entry
		}
	}

	// Try logrus-style: [REMOTE-AGENT:...] key="value" ...
	re := regexp.MustCompile(`^\[REMOTE-AGENT:([^\]]+)\]\s+(.*)$`)
	matches := re.FindStringSubmatch(message)
	if len(matches) == 3 {
		entry := &LogEntry{}
		entry.Instance = "[REMOTE-AGENT:" + matches[1] + "]"
		rest := matches[2]
		// Parse key="value" pairs
		kvRe := regexp.MustCompile(`(\w+)=(".*?"|\S+)`)
		kvs := kvRe.FindAllStringSubmatch(rest, -1)
		for _, kv := range kvs {
			key := kv[1]
			val := kv[2]
			val = strings.Trim(val, `"`)
			switch key {
			case "time":
				entry.Timestamp = val
			case "level":
				entry.Level = val
			case "msg":
				entry.Message = val
			case "caller":
				entry.Caller = val
			case "scope":
				entry.Scope = val
			case "type":
				entry.Type = val
			case "correlationId":
				entry.CorrelationId = val
			}
		}
		return entry
	}

	// Try text format: "timestamp level [scope] message"
	re2 := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+(\w+)\s+\[([^\]]+)\]\s+(.+)$`)
	matches2 := re2.FindStringSubmatch(message)
	if len(matches2) == 5 {
		return &LogEntry{
			Timestamp: matches2[1],
			Level:     matches2[2],
			Scope:     matches2[3],
			Message:   matches2[4],
		}
	}

	return nil
}

func (r *RemoteAgentLogFormatter) logStructuredEntry(operationId string, entry *LogEntry) {
	if entry.Instance != "" && entry.Timestamp != "" && entry.Level != "" && entry.Message != "" {
		fields := map[string]interface{}{
			"instance": entry.Instance,
			"time":     entry.Timestamp,
			"level":    entry.Level,
			"msg":      entry.Message,
		}
		if entry.Caller != "" {
			fields["caller"] = entry.Caller
		}
		if entry.Scope != "" {
			fields["scope"] = entry.Scope
		}
		if entry.Type != "" {
			fields["type"] = entry.Type
		}
		if entry.CorrelationId != "" {
			fields["correlationId"] = entry.CorrelationId
		}
		jsonBytes, _ := json.Marshal(fields)
		os.Stdout.Write(jsonBytes)
		os.Stdout.Write([]byte("\n"))

		return
	}

	// Fallback for non-JSON logs
	fields := map[string]interface{}{}
	if entry.Instance != "" {
		fields["instance"] = entry.Instance
	} else {
		fields["instance"] = "[REMOTE-AGENT:" + operationId + "]"
	}
	if entry.Timestamp != "" {
		fields["time"] = entry.Timestamp
	}
	if entry.Level != "" {
		fields["level"] = entry.Level
	}
	if entry.Message != "" {
		fields["msg"] = entry.Message
	}
	if entry.Caller != "" {
		fields["caller"] = entry.Caller
	}
	if entry.Scope != "" {
		fields["scope"] = entry.Scope
	}
	if entry.Type != "" {
		fields["type"] = entry.Type
	}

	switch strings.ToLower(entry.Level) {
	case "debug":
		r.logger.WithFields(fields).Debug("")
	case "info":
		r.logger.WithFields(fields).Info("")
	case "warn", "warning":
		r.logger.WithFields(fields).Warn("")
	case "error":
		r.logger.WithFields(fields).Error("")
	case "fatal":
		r.logger.WithFields(fields).Error("") // Don't call Fatal to avoid process exit
	default:
		r.logger.WithFields(fields).Info("")
	}
}

func (r *RemoteAgentLogFormatter) LogRemoteAgentLogs(operationId string, logs []string) {
	if len(logs) == 0 {
		return
	}

	r.logger.WithFields(logrus.Fields{
		"instance": "[REMOTE-AGENT:" + operationId + "]",
		"msg":      "Processing remote agent logs",
		"count":    len(logs),
	}).Info("")

	for _, logMessage := range logs {
		if strings.TrimSpace(logMessage) != "" {
			r.ParseAndLogRemoteAgentMessage(operationId, logMessage)
		}
	}
}
