package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/hooks"
)

// helpers

func rotatedName(base string, ts time.Time) (stem, ext, dir, rot string) {
	dir = filepath.Dir(base)
	ext = filepath.Ext(base)
	stem = strings.TrimSuffix(filepath.Base(base), ext)
	rot = filepath.Join(dir, fmt.Sprintf("%s-%s%s", stem, ts.Format(tsLayout), ext))
	return
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for i, s := range lines {
		if _, err := f.WriteString(s); err != nil {
			return err
		}
		if i != len(lines)-1 {
			if _, err := f.WriteString("\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

// tests

func TestFileLogger_ReadsAllSegments(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "agent.log")

	// Pre-create rotated segments: two plain files
	_, _, _, r1 := rotatedName(base, time.Now().Add(-2*time.Hour))
	_, _, _, r2 := rotatedName(base, time.Now().Add(-1*time.Hour))
	if err := writeLines(r1, []string{"r1-a", "r1-b"}); err != nil {
		t.Fatalf("seed rotated r1: %v", err)
	}
	if err := writeLines(r2, []string{"r2-a", "r2-b"}); err != nil {
		t.Fatalf("seed rotated r2: %v", err)
	}
	// Active file with a couple of lines
	if err := writeLines(base, []string{"active-a", "active-b"}); err != nil {
		t.Fatalf("seed active: %v", err)
	}

	// Point logger to this path
	t.Setenv(envLogFilePath, base)
	// Conservative rolling to avoid surprise cleanup during test
	t.Setenv(envLogMaxSizeMB, "10")
	t.Setenv(envLogMaxBackups, "10")
	t.Setenv(envLogMaxAgeDays, "7")

	l := newFileLogger("test", hooks.ContextHookOptions{})
	defer l.Close()

	lines, off, err := l.GetLogsFromOffset(0)
	if err != nil {
		t.Fatalf("GetLogsFromOffset: %v", err)
	}
	if off < 0 {
		t.Fatalf("offset negative: %d", off)
	}
	joined := strings.Join(lines, "\n")
	// Expect order to contain content from both rotated and active files
	if !strings.Contains(joined, "r1-a") || !strings.Contains(joined, "r1-b") {
		t.Fatalf("missing rotated r1 lines: %q", joined)
	}
	if !strings.Contains(joined, "r2-a") || !strings.Contains(joined, "r2-b") {
		t.Fatalf("missing rotated r2 lines: %q", joined)
	}
	if !strings.Contains(joined, "active-a") || !strings.Contains(joined, "active-b") {
		t.Fatalf("missing active lines: %q", joined)
	}
}

func TestFileLogger_EnvPathHonoredAndWrite(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "agent.log")
	t.Setenv(envLogFilePath, base)
	t.Setenv(envLogMaxSizeMB, "10")
	t.Setenv(envLogMaxBackups, "3")
	t.Setenv(envLogMaxAgeDays, "7")

	l := newFileLogger("test", hooks.ContextHookOptions{})
	defer l.Close()

	for i := 0; i < 20; i++ {
		l.Infof("line-%03d", i)
	}

	fi, err := os.Stat(base)
	if err != nil {
		t.Fatalf("log file not found at env path: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("log file is empty at %s", base)
	}

	lines, _, err := l.GetLogsFromOffset(0)
	if err != nil {
		t.Fatalf("GetLogsFromOffset: %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("expected some lines from logger")
	}
}

func TestFileLogger_FromOffsetGTE_TotalEmpty(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "agent.log")

	// Seed only active file
	if err := writeLines(base, []string{"a1", "a2"}); err != nil {
		t.Fatalf("seed active: %v", err)
	}

	// Point logger and ensure no rotation interference
	t.Setenv(envLogFilePath, base)
	t.Setenv(envLogMaxSizeMB, "10")
	t.Setenv(envLogMaxBackups, "10")
	t.Setenv(envLogMaxAgeDays, "7")

	// Persist a cursor pointing to the active file so fromOffset is honored
	cursorPath := base + ".cursor"
	fi, err := os.Stat(base)
	if err != nil {
		t.Fatalf("stat active: %v", err)
	}
	activeSize := fi.Size()
	if err := os.WriteFile(cursorPath, []byte(fmt.Sprintf(`{"segment_path":"%s","segment_offset":%d,"updated_at":"%s"}`, base, 0, time.Now().Format(time.RFC3339Nano))), 0o644); err != nil {
		t.Fatalf("write cursor: %v", err)
	}

	l := newFileLogger("test", hooks.ContextHookOptions{})
	defer l.Close()

	lines, off, err := l.GetLogsFromOffset(activeSize)
	if err != nil {
		t.Fatalf("GetLogsFromOffset: %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected no new lines when fromOffset >= size, got %d", len(lines))
	}
	if off != activeSize {
		t.Fatalf("expected newOffset == active size %d, got %d", activeSize, off)
	}
}

func TestFileLogger_CursorResumeAcrossSegments(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "agent.log")

	// Seed two rotated segments and an active file
	_, _, _, r1 := rotatedName(base, time.Now().Add(-2*time.Hour))
	_, _, _, r2 := rotatedName(base, time.Now().Add(-1*time.Hour))
	if err := writeLines(r1, []string{"r1-a"}); err != nil {
		t.Fatalf("seed r1: %v", err)
	}
	if err := writeLines(r2, []string{"r2-a"}); err != nil {
		t.Fatalf("seed r2: %v", err)
	}
	if err := writeLines(base, []string{"active-a"}); err != nil {
		t.Fatalf("seed active: %v", err)
	}

	t.Setenv(envLogFilePath, base)
	t.Setenv(envLogMaxSizeMB, "10")
	t.Setenv(envLogMaxBackups, "10")
	t.Setenv(envLogMaxAgeDays, "7")

	// Persist a cursor pointing to the first rotated segment so we start there
	cursorPath := base + ".cursor"
	if err := os.WriteFile(cursorPath, []byte(fmt.Sprintf(`{"segment_path":"%s","segment_offset":0,"updated_at":"%s"}`, r1, time.Now().Format(time.RFC3339Nano))), 0o644); err != nil {
		t.Fatalf("write cursor: %v", err)
	}

	l := newFileLogger("test", hooks.ContextHookOptions{})
	defer l.Close()

	lines, _, err := l.GetLogsFromOffset(0)
	if err != nil {
		t.Fatalf("GetLogsFromOffset: %v", err)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "r1-a") {
		t.Fatalf("missing r1-a: %q", joined)
	}
	if !strings.Contains(joined, "r2-a") {
		t.Fatalf("missing r2-a: %q", joined)
	}
	if !strings.Contains(joined, "active-a") {
		t.Fatalf("missing active-a: %q", joined)
	}
}
