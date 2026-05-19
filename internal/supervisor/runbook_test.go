package supervisor

import (
	"fmt"
	"testing"
	"time"
)

func TestRunbook_RecordAndAll(t *testing.T) {
	rb := NewRunbook(10)
	rb.Record("svc-a", "crash", "OOM killed")
	rb.Record("svc-b", "restart", "manual restart")

	all := rb.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Process != "svc-a" || all[0].Event != "crash" {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
}

func TestRunbook_TimestampIsSet(t *testing.T) {
	before := time.Now().UTC()
	rb := NewRunbook(10)
	rb.Record("svc", "start", "initial start")
	after := time.Now().UTC()

	all := rb.All()
	ts := all[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", ts, before, after)
	}
}

func TestRunbook_EvictsOldestWhenFull(t *testing.T) {
	rb := NewRunbook(3)
	for i := 0; i < 5; i++ {
		rb.Record("svc", fmt.Sprintf("event-%d", i), "note")
	}
	all := rb.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	if all[0].Event != "event-2" {
		t.Errorf("expected oldest surviving entry to be event-2, got %s", all[0].Event)
	}
}

func TestRunbook_ForProcess(t *testing.T) {
	rb := NewRunbook(20)
	rb.Record("alpha", "crash", "segfault")
	rb.Record("beta", "restart", "policy")
	rb.Record("alpha", "recover", "restarted ok")

	entries := rb.ForProcess("alpha")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for alpha, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Process != "alpha" {
			t.Errorf("unexpected process in result: %s", e.Process)
		}
	}
}

func TestRunbook_Len(t *testing.T) {
	rb := NewRunbook(10)
	if rb.Len() != 0 {
		t.Fatalf("expected 0, got %d", rb.Len())
	}
	rb.Record("svc", "e", "n")
	if rb.Len() != 1 {
		t.Fatalf("expected 1, got %d", rb.Len())
	}
}

func TestRunbook_Clear(t *testing.T) {
	rb := NewRunbook(10)
	rb.Record("svc", "crash", "note")
	rb.Record("svc", "restart", "note")
	rb.Clear()
	if rb.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", rb.Len())
	}
}

func TestRunbook_AllReturnsCopy(t *testing.T) {
	rb := NewRunbook(10)
	rb.Record("svc", "crash", "note")
	all := rb.All()
	all[0].Note = "mutated"

	original := rb.All()
	if original[0].Note == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
