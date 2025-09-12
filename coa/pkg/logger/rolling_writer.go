// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// SPDX-License-Identifier: MIT

package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const _mb = 1024 * 1024
const _tsLayout = "20060102-150405"

// RollingWriter is an in-repo, lumberjack-like rolling file writer.
// - Rotates by size (MaxSize MB)
// - Keeps up to MaxBackups rotated files and prunes by MaxAge (days)
// - Optionally compresses older rotated files (.gz)
// It implements io.WriteCloser and is safe for concurrent use.
type RollingWriter struct {
	// Config
	Filename   string // active log file path
	MaxSize    int    // MB
	MaxBackups int    // number of rotated files to retain
	MaxAge     int    // days
	Compress   bool   // gzip rotated files (older ones)

	// State
	mu   sync.Mutex
	file *os.File
	size int64
}

// NewRollingWriter creates a new rolling writer with the given configuration.
func NewRollingWriter(filename string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) (*RollingWriter, error) {
	w := &RollingWriter{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}
	if err := w.openOrCreate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RollingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openOrCreate(); err != nil {
			return 0, err
		}
	}

	if w.shouldRotate(len(p)) {
		if err := w.rotate(); err != nil {
			// If rotation fails, try to keep writing to current file to avoid data loss
			// but still return the rotation error to surface the issue.
			n, werr := w.file.Write(p)
			w.size += int64(n)
			if werr != nil {
				return n, fmt.Errorf("rotate failed: %v; write failed: %w", err, werr)
			}
			return n, fmt.Errorf("rotate failed: %w", err)
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *RollingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

func (w *RollingWriter) openOrCreate() error {
	// Ensure directory exists
	dir := filepath.Dir(w.Filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create log dir %s: %w", dir, err)
	}
	// Open file append
	f, err := os.OpenFile(w.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", w.Filename, err)
	}
	w.file = f

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		w.file = nil
		return fmt.Errorf("stat log file %s: %w", w.Filename, err)
	}
	w.size = info.Size()
	return nil
}

func (w *RollingWriter) shouldRotate(add int) bool {
	if w.MaxSize <= 0 {
		return false
	}
	threshold := int64(w.MaxSize) * _mb
	return w.size+int64(add) >= threshold
}

func (w *RollingWriter) rotate() error {
	// Close current file before rename
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close before rotate: %w", err)
		}
		w.file = nil
	}

	ts := time.Now()
	rotated := buildRotatedPath(w.Filename, ts)

	// Rename active -> rotated
	if err := os.Rename(w.Filename, rotated); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rename during rotate: %w", err)
	}

	// Recreate active file
	if err := w.openOrCreate(); err != nil {
		return fmt.Errorf("reopen after rotate: %w", err)
	}
	w.size = 0

	// Retention and compression
	if err := w.cleanup(); err != nil {
		// Non-fatal; keep going
		return fmt.Errorf("cleanup after rotate: %w", err)
	}
	return nil
}

/*
cleanup applies age and backup retention, and optional compression of older files.

Fix: avoid recursion loops by pruning and updating the in-memory list in-place.
*/
func (w *RollingWriter) cleanup() error {
	dir := filepath.Dir(w.Filename)
	base := strings.TrimSuffix(filepath.Base(w.Filename), filepath.Ext(w.Filename))
	ext := filepath.Ext(w.Filename)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("list dir %s: %w", dir, err)
	}

	type seg struct {
		path string
		ts   time.Time
		gz   bool
	}
	var segments []seg

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Accept patterns:
		//
		//	base-YYYYMMDD-HHMMSS + ext
		//	base-YYYYMMDD-HHMMSS + ext + ".gz"
		if !strings.HasPrefix(name, base+"-") {
			continue
		}
		if !(strings.HasSuffix(name, ext) || strings.HasSuffix(name, ext+".gz")) {
			continue
		}
		// Trim .gz if present to find timestamp
		trimmed := strings.TrimSuffix(name, ".gz")
		trimmed = strings.TrimSuffix(trimmed, ext)
		tsStr := strings.TrimPrefix(trimmed, base+"-")
		ts, perr := time.Parse(_tsLayout, tsStr)
		if perr != nil {
			continue
		}
		segments = append(segments, seg{
			path: filepath.Join(dir, name),
			ts:   ts,
			gz:   strings.HasSuffix(name, ".gz"),
		})
	}

	// Age prune (in-place, no recursion)
	if w.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -w.MaxAge)
		kept := make([]seg, 0, len(segments))
		for _, s := range segments {
			if s.ts.Before(cutoff) {
				_ = os.Remove(s.path)
				continue
			}
			kept = append(kept, s)
		}
		segments = kept
	}

	// Sort by time ascending (oldest first)
	sort.Slice(segments, func(i, j int) bool { return segments[i].ts.Before(segments[j].ts) })

	// Enforce backup count (delete oldest beyond limit)
	if w.MaxBackups > 0 && len(segments) > w.MaxBackups {
		toDelete := segments[0 : len(segments)-w.MaxBackups]
		for _, s := range toDelete {
			_ = os.Remove(s.path)
		}
		segments = segments[len(segments)-w.MaxBackups:]
	}

	// Compression: keep newest rotated uncompressed for quick reads; compress older ones
	if w.Compress && len(segments) > 1 {
		for i := 0; i < len(segments)-1; i++ {
			s := segments[i]
			if s.gz {
				continue
			}
			_ = compressFile(s.path)
		}
	}

	return nil
}

func buildRotatedPath(basePath string, ts time.Time) string {
	dir := filepath.Dir(basePath)
	ext := filepath.Ext(basePath)
	base := strings.TrimSuffix(filepath.Base(basePath), ext)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", base, ts.Format(_tsLayout), ext))
}

func compressFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	outPath := path + ".gz"
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	gzw := gzip.NewWriter(out)
	_, copyErr := io.Copy(gzw, in)
	closeErr1 := gzw.Close()
	closeErr2 := out.Close()

	if copyErr != nil {
		_ = os.Remove(outPath)
		return copyErr
	}
	if closeErr1 != nil {
		_ = os.Remove(outPath)
		return closeErr1
	}
	if closeErr2 != nil {
		_ = os.Remove(outPath)
		return closeErr2
	}

	// Remove original after successful compression
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}
