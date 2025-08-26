/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
)

const (
	// Default log file name
	defaultLogFileName = "symphony-remote-agent.log"
	// Environment variable for custom log file path
	envLogFilePath  = "SYMPHONY_REMOTE_AGENT_LOG_FILE_PATH"
	maxHTTPBodySize = 1024 * 1024 // 1MB
)

type FileLogger struct {
	*CoaLogger

	// temp file logging fields
	tempFileEnabled bool
	tempFilePath    string
	tempFile        *os.File
	tempFileWriter  io.Writer
	logOffset       int64
	offsetMutex     sync.Mutex
}

func newFileLogger(name string, contextOptions hooks.ContextHookOptions) *FileLogger {
	logger := newCoaLogger(name, contextOptions)
	fileLogger := &FileLogger{
		CoaLogger: logger,
	}
	fileLogger.enableTempFileLogging()
	return fileLogger
}

// enableTempFileLogging sets up temp file logging for remote agents
func (l *FileLogger) enableTempFileLogging() error {
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
func (l *FileLogger) GetLogOffset() int64 {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	return l.logOffset
}

// SetLogOffset sets the current log offset
func (l *FileLogger) SetLogOffset(offset int64) {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	l.logOffset = offset
}

// GetLogsFromOffset reads logs from the temp file starting at the specified offset
func (l *FileLogger) GetLogsFromOffset(fromOffset int64) ([]string, int64, error) {
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
func (l *FileLogger) Close() error {
	if l.tempFile != nil {
		return l.tempFile.Close()
	}
	return nil
}

// IsTempFileEnabled returns whether temp file logging is enabled
func (l *FileLogger) IsTempFileEnabled() bool {
	return l.tempFileEnabled
}

// GetTempFilePath returns the temp file path
func (l *FileLogger) GetTempFilePath() string {
	return l.tempFilePath
}

// getLogFilePath determines the log file path, checking environment variable first,
// then falling back to a cross-platform default location
func (l *FileLogger) getLogFilePath() string {
	// Check environment variable first
	if customPath := os.Getenv(envLogFilePath); customPath != "" {
		return customPath
	}

	// Use cross-platform temp directory with default filename
	return filepath.Join(os.TempDir(), defaultLogFileName)
}

// ensureLogDirectory creates the directory for the log file if it doesn't exist
func (l *FileLogger) ensureLogDirectory() error {
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
