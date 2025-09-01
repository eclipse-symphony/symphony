package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

const kb = 1024

func TestRotateBySize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rw.log")

	w, err := NewRollingWriter(path, 1, 5, 7, false) // 1 MB
	if err != nil {
		t.Fatalf("NewRollingWriter: %v", err)
	}
	defer w.Close()

	// Fill close to threshold, then exceed to trigger rotation
	if _, err = w.Write(bytes.Repeat([]byte("A"), 900*kb)); err != nil {
		t.Fatalf("write1: %v", err)
	}
	time.Sleep(1100 * time.Millisecond) // avoid same-second filename collisions
	if _, err = w.Write(bytes.Repeat([]byte("B"), 300*kb)); err != nil {
		t.Fatalf("write2: %v", err)
	}

	segs := globSegments(t, path)
	if len(segs) < 1 {
		t.Fatalf("expected at least 1 rotated segment, got %v", segs)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("active file should exist: %v", err)
	}
}

func TestMaxBackupsRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rw.log")

	w, err := NewRollingWriter(path, 1, 2, 7, false) // keep only 2 backups
	if err != nil {
		t.Fatalf("NewRollingWriter: %v", err)
	}
	defer w.Close()

	// Create 3 rotations (oldest should be pruned)
	for i := 0; i < 3; i++ {
		if _, err = w.Write(bytes.Repeat([]byte("X"), 900*kb)); err != nil {
			t.Fatalf("prefill %d: %v", i, err)
		}
		time.Sleep(1100 * time.Millisecond)
		if _, err = w.Write(bytes.Repeat([]byte("Y"), 300*kb)); err != nil {
			t.Fatalf("trigger rotate %d: %v", i, err)
		}
	}

	segs := globSegments(t, path)
	if len(segs) > 2 {
		t.Fatalf("expected at most 2 rotated segments due to MaxBackups, got %d: %v", len(segs), segs)
	}
}

func TestCompressRotated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rw.log")

	w, err := NewRollingWriter(path, 1, 10, 7, true) // compression enabled
	if err != nil {
		t.Fatalf("NewRollingWriter: %v", err)
	}
	defer w.Close()

	// Make at least 2 rotations so the older one can be compressed
	for i := 0; i < 2; i++ {
		if _, err = w.Write(bytes.Repeat([]byte("C"), 900*kb)); err != nil {
			t.Fatalf("prefill %d: %v", i, err)
		}
		time.Sleep(1100 * time.Millisecond)
		if _, err = w.Write(bytes.Repeat([]byte("D"), 300*kb)); err != nil {
			t.Fatalf("trigger rotate %d: %v", i, err)
		}
	}

	segs := globSegments(t, path)
	hasGz := false
	for _, s := range segs {
		if strings.HasSuffix(s, ".gz") {
			hasGz = true
			break
		}
	}
	if !hasGz {
		t.Fatalf("expected at least one compressed (.gz) rotated segment, got %v", segs)
	}
}

func TestMaxAgeRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rw.log")

	// Create fake old rotated files (10 days old)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	old1 := filepath.Join(dir, base+"-"+time.Now().AddDate(0, 0, -10).Format(_tsLayout)+ext)
	old2 := filepath.Join(dir, base+"-"+time.Now().AddDate(0, 0, -11).Format(_tsLayout)+ext)
	if err := os.WriteFile(old1, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed old1: %v", err)
	}
	if err := os.WriteFile(old2, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed old2: %v", err)
	}

	w, err := NewRollingWriter(path, 1, 10, 1, false) // MaxAge=1 day
	if err != nil {
		t.Fatalf("NewRollingWriter: %v", err)
	}
	defer w.Close()

	// Trigger a rotation to invoke cleanup
	if _, err = w.Write(bytes.Repeat([]byte("A"), 900*kb)); err != nil {
		t.Fatalf("prefill: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err = w.Write(bytes.Repeat([]byte("B"), 300*kb)); err != nil {
		t.Fatalf("trigger rotate: %v", err)
	}

	if _, err := os.Stat(old1); !os.IsNotExist(err) {
		t.Fatalf("expected old1 pruned, stat err=%v", err)
	}
	if _, err := os.Stat(old2); !os.IsNotExist(err) {
		t.Fatalf("expected old2 pruned, stat err=%v", err)
	}
}

func TestConcurrentWritesNoRace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rw.log")

	w, err := NewRollingWriter(path, 1, 5, 7, false)
	if err != nil {
		t.Fatalf("NewRollingWriter: %v", err)
	}
	defer w.Close()

	var wg sync.WaitGroup
	errCh := make(chan error, 16)

	workers := 5
	iters := 50
	chunk := 16 * kb

	for g := 0; g < workers; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := bytes.Repeat([]byte("Z"), chunk)
			for i := 0; i < iters; i++ {
				if _, err := w.Write(buf); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Fatalf("write error: %v", e)
	}

	// Should have produced at least one segment or a sizeable active file
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("active file stat: %v", err)
	}
	_ = globSegments(t, path) // just ensure glob logic doesn't panic
}

// Helpers

func globSegments(t *testing.T, base string) []string {
	t.Helper()
	dir, ext := filepath.Dir(base), filepath.Ext(base)
	stem := strings.TrimSuffix(filepath.Base(base), ext)

	gl1 := filepath.Join(dir, stem+"-*"+ext)
	gl2 := filepath.Join(dir, stem+"-*"+ext+".gz")
	m1, _ := filepath.Glob(gl1)
	m2, _ := filepath.Glob(gl2)
	segs := append(m1, m2...)
	sort.Strings(segs)
	return segs
}
