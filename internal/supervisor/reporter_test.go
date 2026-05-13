package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/user/procwatch/internal/logger"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func TestHTTPReporter_StatusEndpoint(t *testing.T) {
	addr := freePort(t)
	sr := NewStatusReporter()
	sr.Update("svc", ProcessStatus{Name: "svc", State: "running"})
	mc := NewMetricsCollector()
	log := logger.New(nil, "info")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rep := NewHTTPReporter(addr, sr, mc, log)
	rep.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var out []ProcessStatus
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].Name != "svc" {
		t.Errorf("unexpected status output: %+v", out)
	}
}

func TestHTTPReporter_MetricsEndpoint(t *testing.T) {
	addr := freePort(t)
	sr := NewStatusReporter()
	mc := NewMetricsCollector()
	mc.RecordStart("worker")
	log := logger.New(nil, "info")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rep := NewHTTPReporter(addr, sr, mc, log)
	rep.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	var out []ProcessMetrics
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].Name != "worker" {
		t.Errorf("unexpected metrics output: %+v", out)
	}
}
