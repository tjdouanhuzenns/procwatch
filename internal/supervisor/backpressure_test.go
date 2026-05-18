package supervisor

import (
	"context"
	"sync"
	"testing"
	"time"
)

func testBPConfig() BackpressureConfig {
	return BackpressureConfig{
		MaxPending:  3,
		WaitTimeout: 100 * time.Millisecond,
	}
}

func TestBackpressure_AcquireUnderLimit(t *testing.T) {
	bp := NewBackpressure(testBPConfig())
	ctx := context.Background()
	if !bp.Acquire(ctx) {
		t.Fatal("expected acquire to succeed under limit")
	}
	if bp.Pending() != 1 {
		t.Fatalf("expected 1 pending, got %d", bp.Pending())
	}
}

func TestBackpressure_BlocksAtLimit(t *testing.T) {
	bp := NewBackpressure(testBPConfig())
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if !bp.Acquire(ctx) {
			t.Fatalf("expected acquire %d to succeed", i)
		}
	}
	if bp.Acquire(ctx) {
		t.Fatal("expected acquire to fail at limit")
	}
}

func TestBackpressure_ReleaseFreesSlot(t *testing.T) {
	bp := NewBackpressure(testBPConfig())
	ctx := context.Background()
	bp.Acquire(ctx)
	bp.Acquire(ctx)
	bp.Acquire(ctx)
	bp.Release()
	if !bp.Acquire(ctx) {
		t.Fatal("expected acquire to succeed after release")
	}
}

func TestBackpressure_ContextCancellation(t *testing.T) {
	bp := NewBackpressure(testBPConfig())
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 3; i++ {
		bp.Acquire(ctx)
	}
	cancel()
	if bp.Acquire(ctx) {
		t.Fatal("expected acquire to fail on cancelled context")
	}
}

func TestBackpressure_Reset(t *testing.T) {
	bp := NewBackpressure(testBPConfig())
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		bp.Acquire(ctx)
	}
	bp.Reset()
	if bp.Pending() != 0 {
		t.Fatalf("expected 0 pending after reset, got %d", bp.Pending())
	}
}

func TestBackpressure_ConcurrentAcquire(t *testing.T) {
	bp := NewBackpressure(BackpressureConfig{MaxPending: 50, WaitTimeout: 200 * time.Millisecond})
	ctx := context.Background()
	var wg sync.WaitGroup
	succeeded := make(chan struct{}, 100)
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bp.Acquire(ctx) {
				succeeded <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(succeeded)
	count := 0
	for range succeeded {
		count++
	}
	if count > 50 {
		t.Fatalf("expected at most 50 successes, got %d", count)
	}
}
