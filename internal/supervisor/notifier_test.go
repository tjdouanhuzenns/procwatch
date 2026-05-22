package supervisor

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNotifier_NoHandlers(t *testing.T) {
	n := NewNotifier()
	errs := n.Notify(context.Background(), NotifyEvent{Process: "p", State: "running"})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}

func TestNotifier_HandlerReceivesEvent(t *testing.T) {
	n := NewNotifier()
	var got NotifyEvent
	n.Register(func(_ context.Context, e NotifyEvent) error {
		got = e
		return nil
	})

	evt := NotifyEvent{Process: "svc", State: "failed", Message: "exit 1"}
	n.Notify(context.Background(), evt)

	if got.Process != "svc" || got.State != "failed" {
		t.Errorf("unexpected event: %+v", got)
	}
}

func TestNotifier_TimestampSetIfZero(t *testing.T) {
	n := NewNotifier()
	var ts time.Time
	n.Register(func(_ context.Context, e NotifyEvent) error {
		ts = e.Timestamp
		return nil
	})
	n.Notify(context.Background(), NotifyEvent{Process: "p", State: "stopped"})
	if ts.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestNotifier_MultipleHandlers(t *testing.T) {
	n := NewNotifier()
	var count atomic.Int32
	for i := 0; i < 3; i++ {
		n.Register(func(_ context.Context, _ NotifyEvent) error {
			count.Add(1)
			return nil
		})
	}
	n.Notify(context.Background(), NotifyEvent{Process: "p", State: "running"})
	if count.Load() != 3 {
		t.Errorf("expected 3 handler calls, got %d", count.Load())
	}
}

func TestNotifier_ErrorsCollected(t *testing.T) {
	n := NewNotifier()
	n.Register(func(_ context.Context, _ NotifyEvent) error { return errors.New("handler error") })
	n.Register(func(_ context.Context, _ NotifyEvent) error { return nil })
	n.Register(func(_ context.Context, _ NotifyEvent) error { return errors.New("another error") })

	errs := n.Notify(context.Background(), NotifyEvent{Process: "p", State: "failed"})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}

func TestNotifier_Len(t *testing.T) {
	n := NewNotifier()
	if n.Len() != 0 {
		t.Fatal("expected 0 handlers")
	}
	n.Register(func(_ context.Context, _ NotifyEvent) error { return nil })
	n.Register(func(_ context.Context, _ NotifyEvent) error { return nil })
	if n.Len() != 2 {
		t.Errorf("expected 2 handlers, got %d", n.Len())
	}
}
