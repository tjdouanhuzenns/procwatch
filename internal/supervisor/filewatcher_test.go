package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fw-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestFileWatcher_EmitsOnModification(t *testing.T) {
	path := writeTempFile(t, "initial")

	cfg := FileWatcherConfig{PollInterval: 20 * time.Millisecond}
	fw := NewFileWatcher(cfg)
	if err := fw.Watch("myproc", path); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fw.Run(ctx)

	time.Sleep(30 * time.Millisecond)
	// Modify the file
	if err := os.WriteFile(path, []byte("updated"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	select {
	case ev := <-fw.Events():
		if ev.Path != path {
			t.Errorf("expected path %q, got %q", path, ev.Path)
		}
		if ev.Process != "myproc" {
			t.Errorf("expected process myproc, got %q", ev.Process)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for file event")
	}
}

func TestFileWatcher_NoEventWhenUnchanged(t *testing.T) {
	path := writeTempFile(t, "static")

	cfg := FileWatcherConfig{PollInterval: 20 * time.Millisecond}
	fw := NewFileWatcher(cfg)
	_ = fw.Watch("myproc", path)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fw.Run(ctx)

	select {
	case ev := <-fw.Events():
		t.Errorf("unexpected event: %+v", ev)
	case <-time.After(100 * time.Millisecond):
		// expected: no event
	}
}

func TestFileWatcher_UnwatchStopsEvents(t *testing.T) {
	path := writeTempFile(t, "hello")

	cfg := FileWatcherConfig{PollInterval: 20 * time.Millisecond}
	fw := NewFileWatcher(cfg)
	_ = fw.Watch("myproc", path)
	fw.Unwatch(path)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fw.Run(ctx)

	// Modify after unwatch
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(path, []byte("changed"), 0644)

	select {
	case ev := <-fw.Events():
		t.Errorf("unexpected event after unwatch: %+v", ev)
	case <-time.After(120 * time.Millisecond):
		// expected
	}
}

func TestFileWatcher_EmptyProcessErrors(t *testing.T) {
	fw := NewFileWatcher(DefaultFileWatcherConfig())
	if err := fw.Watch("", "/some/path"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestFileWatcher_EmptyPathErrors(t *testing.T) {
	fw := NewFileWatcher(DefaultFileWatcherConfig())
	if err := fw.Watch("myproc", ""); err == nil {
		t.Error("expected error for empty path")
	}
}

func TestFileWatcher_ClosesChannelOnCancel(t *testing.T) {
	fw := NewFileWatcher(FileWatcherConfig{PollInterval: 20 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		fw.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Run did not return after context cancel")
	}
	_ = filepath.Join(".") // suppress unused import
}
