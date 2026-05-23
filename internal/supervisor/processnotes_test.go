package supervisor

import (
	"testing"
)

func TestProcessNotes_AddAndForProcess(t *testing.T) {
	pn := NewProcessNotes(10)
	if err := pn.Add("web", "deployed v1.2", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	notes := pn.ForProcess("web")
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Note != "deployed v1.2" {
		t.Errorf("unexpected note text: %s", notes[0].Note)
	}
	if notes[0].Author != "alice" {
		t.Errorf("unexpected author: %s", notes[0].Author)
	}
}

func TestProcessNotes_TimestampIsSet(t *testing.T) {
	pn := NewProcessNotes(10)
	_ = pn.Add("web", "check timestamp", "bob")
	notes := pn.ForProcess("web")
	if notes[0].CreatedAt.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestProcessNotes_EvictsOldestWhenFull(t *testing.T) {
	pn := NewProcessNotes(3)
	for i := 0; i < 5; i++ {
		_ = pn.Add("svc", fmt.Sprintf("note %d", i), "sys")
	}
	notes := pn.ForProcess("svc")
	if len(notes) != 3 {
		t.Fatalf("expected 3 notes after eviction, got %d", len(notes))
	}
	if notes[0].Note != "note 2" {
		t.Errorf("expected oldest surviving note to be 'note 2', got %q", notes[0].Note)
	}
}

func TestProcessNotes_EmptyProcessErrors(t *testing.T) {
	pn := NewProcessNotes(10)
	if err := pn.Add("", "hello", "x"); err == nil {
		t.Error("expected error for empty process name")
	}
}

func TestProcessNotes_EmptyNoteErrors(t *testing.T) {
	pn := NewProcessNotes(10)
	if err := pn.Add("web", "", "x"); err == nil {
		t.Error("expected error for empty note")
	}
}

func TestProcessNotes_All(t *testing.T) {
	pn := NewProcessNotes(10)
	_ = pn.Add("web", "note a", "alice")
	_ = pn.Add("worker", "note b", "bob")
	all := pn.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(all))
	}
}

func TestProcessNotes_Clear(t *testing.T) {
	pn := NewProcessNotes(10)
	_ = pn.Add("web", "to be cleared", "alice")
	pn.Clear("web")
	if notes := pn.ForProcess("web"); len(notes) != 0 {
		t.Errorf("expected 0 notes after clear, got %d", len(notes))
	}
}

func TestProcessNotes_DefaultMaxPerProcess(t *testing.T) {
	pn := NewProcessNotes(0) // should default to 50
	for i := 0; i < 60; i++ {
		_ = pn.Add("svc", fmt.Sprintf("note %d", i), "sys")
	}
	if got := len(pn.ForProcess("svc")); got != 50 {
		t.Errorf("expected 50 notes with default cap, got %d", got)
	}
}
