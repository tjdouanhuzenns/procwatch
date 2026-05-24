package supervisor

import (
	"errors"
	"testing"
	"time"
)

func TestProcessTracing_RecordAndForProcess(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	if err := s.Record("web", "start", now, now.Add(10*time.Millisecond), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := s.ForProcess("web")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Operation != "start" {
		t.Errorf("expected operation 'start', got %q", entries[0].Operation)
	}
}

func TestProcessTracing_DurationIsSet(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	end := now.Add(50 * time.Millisecond)
	s.Record("api", "health_check", now, end, nil)
	entries := s.ForProcess("api")
	if entries[0].Duration != 50*time.Millisecond {
		t.Errorf("expected 50ms duration, got %v", entries[0].Duration)
	}
}

func TestProcessTracing_ErrorIsRecorded(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	s.Record("worker", "restart", now, now.Add(time.Millisecond), errors.New("exit code 1"))
	entries := s.ForProcess("worker")
	if entries[0].Error != "exit code 1" {
		t.Errorf("expected error string, got %q", entries[0].Error)
	}
}

func TestProcessTracing_EvictsOldestWhenFull(t *testing.T) {
	s := NewProcessTracingStore(3)
	now := time.Now()
	for i := 0; i < 4; i++ {
		s.Record("svc", "op", now, now.Add(time.Millisecond), nil)
	}
	if s.Len() != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", s.Len())
	}
}

func TestProcessTracing_ForProcessMissingReturnsEmpty(t *testing.T) {
	s := NewProcessTracingStore(16)
	if entries := s.ForProcess("ghost"); len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestProcessTracing_EmptyProcessErrors(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	if err := s.Record("", "op", now, now.Add(time.Millisecond), nil); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessTracing_EmptyOperationErrors(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	if err := s.Record("svc", "", now, now.Add(time.Millisecond), nil); err == nil {
		t.Error("expected error for empty operation")
	}
}

func TestProcessTracing_AllReturnsCopy(t *testing.T) {
	s := NewProcessTracingStore(16)
	now := time.Now()
	s.Record("a", "init", now, now.Add(time.Millisecond), nil)
	s.Record("b", "init", now, now.Add(time.Millisecond), nil)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	all[0].Process = "mutated"
	original := s.All()
	if original[0].Process == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
