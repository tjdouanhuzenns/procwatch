package supervisor

import (
	"testing"
	"time"
)

func TestHealthReporter_RecordAndGet(t *testing.T) {
	r := NewHealthReporter()
	r.Record("svc", HealthHealthy, "ok")

	rep, ok := r.Get("svc")
	if !ok {
		t.Fatal("expected report to exist")
	}
	if rep.Status != HealthHealthy {
		t.Errorf("expected healthy, got %v", rep.Status)
	}
	if rep.Message != "ok" {
		t.Errorf("expected message 'ok', got %q", rep.Message)
	}
	if rep.Consecutive != 1 {
		t.Errorf("expected consecutive=1, got %d", rep.Consecutive)
	}
}

func TestHealthReporter_ConsecutiveIncrements(t *testing.T) {
	r := NewHealthReporter()
	r.Record("svc", HealthUnhealthy, "fail")
	r.Record("svc", HealthUnhealthy, "fail")
	r.Record("svc", HealthUnhealthy, "fail")

	rep, _ := r.Get("svc")
	if rep.Consecutive != 3 {
		t.Errorf("expected consecutive=3, got %d", rep.Consecutive)
	}
}

func TestHealthReporter_ConsecutiveResetsOnStatusChange(t *testing.T) {
	r := NewHealthReporter()
	r.Record("svc", HealthUnhealthy, "fail")
	r.Record("svc", HealthUnhealthy, "fail")
	r.Record("svc", HealthHealthy, "ok")

	rep, _ := r.Get("svc")
	if rep.Consecutive != 1 {
		t.Errorf("expected consecutive reset to 1, got %d", rep.Consecutive)
	}
}

func TestHealthReporter_GetMissing(t *testing.T) {
	r := NewHealthReporter()
	_, ok := r.Get("nope")
	if ok {
		t.Error("expected missing report to return false")
	}
}

func TestHealthReporter_All(t *testing.T) {
	r := NewHealthReporter()
	r.Record("a", HealthHealthy, "")
	r.Record("b", HealthDegraded, "")
	r.Record("c", HealthUnhealthy, "")

	all := r.All()
	if len(all) != 3 {
		t.Errorf("expected 3 reports, got %d", len(all))
	}
}

func TestHealthReporter_Remove(t *testing.T) {
	r := NewHealthReporter()
	r.Record("svc", HealthHealthy, "ok")
	r.Remove("svc")
	_, ok := r.Get("svc")
	if ok {
		t.Error("expected report to be removed")
	}
}

func TestHealthReporter_TimestampIsSet(t *testing.T) {
	before := time.Now()
	r := NewHealthReporter()
	r.Record("svc", HealthHealthy, "ok")
	after := time.Now()

	rep, _ := r.Get("svc")
	if rep.LastChecked.Before(before) || rep.LastChecked.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", rep.LastChecked, before, after)
	}
}

func TestHealthStatus_String(t *testing.T) {
	cases := []struct {
		status HealthStatus
		want   string
	}{
		{HealthHealthy, "healthy"},
		{HealthDegraded, "degraded"},
		{HealthUnhealthy, "unhealthy"},
		{HealthUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("status %d: expected %q, got %q", tc.status, tc.want, got)
		}
	}
}
