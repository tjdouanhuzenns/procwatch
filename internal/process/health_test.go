package process

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHealthChecker_PassesOnOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var failed atomic.Bool
	cfg := HealthCheck{
		URL:      server.URL,
		Interval: 50 * time.Millisecond,
		Timeout:  500 * time.Millisecond,
		Retries:  1,
	}
	checker := NewHealthChecker(cfg, func(err error) { failed.Store(true) })

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	checker.Run(ctx)

	if failed.Load() {
		t.Error("expected no health check failure for 200 OK")
	}
}

func TestHealthChecker_FailsOn500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var failCount atomic.Int32
	cfg := HealthCheck{
		URL:      server.URL,
		Interval: 50 * time.Millisecond,
		Timeout:  500 * time.Millisecond,
		Retries:  2,
	}
	checker := NewHealthChecker(cfg, func(err error) { failCount.Add(1) })

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()
	checker.Run(ctx)

	if failCount.Load() == 0 {
		t.Error("expected at least one health check failure for 500 response")
	}
}

func TestHealthChecker_FailsOnUnreachable(t *testing.T) {
	var failCount atomic.Int32
	cfg := HealthCheck{
		URL:      "http://127.0.0.1:1", // nothing listening
		Interval: 50 * time.Millisecond,
		Timeout:  100 * time.Millisecond,
		Retries:  1,
	}
	checker := NewHealthChecker(cfg, func(err error) { failCount.Add(1) })

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()
	checker.Run(ctx)

	if failCount.Load() == 0 {
		t.Error("expected health check failure for unreachable host")
	}
}

func TestDefaultHealthCheck(t *testing.T) {
	d := DefaultHealthCheck()
	if d.Interval != 10*time.Second {
		t.Errorf("expected interval 10s, got %v", d.Interval)
	}
	if d.Timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %v", d.Timeout)
	}
	if d.Retries != 3 {
		t.Errorf("expected retries 3, got %d", d.Retries)
	}
}
