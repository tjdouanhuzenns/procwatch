package supervisor_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/user/procwatch/internal/config"
	"github.com/user/procwatch/internal/logger"
	"github.com/user/procwatch/internal/supervisor"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.New(io.Discard, "info")
}

func TestSupervisor_StartsProcesses(t *testing.T) {
	cfg := &config.Config{
		Processes: []config.ProcessConfig{
			{Name: "echo1", Command: "echo", Args: []string{"hello"}, Restart: "never"},
			{Name: "echo2", Command: "echo", Args: []string{"world"}, Restart: "never"},
		},
	}
	log := newTestLogger(t)
	sv := supervisor.New(cfg, log)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := sv.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	procs := sv.Processes()
	if len(procs) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(procs))
	}
}

func TestSupervisor_CancelStopsAll(t *testing.T) {
	cfg := &config.Config{
		Processes: []config.ProcessConfig{
			{Name: "sleep1", Command: "sleep", Args: []string{"30"}, Restart: "never"},
		},
	}
	log := newTestLogger(t)
	sv := supervisor.New(cfg, log)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- sv.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("supervisor did not stop in time")
	}
}
