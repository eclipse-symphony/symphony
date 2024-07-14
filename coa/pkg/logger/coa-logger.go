/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
	"github.com/sirupsen/logrus"
)

// daprLogger is the implemention for logrus
type daprLogger struct {
	// name is the name of logger that is published to log as a scope
	name string
	// logger is the logrus logger
	logger *logrus.Logger

	// sharedFieldsLock is the mutex for sharedFields
	sharedFieldsLock sync.Mutex
	// sharedFields is the fields that are shared among loggers
	sharedFields logrus.Fields

	callerSkip uint
}

var DaprVersion string = "unknown"

func newDaprLogger(name string) *daprLogger {
	newLogger := logrus.New()
	newLogger.AddHook(hooks.NewContextHook())
	newLogger.SetOutput(os.Stdout)

	dl := &daprLogger{
		name:   name,
		logger: newLogger,
		sharedFields: logrus.Fields{
			logFieldScope: name,
			logFieldType:  LogTypeLog,
		},
		callerSkip: 2, // skip 2 frames to get the caller of log functions
	}

	dl.EnableJSONOutput(defaultJSONOutput)

	return dl
}

// EnableJSONOutput enables JSON formatted output log
func (l *daprLogger) EnableJSONOutput(enabled bool) {
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
		logFieldType:     LogTypeLog,
		logFieldInstance: hostname,
		logFieldDaprVer:  DaprVersion,
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
	// l.logger.SetReportCaller(true)
}

// SetAppID sets app_id field in the log. Default value is empty string
func (l *daprLogger) SetAppID(id string) {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	l.sharedFields[logFieldAppID] = id
}

func toLogrusLevel(lvl LogLevel) logrus.Level {
	// ignore error because it will never happens
	l, _ := logrus.ParseLevel(string(lvl))
	return l
}

// SetOutputLevel sets log output level
func (l *daprLogger) SetOutputLevel(outputLevel LogLevel) {
	l.logger.SetLevel(toLogrusLevel(outputLevel))
}

// WithLogType specify the log_type field in log. Default value is LogTypeLog
func (l *daprLogger) WithLogType(logType string) Logger {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	l.sharedFields[logFieldType] = logType
	return l
}

func (l *daprLogger) GetSharedFields() logrus.Fields {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	return l.sharedFields
}

// Info logs a message at level Info.
func (l *daprLogger) InfoCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *daprLogger) InfofCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *daprLogger) DebugCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *daprLogger) DebugfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *daprLogger) WarnCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *daprLogger) WarnfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *daprLogger) ErrorCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *daprLogger) ErrorfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *daprLogger) FatalCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *daprLogger) FatalfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatalf(format, args...)
}

// Info logs a message at level Info.
func (l *daprLogger) Info(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *daprLogger) Infof(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *daprLogger) Debug(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *daprLogger) Debugf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *daprLogger) Warn(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *daprLogger) Warnf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *daprLogger) Error(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *daprLogger) Errorf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *daprLogger) Fatal(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *daprLogger) Fatalf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatalf(format, args...)
}

func getCaller(extraSkip int) string {
	callerPc := make([]uintptr, 1)
	runtime.Callers(1+extraSkip, callerPc) // skipping caller of getCaller().
	callerFrame, _ := runtime.CallersFrames(callerPc).Next()
	return fmt.Sprintf("%s:%d", callerFrame.Function, callerFrame.Line)
}
