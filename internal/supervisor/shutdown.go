package supervisor

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// ShutdownHandler listens for OS signals and triggers a graceful drain
// before cancelling the root context.
type ShutdownHandler struct {
	cancel  context.CancelFunc
	drainer *Drainer
	log     *slog.Logger
}

// NewShutdownHandler creates a ShutdownHandler.
func NewShutdownHandler(cancel context.CancelFunc, drainer *Drainer, log *slog.Logger) *ShutdownHandler {
	return &ShutdownHandler{
		cancel:  cancel,
		drainer: drainer,
		log:     log,
	}
}

// Listen blocks until SIGINT or SIGTERM is received, then drains and
// cancels the context. It should be run in its own goroutine.
func (s *ShutdownHandler) Listen(ctx context.Context) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	select {
	case sig := <-sigs:
		s.log.Info("shutdown signal received", "signal", sig.String())
		s.log.Info("draining processes", "active", s.drainer.ActiveCount())
		if s.drainer.Wait(ctx) {
			s.log.Info("all processes drained cleanly")
		} else {
			s.log.Warn("drain timed out, forcing shutdown",
				"remaining", s.drainer.ActiveCount())
		}
		s.cancel()
	case <-ctx.Done():
	}
}
