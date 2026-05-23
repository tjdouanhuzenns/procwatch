package supervisor

import (
	"testing"
	"time"
)

func TestProcessHealthHistory_RecordAndForProcess(t *testing.T) {
	h := NewProcessHealthHistory(10)
	if err := h.Record("svc", true, "ok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := h.ForProcess("svc")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].Healthy {
		t.Error("expected healthy=true")
	}
	if events[0].Message != "ok" {
		t.Errorf("expected message 'ok', got %q", events[0].Message)
	}
}

func TestProcessHealthHistory_TimestampIsSet(t *testing.T) {
	h := NewProcessHealthHistory(10)
	before := time.Now()
	_ = h.Record("svc", false, "timeout")
	after := time.Now()
	events := h.ForProcess("svc")
	if events[0].Timestamp.Before(before) || events[0].Timestamp.After(after) {
		t.Error("timestamp out of expected range")
	}
}

func TestProcessHealthHistory_EvictsOldestWhenFull(t *testing.T) {
	h := NewProcessHealthHistory(3)
	for i := 0; i < 5; i++ {
		msg := "event"
		if i == 0 {
			msg = "first"
		}
		_ = h.Record("svc", true, msg)
	}
	events := h.ForProcess("svc")
	if len(events) != 3 {
		t.Fatalf("expected 3 events after eviction, got %d", len(events))
	}
	for _, ev := range events {
		if ev.Message == "first" {
			t.Error("oldest event should have been evicted")
		}
	}
}

func TestProcessHealthHistory_EmptyProcessErrors(t *testing.T) {
	h := NewProcessHealthHistory(10)
	if err := h.Record("", true, "ok"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessHealthHistory_ForProcessMissingReturnsEmpty(t *testing.T) {
	h := NewProcessHealthHistory(10)
	events := h.ForProcess("missing")
	if len(events) != 0 {
		t.Errorf("expected empty slice, got %d events", len(events))
	}
}

func TestProcessHealthHistory_Clear(t *testing.T) {
	h := NewProcessHealthHistory(10)
	_ = h.Record("svc", true, "ok")
	h.Clear("svc")
	if h.Len("svc") != 0 {
		t.Error("expected history to be cleared")
	}
}

func TestProcessHealthHistory_ForProcessReturnsCopy(t *testing.T) {
	h := NewProcessHealthHistory(10)
	_ = h.Record("svc", true, "ok")
	events := h.ForProcess("svc")
	events[0].Message = "mutated"
	original := h.ForProcess("svc")
	if original[0].Message == "mutated" {
		t.Error("ForProcess should return a copy, not a reference")
	}
}
