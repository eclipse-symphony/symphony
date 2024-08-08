/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"bytes"
	"context"
	"os"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
	"github.com/sirupsen/logrus"
)

// user-facing logger is the implemention for logrus
type userLogger struct {
	// name is the name of logger that is published to log as a scope
	name string

	// logger is the logrus logger
	logger *logrus.Logger

	// sharedFieldsLock is the mutex for sharedFields
	sharedFieldsLock sync.Mutex
	// sharedFields is the fields that are shared among loggers
	sharedFields logrus.Fields

	logType string
}

type RetentionBuffer struct {
	buffer  *bytes.Buffer
	maxSize int
	mu      sync.Mutex
}

func NewRetentionBuffer(maxSize int) *RetentionBuffer {
	return &RetentionBuffer{
		buffer:  new(bytes.Buffer),
		maxSize: maxSize,
	}
}

func (rb *RetentionBuffer) Write(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.buffer.Len()+len(p) > rb.maxSize {
		// Discard old data to make space
		excess := rb.buffer.Len() + len(p) - rb.maxSize
		rb.buffer.Next(excess)
	}

	return rb.buffer.Write(p)
}

func (rb *RetentionBuffer) Bytes() []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.buffer.Bytes()
}

func newUserLogger(name string, logType string, contextOptions hooks.ContextHookOptions) *userLogger {
	newLogger := logrus.New()
	// Create a custom buffer with a maximum size of 10KB
	retentionBuffer := NewRetentionBuffer(10 * 1024)
	newLogger.SetOutput(retentionBuffer) // disable output, user logger should not output to console
	newLogger.AddHook(hooks.NewContextHookWithOptions(contextOptions))

	ul := &userLogger{
		name:   name,
		logger: newLogger,
		sharedFields: logrus.Fields{
			logFieldScope: name,
			logFieldType:  logType,
		},
		logType: logType,
	}

	ul.EnableJSONOutput(defaultJSONOutput)

	return ul
}

// EnableJSONOutput enables JSON formatted output log
func (l *userLogger) EnableJSONOutput(enabled bool) {
	// user logger should not output to console
	var formatter logrus.Formatter

	fieldMap := logrus.FieldMap{
		// If time field name is conflicted, logrus adds "fields." prefix.
		// So rename to unused field @time to avoid the confliction.
		logrus.FieldKeyTime:  logFieldTimeStamp,
		logrus.FieldKeyLevel: logFieldLevel,
		logrus.FieldKeyMsg:   logFieldMessage,
	}

	hostname, _ := os.Hostname()
	l.sharedFieldsLock.Lock()
	l.sharedFields = logrus.Fields{
		logFieldScope:    l.sharedFields[logFieldScope],
		logFieldType:     l.logType,
		logFieldInstance: hostname,
		logFieldDaprVer:  CoaVersion,
	}
	l.sharedFieldsLock.Unlock()

	if enabled {
		formatter = &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	} else {
		formatter = &logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	}

	l.logger.SetFormatter(formatter)
}

// SetOutputLevel sets log output level
func (l *userLogger) SetOutputLevel(outputLevel LogLevel) {
	l.logger.SetLevel(toLogrusLevel(outputLevel))
}

// WithLogType specify the log_type field in log. Default value is LogTypeLog
func (l *userLogger) WithLogType(logType string) Logger {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	l.sharedFields[logFieldType] = logType
	return l
}

// SetAppID sets app_id field in the log. Default value is empty string
func (l *userLogger) SetAppID(id string) {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	l.sharedFields[logFieldAppID] = id
}

func (l *userLogger) GetSharedFields() logrus.Fields {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	return l.sharedFields
}

// Info logs a message at level Info.
func (l *userLogger) InfoCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *userLogger) InfofCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *userLogger) DebugCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *userLogger) DebugfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *userLogger) WarnCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *userLogger) WarnfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *userLogger) ErrorCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *userLogger) ErrorfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *userLogger) FatalCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *userLogger) FatalfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).Fatalf(format, args...)
}

// Info logs a message at level Info.
func (l *userLogger) Info(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *userLogger) Infof(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *userLogger) Debug(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *userLogger) Debugf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *userLogger) Warn(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *userLogger) Warnf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *userLogger) Error(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *userLogger) Errorf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *userLogger) Fatal(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *userLogger) Fatalf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).Fatalf(format, args...)
}
