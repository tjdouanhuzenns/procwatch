package supervisor

import (
	"testing"
	"time"
)

func TestProcessEventLog_RecordAndAll(t *testing.T) {
	l := NewProcessEventLog(10)
	if err := l.Record("svc", "started", "process started"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all := l.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 event, got %d", len(all))
	}
	if all[0].Process != "svc" || all[0].Kind != "started" {
		t.Errorf("unexpected event: %+v", all[0])
	}
}

func TestProcessEventLog_TimestampIsSet(t *testing.T) {
	l := NewProcessEventLog(10)
	before := time.Now()
	_ = l.Record("svc", "stopped", "")
	after := time.Now()
	all := l.All()
	if all[0].Timestamp.Before(before) || all[0].Timestamp.After(after) {
		t.Errorf("timestamp out of range: %v", all[0].Timestamp)
	}
}

func TestProcessEventLog_EvictsOldestWhenFull(t *testing.T) {
	l := NewProcessEventLog(3)
	_ = l.Record("svc", "a", "first")
	_ = l.Record("svc", "b", "second")
	_ = l.Record("svc", "c", "third")
	_ = l.Record("svc", "d", "fourth")
	all := l.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 events after eviction, got %d", len(all))
	}
	if all[0].Kind != "b" {
		t.Errorf("expected oldest to be evicted, got kind=%q", all[0].Kind)
	}
}

func TestProcessEventLog_ForProcess(t *testing.T) {
	l := NewProcessEventLog(20)
	_ = l.Record("alpha", "started", "")
	_ = l.Record("beta", "started", "")
	_ = l.Record("alpha", "stopped", "")
	events := l.ForProcess("alpha")
	if len(events) != 2 {
		t.Fatalf("expected 2 events for alpha, got %d", len(events))
	}
	for _, e := range events {
		if e.Process != "alpha" {
			t.Errorf("unexpected process in result: %q", e.Process)
		}
	}
}

func TestProcessEventLog_EmptyProcessErrors(t *testing.T) {
	l := NewProcessEventLog(10)
	if err := l.Record("", "started", ""); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessEventLog_EmptyKindErrors(t *testing.T) {
	l := NewProcessEventLog(10)
	if err := l.Record("svc", "", ""); err == nil {
		t.Error("expected error for empty kind")
	}
}

func TestProcessEventLog_Len(t *testing.T) {
	l := NewProcessEventLog(10)
	_ = l.Record("svc", "a", "")
	_ = l.Record("svc", "b", "")
	if l.Len() != 2 {
		t.Errorf("expected Len 2, got %d", l.Len())
	}
}

func TestProcessEventLog_DefaultMaxSize(t *testing.T) {
	l := NewProcessEventLog(0)
	if l.maxSize != 256 {
		t.Errorf("expected default maxSize 256, got %d", l.maxSize)
	}
}
