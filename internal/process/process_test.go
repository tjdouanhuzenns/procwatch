package process

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/procwatch/internal/logger"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, err := logger.New(os.Stdout, "debug")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	return l
}

func TestProcess_StartsAndExits(t *testing.T) {
	log := newTestLogger(t)
	rt := NewRestartTracker(DefaultRestartConfig())
	p := New("echo", "echo", []string{"hello"}, nil, log, rt)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p.Start(ctx)

	// Give the process time to start and exit
	time.Sleep(300 * time.Millisecond)

	state := p.CurrentState()
	if state == StateStarting {
		t.Error("process should have progressed past StateStarting")
	}
}

func TestProcess_StopKillsProcess(t *testing.T) {
	log := newTestLogger(t)
	cfg := DefaultRestartConfig()
	cfg.Policy = "never"
	rt := NewRestartTracker(cfg)
	p := New("sleep", "sleep", []string{"30"}, nil, log, rt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	if p.CurrentState() != StateRunning {
		t.Fatalf("expected StateRunning, got %v", p.CurrentState())
	}

	done := make(chan struct{})
	go func() {
		p.Stop()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() timed out")
	}

	if p.CurrentState() != StateStopped && p.CurrentState() != StateFailed {
		t.Errorf("expected stopped or failed state after Stop(), got %v", p.CurrentState())
	}
}

func TestProcess_ContextCancellationStops(t *testing.T) {
	log := newTestLogger(t)
	cfg := DefaultRestartConfig()
	cfg.Policy = "always"
	rt := NewRestartTracker(cfg)
	p := New("sleep", "sleep", []string{"30"}, nil, log, rt)

	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case <-p.doneCh:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("process did not stop after context cancellation")
	}
}
