/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
	"github.com/sirupsen/logrus"
)

// CoaLogger is the implemention for logrus
type CoaLogger struct {
	// name is the name of logger that is published to log as a scope
	name string
	// logger is the logrus logger
	logger *logrus.Logger

	// sharedFieldsLock is the mutex for sharedFields
	sharedFieldsLock sync.Mutex
	// sharedFields is the fields that are shared among loggers
	sharedFields logrus.Fields

	callerSkip uint

	// temp file logging fields
	tempFileEnabled bool
	tempFilePath    string
	tempFile        *os.File
	tempFileWriter  io.Writer
	logOffset       int64
	offsetMutex     sync.Mutex
}

const (
	// Default log file name
	defaultLogFileName = "symphony-remote-agent.log"
	// Environment variable for custom log file path
	envLogFilePath = "SYMPHONY_REMOTE_AGENT_LOG_FILE_PATH"
)

var CoaVersion string = "unknown"

func GetCoaVersion() string {
	if envVersion, ok := os.LookupEnv("CHART_VERSION"); ok {
		return envVersion
	}
	return CoaVersion
}

func newCoaLogger(name string, contextOptions hooks.ContextHookOptions) *CoaLogger {
	newLogger := logrus.New()
	newLogger.AddHook(hooks.NewContextHookWithOptions(contextOptions))
	newLogger.SetOutput(os.Stdout)

	dl := &CoaLogger{
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
func (l *CoaLogger) EnableJSONOutput(enabled bool) {
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
		logFieldCoaVer:   GetCoaVersion(),
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
func (l *CoaLogger) SetAppID(id string) {
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
func (l *CoaLogger) SetOutputLevel(outputLevel LogLevel) {
	l.logger.SetLevel(toLogrusLevel(outputLevel))
}

// WithLogType specify the log_type field in log. Default value is LogTypeLog
func (l *CoaLogger) WithLogType(logType string) Logger {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	l.sharedFields[logFieldType] = logType
	return l
}

func (l *CoaLogger) GetSharedFields() logrus.Fields {
	l.sharedFieldsLock.Lock()
	defer l.sharedFieldsLock.Unlock()
	return l.sharedFields
}

// Info logs a message at level Info.
func (l *CoaLogger) InfoCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *CoaLogger) InfofCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *CoaLogger) DebugCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *CoaLogger) DebugfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *CoaLogger) WarnCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *CoaLogger) WarnfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *CoaLogger) ErrorCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *CoaLogger) ErrorfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *CoaLogger) FatalCtx(ctx context.Context, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *CoaLogger) FatalfCtx(ctx context.Context, format string, args ...interface{}) {
	l.logger.WithContext(ctx).WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatalf(format, args...)
}

// Info logs a message at level Info.
func (l *CoaLogger) Info(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.InfoLevel, args...)
}

// Infof logs a message at level Info.
func (l *CoaLogger) Infof(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.InfoLevel, format, args...)
}

// Debug logs a message at level Debug.
func (l *CoaLogger) Debug(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.DebugLevel, args...)
}

// Debugf logs a message at level Debug.
func (l *CoaLogger) Debugf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.DebugLevel, format, args...)
}

// Warn logs a message at level Warn.
func (l *CoaLogger) Warn(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.WarnLevel, args...)
}

// Warnf logs a message at level Warn.
func (l *CoaLogger) Warnf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.WarnLevel, format, args...)
}

// Error logs a message at level Error.
func (l *CoaLogger) Error(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Log(logrus.ErrorLevel, args...)
}

