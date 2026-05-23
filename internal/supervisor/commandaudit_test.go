package supervisor

import (
	"testing"
)

func TestCommandAudit_RecordAndAll(t *testing.T) {
	a := NewCommandAudit(10)
	if err := a.Record("web", "start", "admin"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := a.Record("db", "stop", "ops"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := a.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Process != "web" || entries[0].Command != "start" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
}

func TestCommandAudit_TimestampIsSet(t *testing.T) {
	a := NewCommandAudit(10)
	_ = a.Record("svc", "restart", "ci")
	entries := a.All()
	if entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestCommandAudit_EvictsOldestWhenFull(t *testing.T) {
	a := NewCommandAudit(3)
	_ = a.Record("a", "start", "u")
	_ = a.Record("b", "start", "u")
	_ = a.Record("c", "start", "u")
	_ = a.Record("d", "start", "u")
	entries := a.All()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Process != "b" {
		t.Errorf("expected oldest evicted, got %s", entries[0].Process)
	}
}

func TestCommandAudit_ForProcess(t *testing.T) {
	a := NewCommandAudit(20)
	_ = a.Record("web", "start", "u")
	_ = a.Record("db", "stop", "u")
	_ = a.Record("web", "restart", "u")
	results := a.ForProcess("web")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for web, got %d", len(results))
	}
	for _, e := range results {
		if e.Process != "web" {
			t.Errorf("unexpected process in result: %s", e.Process)
		}
	}
}

func TestCommandAudit_EmptyProcessErrors(t *testing.T) {
	a := NewCommandAudit(10)
	if err := a.Record("", "start", "u"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestCommandAudit_EmptyCommandErrors(t *testing.T) {
	a := NewCommandAudit(10)
	if err := a.Record("web", "", "u"); err == nil {
		t.Error("expected error for empty command")
	}
}

func TestCommandAudit_Len(t *testing.T) {
	a := NewCommandAudit(10)
	if a.Len() != 0 {
		t.Fatalf("expected 0, got %d", a.Len())
	}
	_ = a.Record("svc", "start", "u")
	if a.Len() != 1 {
		t.Fatalf("expected 1, got %d", a.Len())
	}
}

func TestCommandAudit_DefaultMaxSize(t *testing.T) {
	a := NewCommandAudit(0)
	if a.maxSize != 256 {
		t.Errorf("expected default maxSize 256, got %d", a.maxSize)
	}
}
