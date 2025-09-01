/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package logger

/*
Environment variables for file-logger rolling:
- SYMPHONY_REMOTE_AGENT_LOG_FILE_PATH:
  Absolute path to the active log file. Defaults to $TMPDIR/symphony-remote-agent.log
- SYMPHONY_REMOTE_AGENT_LOG_MAX_SIZE_MB:
  Max size (in MB) of the active log before rotation. Default: 10
- SYMPHONY_REMOTE_AGENT_LOG_MAX_BACKUPS:
  Max number of rotated history files to retain (excluding the active file). Default: 3
- SYMPHONY_REMOTE_AGENT_LOG_MAX_AGE_DAYS:
  Max age (in days) for rotated files; files older than this are pruned. Default: 7

Behavior:
- Writing: logs go to stdout and a size-rolling file via in-repo RollingWriter (rename-and-recreate).
- Rotation: active file is renamed to base-YYYYMMDD-HHMMSS.log, then a fresh active file is created. No copytruncate is used.
- Reading (GetLogsFromOffset):
  - Honors an intra-segment byte offset. The logger persists a cursor {segmentPath, segmentOffset}.
  - Starting point:
      * If the persisted segment exists: start at offset = clamp(fromOffset, 0..segmentSize).
      * If the persisted segment was pruned/renamed: start at the oldest available segment at offset 0.
  - Streams forward across newer segments (oldest → newest → active) until EOF.
  - If the combined payload exceeds 1MB (maxHTTPBodySize), the oldest lines in THIS RESPONSE WINDOW are dropped until within limit.
  - Returns newOffset as the byte offset within the last segment read. The persisted cursor is updated to the last segment and its size.
- Compression: disabled. Rotated segments are not compressed. .gz is not supported.
*/

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
)

const (
	// Default log file name
	defaultLogFileName = "symphony-remote-agent.log"
	// Environment variable for custom log file path
	envLogFilePath  = "SYMPHONY_REMOTE_AGENT_LOG_FILE_PATH"
	maxHTTPBodySize = 1024 * 1024 // 1MB

	// Rolling defaults and env keys
	defaultMaxSizeMB  = 1
	defaultMaxBackups = 5
	defaultMaxAgeDays = 7

	envLogMaxSizeMB  = "SYMPHONY_REMOTE_AGENT_LOG_MAX_SIZE_MB"
	envLogMaxBackups = "SYMPHONY_REMOTE_AGENT_LOG_MAX_BACKUPS"
	envLogMaxAgeDays = "SYMPHONY_REMOTE_AGENT_LOG_MAX_AGE_DAYS"
	envLogCompress   = "SYMPHONY_REMOTE_AGENT_LOG_COMPRESS"

	// Timestamp layout for rotated segments (matches rolling_writer)
	tsLayout = "20060102-150405"
)

type FileLogger struct {
	*CoaLogger

	// temp file logging fields
	tempFileEnabled bool
	tempFilePath    string
	tempFile        *os.File
	tempFileWriter  io.Writer
	rolling         io.WriteCloser
	logOffset       int64
	offsetMutex     sync.Mutex
}

// persisted cursor for incremental reads (segment + intra-segment offset)
type logCursor struct {
	SegmentPath   string    `json:"segment_path"`
	SegmentOffset int64     `json:"segment_offset"`
	UpdatedAt     time.Time `json:"updated_at"`
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

	// Compression is disabled by design
	rw, err := NewRollingWriter(
		l.tempFilePath,
		getEnvInt(envLogMaxSizeMB, defaultMaxSizeMB),
		getEnvInt(envLogMaxBackups, defaultMaxBackups),
		getEnvInt(envLogMaxAgeDays, defaultMaxAgeDays),
		false, // Compress disabled
	)
	if err != nil {
		return fmt.Errorf("failed to create rolling writer at %s: %w", l.tempFilePath, err)
	}
	l.rolling = rw
	l.tempFile = nil

	// stdout + rolling writer
	l.tempFileWriter = io.MultiWriter(os.Stdout, rw)
	l.logger.SetOutput(l.tempFileWriter)

	return nil
}

// GetLogOffset returns the current log offset (intra-segment for the last segment read)
func (l *FileLogger) GetLogOffset() int64 {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	return l.logOffset
}

// SetLogOffset sets the current log offset (intra-segment)
func (l *FileLogger) SetLogOffset(offset int64) {
	l.offsetMutex.Lock()
	defer l.offsetMutex.Unlock()
	l.logOffset = offset
}

