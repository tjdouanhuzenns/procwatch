package supervisor

import (
	"fmt"
	"testing"
	"time"
)

func TestEventLog_RecordAndAll(t *testing.T) {
	el := NewEventLog(10)
	el.Record("svc-a", EventStarted, "process started")
	el.Record("svc-b", EventFailed, "exit code 1")

	events := el.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Process != "svc-a" || events[0].Kind != EventStarted {
		t.Errorf("unexpected first event: %+v", events[0])
	}
	if events[1].Process != "svc-b" || events[1].Kind != EventFailed {
		t.Errorf("unexpected second event: %+v", events[1])
	}
}

func TestEventLog_TimestampIsSet(t *testing.T) {
	before := time.Now()
	el := NewEventLog(10)
	el.Record("svc", EventStarted, "")
	after := time.Now()

	events := el.All()
	if len(events) != 1 {
		t.Fatal("expected 1 event")
	}
	ts := events[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestEventLog_EvictsOldestWhenFull(t *testing.T) {
	el := NewEventLog(3)
	for i := 0; i < 5; i++ {
		el.Record("svc", EventStarted, fmt.Sprintf("msg-%d", i))
	}

	events := el.All()
	if len(events) != 3 {
		t.Fatalf("expected 3 events after eviction, got %d", len(events))
	}
	// Oldest two (msg-0, msg-1) should have been evicted.
	if events[0].Message != "msg-2" {
		t.Errorf("expected msg-2 as oldest retained, got %q", events[0].Message)
	}
	if events[2].Message != "msg-4" {
		t.Errorf("expected msg-4 as newest, got %q", events[2].Message)
	}
}

func TestEventLog_ForProcess(t *testing.T) {
	el := NewEventLog(20)
	el.Record("alpha", EventStarted, "")
	el.Record("beta", EventFailed, "")
	el.Record("alpha", EventRestarted, "")
	el.Record("beta", EventStopped, "")

	alpha := el.ForProcess("alpha")
	if len(alpha) != 2 {
		t.Fatalf("expected 2 events for alpha, got %d", len(alpha))
	}
	for _, ev := range alpha {
		if ev.Process != "alpha" {
			t.Errorf("unexpected process in result: %q", ev.Process)
		}
	}
}

func TestEventLog_Len(t *testing.T) {
	el := NewEventLog(10)
	if el.Len() != 0 {
		t.Fatalf("expected 0, got %d", el.Len())
	}
	el.Record("svc", EventStarted, "")
	el.Record("svc", EventStopped, "")
	if el.Len() != 2 {
		t.Fatalf("expected 2, got %d", el.Len())
	}
}

func TestEventLog_DefaultCapWhenZero(t *testing.T) {
	el := NewEventLog(0)
	if el.cap != 256 {
		t.Errorf("expected default cap 256, got %d", el.cap)
	}
}

func TestEventLog_AllReturnsCopy(t *testing.T) {
	el := NewEventLog(10)
	el.Record("svc", EventStarted, "original")

	snap := el.All()
	snap[0].Message = "mutated"

	fresh := el.All()
	if fresh[0].Message != "original" {
		t.Errorf("All() should return a copy; got %q", fresh[0].Message)
	}
}
