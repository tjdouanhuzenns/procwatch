package supervisor

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestHTTPReporter_StatusEndpoint(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	sr := NewStatusReporter()
	sr.Update(ProcessStatus{Name: "svc", State: "running"})

	rep := NewHTTPReporter(addr, sr, NewMetricsCollector(), NewHealthReporter())
	go rep.ListenAndServe()
	defer rep.Close()
	time.Sleep(20 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/status")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var statuses []ProcessStatus
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Name != "svc" {
		t.Errorf("unexpected statuses: %+v", statuses)
	}
}

func TestHTTPReporter_MetricsEndpoint(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	mc := NewMetricsCollector()
	mc.RecordStart("svc")

	rep := NewHTTPReporter(addr, NewStatusReporter(), mc, NewHealthReporter())
	go rep.ListenAndServe()
	defer rep.Close()
	time.Sleep(20 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/metrics")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var all []ProcessMetrics
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(all) != 1 || all[0].Name != "svc" {
		t.Errorf("unexpected metrics: %+v", all)
	}
}

func TestHTTPReporter_HealthEndpoint(t *testing.T) {
	port := freePort(t)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	hr := NewHealthReporter()
	hr.Record("svc", HealthHealthy, "ok")

	rep := NewHTTPReporter(addr, NewStatusReporter(), NewMetricsCollector(), hr)
	go rep.ListenAndServe()
	defer rep.Close()
	time.Sleep(20 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var reports []HealthReport
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(reports) != 1 || reports[0].Process != "svc" {
		t.Errorf("unexpected health reports: %+v", reports)
	}
}