/*
GetLogsFromOffset reads logs starting from the persisted segment and the caller-provided
intra-segment offset, then streams forward through newer segments up to the active file.

Semantics:
- If a cursor exists and its segment still exists, start at offset = clamp(fromOffset, 0..segmentSize).
- If no cursor or the segment no longer exists, start at the oldest available segment with offset=0.
- Reads across segments (oldest→...→active). If payload exceeds 1MB, drop earliest lines in THIS response.
- Returns the newOffset which is the byte offset (end-of-file) of the last segment read; also persists the cursor.

Note:
- No gzip support. Only plain rotated segments (base-YYYYMMDD-HHMMSS.ext) and the active file are considered.
*/
func (l *FileLogger) GetLogsFromOffset(fromOffset int64) ([]string, int64, error) {
	if !l.tempFileEnabled {
		return nil, 0, fmt.Errorf("temp file logging not enabled")
	}

	// Snapshot segments (rotated + active), oldest -> newest
	segs, err := l.listLogSegments(l.tempFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("list segments: %w", err)
	}
	if len(segs) == 0 {
		return []string{}, 0, nil
	}

	// Load persisted cursor
	cur, _ := l.readCursor()

	// Determine start segment and offset
	startIdx := 0
	startOff := int64(0)
	if cur != nil && cur.SegmentPath != "" {
		for i, p := range segs {
			if p == cur.SegmentPath {
				startIdx = i
				fi, err := os.Stat(p)
				if err == nil {
					if fromOffset < 0 {
						startOff = 0
					} else if fromOffset > fi.Size() {
						startOff = fi.Size()
					} else {
						startOff = fromOffset
					}
				}
				break
			}
		}
		// If not found, fall back to oldest at offset 0
	}

	type logLine struct {
		line   string
		length int
	}
	var (
		lines     []logLine
		totalSize int
	)

	// Helper to append a line with size accounting (+1 newline +3 overhead for transport)
	appendLine := func(s string) {
		if strings.TrimSpace(s) == "" {
			return
		}
		llen := len(s) + 4 // 1 newline + ~3 overhead
		lines = append(lines, logLine{s, llen})
		totalSize += llen
		for totalSize > maxHTTPBodySize && len(lines) > 0 {
			totalSize -= lines[0].length
			lines = lines[1:]
		}
	}

	// Read from start segment at startOff, then subsequent segments
	var lastSegPath string
	for i := startIdx; i < len(segs); i++ {
		p := segs[i]
		f, err := os.Open(p)
		if err != nil {
			// skip unreadable segment
			continue
		}
		if i == startIdx && startOff > 0 {
			_, _ = f.Seek(startOff, 0)
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			appendLine(sc.Text())
		}
		_ = f.Close()
		lastSegPath = p
	}

	// Prepare output strings
	out := make([]string, len(lines))
	for i := range lines {
		out[i] = lines[i].line
	}

	// Determine new offset as EOF of the last segment read
	newOffset := int64(0)
	if lastSegPath == "" {
		// no readable segments; anchor to active file if exists
		lastSegPath = l.tempFilePath
	}
	if fi, err := os.Stat(lastSegPath); err == nil {
		newOffset = fi.Size()
	}

	// Persist cursor and update internal offset
	_ = l.writeCursor(&logCursor{
		SegmentPath:   lastSegPath,
		SegmentOffset: newOffset,
		UpdatedAt:     time.Now(),
	})
	l.SetLogOffset(newOffset)

	return out, newOffset, nil
}

// Close closes the rolling writer and temp file if they exist
func (l *FileLogger) Close() error {
	if l.rolling != nil {
		if err := l.rolling.Close(); err != nil {
			return err
		}
	}
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

// getEnvInt parses an int env var with default.
func getEnvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return def
}

// listLogSegments returns absolute paths of rotated segments (oldest->newest) plus the active file.
// Only plain files are included. .gz is intentionally ignored.
func (l *FileLogger) listLogSegments(basePath string) ([]string, error) {
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	base := strings.TrimSuffix(filepath.Base(basePath), ext)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type seg struct {
		path string
		ts   time.Time
	}
	var segments []seg

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Match base-YYYYMMDD-HHMMSS.ext (no .gz)
		if !strings.HasPrefix(name, base+"-") {
			continue
		}
		if !strings.HasSuffix(name, ext) {
			continue
		}
		trimmed := strings.TrimSuffix(name, ext)
		tsStr := strings.TrimPrefix(trimmed, base+"-")
		ts, err := time.Parse(tsLayout, tsStr)
		if err != nil {
			continue
		}
		segments = append(segments, seg{
			path: filepath.Join(dir, name),
			ts:   ts,
		})
	}

	// Sort rotated by time ascending
	sort.Slice(segments, func(i, j int) bool { return segments[i].ts.Before(segments[j].ts) })

	var out []string
	for _, s := range segments {
		out = append(out, s.path)
	}
	// Append active file last if exists
	if _, err := os.Stat(basePath); err == nil {
		out = append(out, basePath)
	}
	return out, nil
}

// cursorPath returns the sidecar file path used to persist the read cursor.
func (l *FileLogger) cursorPath() string {
	return l.tempFilePath + ".cursor"
}

// readCursor loads the persisted read cursor if present.
func (l *FileLogger) readCursor() (*logCursor, error) {
	p := l.cursorPath()
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var c logCursor
	if jerr := json.Unmarshal(b, &c); jerr != nil {
		return nil, jerr
	}
	return &c, nil
}

// writeCursor persists the read cursor to the sidecar file.
func (l *FileLogger) writeCursor(c *logCursor) error {
	p := l.cursorPath()
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	// 0644 is sufficient; directory already ensured
	return os.WriteFile(p, data, 0644)
}
