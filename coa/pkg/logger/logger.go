/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"context"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
)

const (
	// LogTypeLog is normal log type
	LogTypeLog = "log"
	// LogTypeRequest is Request log type
	LogTypeRequest = "request"
	// LogTypeUserAudits is User Audit log type
	LogTypeUserAudits = "userAudits"
	// LogTypeUserDiagnostics is User Diagnostic log type
	LogTypeUserDiagnostics = "userDiagnostics"

	// Field names that defines Dapr log schema
	logFieldTimeStamp = "time"
	logFieldLevel     = "level"
	logFieldType      = "type"
	logFieldScope     = "scope"
	logFieldMessage   = "msg"
	logFieldInstance  = "instance"
	logFieldDaprVer   = "ver"
	logFieldAppID     = "app_id"
)

// LogLevel is Dapr Logger Level type
type LogLevel string

const (
	// DebugLevel has verbose message
	DebugLevel LogLevel = "debug"
	// InfoLevel is default log level
	InfoLevel LogLevel = "info"
	// WarnLevel is for logging messages about possible issues
	WarnLevel LogLevel = "warn"
	// ErrorLevel is for logging errors
	ErrorLevel LogLevel = "error"
	// FatalLevel is for logging fatal messages. The system shuts down after logging the message.
	FatalLevel LogLevel = "fatal"

	// UndefinedLevel is for undefined log level
	UndefinedLevel LogLevel = "undefined"
)

// globalLoggers is the collection of Dapr Logger that is shared globally.
// TODO: User will disable or enable logger on demand.
var globalLoggers = map[string]Logger{}
var globalLoggersLock = sync.RWMutex{}
var globalUserAuditsLoggerOnce sync.Once
var globalUserAuditsLogger Logger
var globalUserDiagnosticsLoggerOnce sync.Once
var globalUserDiagnosticsLogger Logger

// Logger includes the logging api sets
type Logger interface {
	// EnableJSONOutput enables JSON formatted output log
	EnableJSONOutput(enabled bool)

	// SetAppID sets dapr_id field in the log. Default value is empty string
	SetAppID(id string)
	// SetOutputLevel sets log output level
	SetOutputLevel(outputLevel LogLevel)
	// WithLogType specify the log_type field in log. Default value is LogTypeLog
	WithLogType(logType string) Logger

	// Info logs a message at level Info.
	InfoCtx(ctx context.Context, args ...interface{})
	// Infof logs a message at level Info.
	InfofCtx(ctx context.Context, format string, args ...interface{})
	// Debug logs a message at level Debug.
	DebugCtx(ctx context.Context, args ...interface{})
	// Debugf logs a message at level Debug.
	DebugfCtx(ctx context.Context, format string, args ...interface{})
	// Warn logs a message at level Warn.
	WarnCtx(ctx context.Context, args ...interface{})
	// Warnf logs a message at level Warn.
	WarnfCtx(ctx context.Context, format string, args ...interface{})
	// Error logs a message at level Error.
	ErrorCtx(ctx context.Context, args ...interface{})
	// Errorf logs a message at level Error.
	ErrorfCtx(ctx context.Context, format string, args ...interface{})
	// Fatal logs a message at level Fatal then the process will exit with status set to 1.
	FatalCtx(ctx context.Context, args ...interface{})
	// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
	FatalfCtx(ctx context.Context, format string, args ...interface{})

	// Info logs a message at level Info.
	Info(args ...interface{})
	// Infof logs a message at level Info.
	Infof(format string, args ...interface{})
	// Debug logs a message at level Debug.
	Debug(args ...interface{})
	// Debugf logs a message at level Debug.
	Debugf(format string, args ...interface{})
	// Warn logs a message at level Warn.
	Warn(args ...interface{})
	// Warnf logs a message at level Warn.
	Warnf(format string, args ...interface{})
	// Error logs a message at level Error.
	Error(args ...interface{})
	// Errorf logs a message at level Error.
	Errorf(format string, args ...interface{})
	// Fatal logs a message at level Fatal then the process will exit with status set to 1.
	Fatal(args ...interface{})
	// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
	Fatalf(format string, args ...interface{})
}

// toLogLevel converts to LogLevel
func toLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	}

	// unsupported log level by Dapr
	return UndefinedLevel
}

// NewLogger creates new Logger instance.
func NewLogger(name string) Logger {
	globalLoggersLock.Lock()
	defer globalLoggersLock.Unlock()

	logger, ok := globalLoggers[name]
	if !ok {
		logger = newDaprLogger(name, hooks.ContextHookOptions{DiagnosticLogContextEnabled: true, ActivityLogContextEnabled: false, Folding: true})
		globalLoggers[name] = logger
	}

	return logger
}

// newUserAuditsLogger creates new Logger instance for user audit log.
func newUserAuditsLogger(name string) Logger {
	return newUserLogger(name, LogTypeUserAudits, hooks.ContextHookOptions{DiagnosticLogContextEnabled: false, ActivityLogContextEnabled: true, Folding: false, OtelLogrusHookEnabled: true, OtelLogrusHookName: name})
}

// newUserDiagnosticsLogger creates new Logger instance for user diagnostic log.
func newUserDiagnosticsLogger(name string) Logger {
	return newUserLogger(name, LogTypeUserDiagnostics, hooks.ContextHookOptions{DiagnosticLogContextEnabled: true, ActivityLogContextEnabled: true, Folding: false, OtelLogrusHookEnabled: true, OtelLogrusHookName: name})
}

func GetUserAuditsLogger() Logger {
	globalUserAuditsLoggerOnce.Do(func() {
		globalUserAuditsLogger = newUserAuditsLogger("coa.runtime.user.audits")
	})
	return globalUserAuditsLogger
}

func GetUserDiagnosticsLogger() Logger {
	globalUserDiagnosticsLoggerOnce.Do(func() {
		globalUserDiagnosticsLogger = newUserDiagnosticsLogger("coa.runtime.user.diagnostics")
	})
	return globalUserDiagnosticsLogger
}

func getLoggers() map[string]Logger {
	globalLoggersLock.RLock()
	defer globalLoggersLock.RUnlock()

	l := map[string]Logger{}
	for k, v := range globalLoggers {
		l[k] = v
	}

	return l
}
