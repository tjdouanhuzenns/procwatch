package supervisor

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestDrainer_ConcurrentReleases verifies that releasing many processes
// concurrently does not cause races or panics.
func TestDrainer_ConcurrentReleases(t *testing.T) {
	cfg := DrainConfig{Timeout: 3 * time.Second, PollInterval: 10 * time.Millisecond}
	d := NewDrainer(cfg)

	names := []string{"svc-a", "svc-b", "svc-c", "svc-d", "svc-e"}
	for _, n := range names {
		d.Track(n)
	}

	var wg sync.WaitGroup
	for _, n := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			d.Release(name)
		}(n)
	}

	ctx := context.Background()
	if !d.Wait(ctx) {
		t.Fatal("expected all concurrent releases to drain")
	}
	wg.Wait()
}

// TestDrainer_ReleaseUnknownIsNoop verifies releasing an untracked name
// does not break the drainer.
func TestDrainer_ReleaseUnknownIsNoop(t *testing.T) {
	cfg := DefaultDrainConfig()
	d := NewDrainer(cfg)
	d.Track("known")
	d.Release("unknown") // should be a no-op

	if d.ActiveCount() != 1 {
		t.Fatalf("expected 1 active after releasing unknown, got %d", d.ActiveCount())
	}

	d.Release("known")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !d.Wait(ctx) {
		t.Fatal("expected drain to succeed")
	}
}
