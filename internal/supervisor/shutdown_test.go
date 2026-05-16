package supervisor

import (
	"bytes"
	"context"
	"log/slog"
	"syscall"
	"testing"
	"time"
)

func newShutdownLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestShutdownHandler_CancelsOnSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	drainer := NewDrainer(DefaultDrainConfig())
	var buf bytes.Buffer
	log := newShutdownLogger(&buf)

	h := NewShutdownHandler(cancel, drainer, log)

	done := make(chan struct{})
	go func() {
		h.Listen(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM) //nolint:errcheck

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown handler did not complete after SIGTERM")
	}

	select {
	case <-ctx.Done():
		// context was cancelled as expected
	default:
		t.Fatal("context was not cancelled after signal")
	}
}

func TestShutdownHandler_ExitsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	drainer := NewDrainer(DefaultDrainConfig())
	var buf bytes.Buffer
	log := newShutdownLogger(&buf)

	h := NewShutdownHandler(cancel, drainer, log)

	done := make(chan struct{})
	go func() {
		h.Listen(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("shutdown handler did not exit after context cancel")
	}
}
