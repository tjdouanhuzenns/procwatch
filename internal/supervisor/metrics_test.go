package supervisor

import (
	"testing"
	"time"
)

func TestMetrics_RecordStartIncrementsRestarts(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordStart("svc")
	mc.RecordExit("svc")
	mc.RecordStart("svc")
	mc.RecordExit("svc")
	mc.RecordStart("svc")

	m, ok := mc.Get("svc")
	if !ok {
		t.Fatal("expected metrics for svc")
	}
	if m.Restarts != 2 {
		t.Errorf("expected 2 restarts, got %d", m.Restarts)
	}
}

func TestMetrics_RecordExitAccumulatesUptime(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordStart("svc")
	time.Sleep(20 * time.Millisecond)
	mc.RecordExit("svc")

	m, ok := mc.Get("svc")
	if !ok {
		t.Fatal("expected metrics for svc")
	}
	if m.TotalUptime < 10*time.Millisecond {
		t.Errorf("expected TotalUptime >= 10ms, got %v", m.TotalUptime)
	}
}

func TestMetrics_UptimeWhileRunning(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordStart("svc")
	time.Sleep(10 * time.Millisecond)

	m, ok := mc.Get("svc")
	if !ok {
		t.Fatal("expected metrics for svc")
	}
	if m.Uptime < 5*time.Millisecond {
		t.Errorf("expected live Uptime >= 5ms, got %v", m.Uptime)
	}
}

func TestMetrics_GetMissing(t *testing.T) {
	mc := NewMetricsCollector()
	_, ok := mc.Get("nonexistent")
	if ok {
		t.Error("expected false for missing process")
	}
}

func TestMetrics_All(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordStart("a")
	mc.RecordStart("b")

	all := mc.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
