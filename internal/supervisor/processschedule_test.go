package supervisor

import (
	"testing"
	"time"
)

func TestProcessSchedule_SetAndGet(t *testing.T) {
	ps := NewProcessSchedule()
	next := time.Now().Add(time.Minute)
	if err := ps.Set("worker", "* * * * *", next); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := ps.Get("worker")
	if !ok {
		t.Fatal("expected schedule to exist")
	}
	if s.Cron != "* * * * *" {
		t.Errorf("unexpected cron: %s", s.Cron)
	}
	if !s.Enabled {
		t.Error("expected schedule to be enabled by default")
	}
}

func TestProcessSchedule_GetMissing(t *testing.T) {
	ps := NewProcessSchedule()
	_, ok := ps.Get("ghost")
	if ok {
		t.Error("expected missing schedule to return false")
	}
}

func TestProcessSchedule_SetEmptyNameErrors(t *testing.T) {
	ps := NewProcessSchedule()
	if err := ps.Set("", "* * * * *", time.Now()); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessSchedule_SetEmptyCronErrors(t *testing.T) {
	ps := NewProcessSchedule()
	if err := ps.Set("worker", "", time.Now()); err == nil {
		t.Error("expected error for empty cron expression")
	}
}

func TestProcessSchedule_RecordRun(t *testing.T) {
	ps := NewProcessSchedule()
	_ = ps.Set("worker", "* * * * *", time.Now())
	next := time.Now().Add(time.Minute)
	if err := ps.RecordRun("worker", next); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, _ := ps.Get("worker")
	if s.LastRun.IsZero() {
		t.Error("expected LastRun to be set")
	}
	if !s.NextRun.Equal(next) {
		t.Errorf("unexpected NextRun: %v", s.NextRun)
	}
}

func TestProcessSchedule_RecordRunMissing(t *testing.T) {
	ps := NewProcessSchedule()
	if err := ps.RecordRun("ghost", time.Now()); err == nil {
		t.Error("expected error for missing process")
	}
}

func TestProcessSchedule_SetEnabled(t *testing.T) {
	ps := NewProcessSchedule()
	_ = ps.Set("worker", "* * * * *", time.Now())
	if err := ps.SetEnabled("worker", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, _ := ps.Get("worker")
	if s.Enabled {
		t.Error("expected schedule to be disabled")
	}
}

func TestProcessSchedule_Remove(t *testing.T) {
	ps := NewProcessSchedule()
	_ = ps.Set("worker", "* * * * *", time.Now())
	ps.Remove("worker")
	_, ok := ps.Get("worker")
	if ok {
		t.Error("expected schedule to be removed")
	}
}

func TestProcessSchedule_All(t *testing.T) {
	ps := NewProcessSchedule()
	_ = ps.Set("a", "* * * * *", time.Now())
	_ = ps.Set("b", "0 * * * *", time.Now())
	all := ps.All()
	if len(all) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(all))
	}
}
