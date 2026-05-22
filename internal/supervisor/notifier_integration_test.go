package supervisor

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNotifier_ConcurrentRegisterAndNotify(t *testing.T) {
	n := NewNotifier()
	var wg sync.WaitGroup
	var callCount atomic.Int32

	// Register handlers concurrently.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.Register(func(_ context.Context, _ NotifyEvent) error {
				callCount.Add(1)
				return nil
			})
		}()
	}
	wg.Wait()

	// Notify once — all 10 handlers should fire.
	n.Notify(context.Background(), NotifyEvent{Process: "svc", State: "running"})
	if callCount.Load() != 10 {
		t.Errorf("expected 10 calls, got %d", callCount.Load())
	}
}

func TestNotifier_ConcurrentNotify(t *testing.T) {
	n := NewNotifier()
	var total atomic.Int32
	n.Register(func(_ context.Context, _ NotifyEvent) error {
		total.Add(1)
		return nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.Notify(context.Background(), NotifyEvent{Process: "p", State: "stopped"})
		}()
	}
	wg.Wait()

	if total.Load() != 50 {
		t.Errorf("expected 50 notifications, got %d", total.Load())
	}
}
