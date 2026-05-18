package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestBackpressure_ProducerConsumer simulates a producer-consumer pattern where
// restarts are produced faster than they are consumed, verifying backpressure holds.
func TestBackpressure_ProducerConsumer(t *testing.T) {
	cfg := BackpressureConfig{
		MaxPending:  5,
		WaitTimeout: 50 * time.Millisecond,
	}
	bp := NewBackpressure(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var accepted, rejected int64
	var wg sync.WaitGroup

	// Consumer: releases one slot every 30ms.
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bp.Release()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Producers: 10 goroutines each trying to acquire every 10ms.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if bp.Acquire(ctx) {
						atomic.AddInt64(&accepted, 1)
					} else {
						atomic.AddInt64(&rejected, 1)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt64(&rejected) == 0 {
		t.Error("expected some requests to be rejected under backpressure")
	}
	if atomic.LoadInt64(&accepted) == 0 {
		t.Error("expected some requests to be accepted")
	}
}
