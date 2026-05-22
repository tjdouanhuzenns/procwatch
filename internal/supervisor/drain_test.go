package supervisor

import (
	"context"
	"testing"
	"time"
)

func TestDrainer_AllReleasedBeforeTimeout(t *testing.T) {
	cfg := DrainConfig{Timeout: 2 * time.Second, PollInterval: 10 * time.Millisecond}
	d := NewDrainer(cfg)
	d.Track("web")
	d.Track("worker")

	go func() {
		time.Sleep(50 * time.Millisecond)
		d.Release("web")
		d.Release("worker")
	}()

	ctx := context.Background()
	if !d.Wait(ctx) {
		t.Fatal("expected drain to succeed")
	}
}

func TestDrainer_TimesOut(t *testing.T) {
	cfg := DrainConfig{Timeout: 50 * time.Millisecond, PollInterval: 10 * time.Millisecond}
	d := NewDrainer(cfg)
	d.Track("stuck")

	ctx := context.Background()
	if d.Wait(ctx) {
		t.Fatal("expected drain to time out")
	}
}

func TestDrainer_ContextCancellation(t *testing.T) {
	cfg := DrainConfig{Timeout: 5 * time.Second, PollInterval: 10 * time.Millisecond}
	d := NewDrainer(cfg)
	d.Track("web")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	if d.Wait(ctx) {
		t.Fatal("expected drain to fail due to context cancel")
	}
}

func TestDrainer_ActiveCount(t *testing.T) {
	cfg := DefaultDrainConfig()
	d := NewDrainer(cfg)

	if d.ActiveCount() != 0 {
		t.Fatalf("expected 0 active, got %d", d.ActiveCount())
	}
	d.Track("a")
	d.Track("b")
	if d.ActiveCount() != 2 {
		t.Fatalf("expected 2 active, got %d", d.ActiveCount())
	}
	d.Release("a")
	if d.ActiveCount() != 1 {
		t.Fatalf("expected 1 active, got %d", d.ActiveCount())
	}
}

func TestDrainer_EmptyDrainImmediately(t *testing.T) {
	cfg := DefaultDrainConfig()
	d := NewDrainer(cfg)
	// No processes tracked — done channel is never closed so poll fires.
	// Expect immediate success via poll.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if !d.Wait(ctx) {
		t.Fatal("expected empty drainer to drain immediately")
	}
}

// TestDrainer_ReleaseUntracked verifies that releasing a name that was never
// tracked does not affect the active count or cause a panic.
func TestDrainer_ReleaseUntracked(t *testing.T) {
	cfg := DefaultDrainConfig()
	d := NewDrainer(cfg)
	d.Track("a")

	// Release a name that was never tracked; count should remain 1.
	d.Release("unknown")
	if d.ActiveCount() != 1 {
		t.Fatalf("expected 1 active after releasing untracked name, got %d", d.ActiveCount())
	}
}

func TestDefaultDrainConfig(t *testing.T) {
	cfg := DefaultDrainConfig()
	if cfg.Timeout <= 0 {
		t.Fatal("expected positive timeout")
	}
	if cfg.PollInterval <= 0 {
		t.Fatal("expected positive poll interval")
	}
}