// Errorf logs a message at level Error.
func (l *CoaLogger) Errorf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Logf(logrus.ErrorLevel, format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func (l *CoaLogger) Fatal(args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *CoaLogger) Fatalf(format string, args ...interface{}) {
	l.logger.WithFields(l.GetSharedFields()).WithField("caller", getCaller(int(l.callerSkip))).Fatalf(format, args...)
}

// enableTempFileLogging sets up temp file logging for remote agents
func (l *CoaLogger) enableTempFileLogging() error {
	l.tempFileEnabled = true
	l.tempFilePath = l.getLogFilePath()

	// Ensure the directory exists
	if err := l.ensureLogDirectory(); err != nil {
		return fmt.Errorf("failed to ensure log directory: %w", err)
	}

	// Create temp file with better error handling
	file, err := os.OpenFile(l.tempFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp log file at %s: %w", l.tempFilePath, err)
	}
	l.tempFile = file

	// Create multi-writer (stdout + temp file)
	l.tempFileWriter = io.MultiWriter(os.Stdout, file)
	l.logger.SetOutput(l.tempFileWriter)

	return nil
}

// GetLogOffset returns the current log offset
func (l *CoaLogger) GetLogOffset() int64 {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	return l.logOffset
}

// SetLogOffset sets the current log offset
func (l *CoaLogger) SetLogOffset(offset int64) {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	l.logOffset = offset
}

const maxHTTPBodySize = 1024 * 1024 // 1MB

// GetLogsFromOffset reads logs from the temp file starting at the specified offset
func (l *CoaLogger) GetLogsFromOffset(fromOffset int64) ([]string, int64, error) {
	if !l.tempFileEnabled || l.tempFile == nil {
		return nil, 0, fmt.Errorf("temp file logging not enabled")
	}

	file, err := os.Open(l.tempFilePath)
	if err != nil {
		return nil, fromOffset, fmt.Errorf("failed to open temp log file for reading: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fromOffset, fmt.Errorf("failed to get file info: %w", err)
	}

	if fromOffset >= fileInfo.Size() {
		return []string{}, fromOffset, nil
	}

	_, err = file.Seek(fromOffset, 0)
	if err != nil {
		return nil, fromOffset, fmt.Errorf("failed to seek to offset %d: %w", fromOffset, err)
	}

	type logLine struct {
		line   string
		length int
	}
	var (
		lines         []logLine
		totalSize     int
		currentOffset = fromOffset
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			// +1 for newline, +2 for quotes and +1 for comma in JSON array (approx)
			lineLen := len(line) + 1 + 3
			lines = append(lines, logLine{line, lineLen})
			totalSize += lineLen
		}
		currentOffset += int64(len(line)) + 1
	}
	if err := scanner.Err(); err != nil {
		return nil, currentOffset, fmt.Errorf("error reading log file: %w", err)
	}

	// Truncate if over the limit, keep latest logs
	startIdx := 0
	for totalSize > maxHTTPBodySize && startIdx < len(lines) {
		totalSize -= lines[startIdx].length
		startIdx++
	}

	truncatedLogs := make([]string, len(lines)-startIdx)
	for i := startIdx; i < len(lines); i++ {
		truncatedLogs[i-startIdx] = lines[i].line
	}

	// Calculate new offset if truncated
	newOffset := fromOffset
	if startIdx > 0 {
		// Re-read file to calculate offset of the first included line
		file.Seek(fromOffset, 0)
		scanner2 := bufio.NewScanner(file)
		for i := 0; i < startIdx && scanner2.Scan(); i++ {
			line := scanner2.Text()
			newOffset += int64(len(line)) + 1
		}
	}

	// Update the internal offset
	l.SetLogOffset(currentOffset)

	if startIdx > 0 {
		return truncatedLogs, newOffset, nil
	}
	return truncatedLogs, currentOffset, nil
}

// Close closes the temp file if it exists
func (l *CoaLogger) Close() error {
	if l.tempFile != nil {
		return l.tempFile.Close()
	}
	return nil
}

// IsTempFileEnabled returns whether temp file logging is enabled
func (l *CoaLogger) IsTempFileEnabled() bool {
	return l.tempFileEnabled
}

// GetTempFilePath returns the temp file path
func (l *CoaLogger) GetTempFilePath() string {
	return l.tempFilePath
}

// getLogFilePath determines the log file path, checking environment variable first,
// then falling back to a cross-platform default location
func (l *CoaLogger) getLogFilePath() string {
	// Check environment variable first
	if customPath := os.Getenv(envLogFilePath); customPath != "" {
		return customPath
	}

	// Use cross-platform temp directory with default filename
	return filepath.Join(os.TempDir(), defaultLogFileName)
}

// ensureLogDirectory creates the directory for the log file if it doesn't exist
func (l *CoaLogger) ensureLogDirectory() error {
	logDir := filepath.Dir(l.tempFilePath)

	// Check if directory exists
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		// Create directory with appropriate permissions
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}

	return nil
}

func getCaller(extraSkip int) string {
	callerPc := make([]uintptr, 1)
	runtime.Callers(1+extraSkip, callerPc) // skipping caller of getCaller().
	callerFrame, _ := runtime.CallersFrames(callerPc).Next()
	return fmt.Sprintf("%s:%d", callerFrame.Function, callerFrame.Line)
}
