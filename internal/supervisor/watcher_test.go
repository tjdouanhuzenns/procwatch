package supervisor

import (
	"context"
	"testing"
	"time"
)

func TestWatcher_EmitsOnNewProcess(t *testing.T) {
	reporter := NewStatusReporter()
	reporter.Update("alpha", ProcessStatus{State: "running"})

	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	select {
	case ev := <-w.Events():
		if ev.ProcessName != "alpha" {
			t.Fatalf("expected alpha, got %s", ev.ProcessName)
		}
		if ev.NewStatus != "running" {
			t.Fatalf("expected running, got %s", ev.NewStatus)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestWatcher_EmitsOnStatusChange(t *testing.T) {
	reporter := NewStatusReporter()
	reporter.Update("beta", ProcessStatus{State: "running"})

	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	// Drain the initial event.
	select {
	case <-w.Events():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for initial event")
	}

	reporter.Update("beta", ProcessStatus{State: "stopped"})

	select {
	case ev := <-w.Events():
		if ev.OldStatus != "running" || ev.NewStatus != "stopped" {
			t.Fatalf("unexpected transition: %s -> %s", ev.OldStatus, ev.NewStatus)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for change event")
	}
}

func TestWatcher_NoEventWhenUnchanged(t *testing.T) {
	reporter := NewStatusReporter()
	reporter.Update("gamma", ProcessStatus{State: "running"})

	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	// Drain first event.
	select {
	case <-w.Events():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for initial event")
	}

	// No further events expected since state is unchanged.
	select {
	case ev := <-w.Events():
		t.Fatalf("unexpected event: %+v", ev)
	case <-time.After(100 * time.Millisecond):
		// pass
	}
}

func TestWatcher_ClosesChannelOnCancel(t *testing.T) {
	reporter := NewStatusReporter()
	w := NewWatcher(reporter, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go w.Run(ctx)
	cancel()

	select {
	case _, ok := <-w.Events():
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}
