package supervisor

import (
	"testing"
)

func TestPauseResume_DefaultIsActive(t *testing.T) {
	c := NewPauseResumeController()
	if c.IsPaused("svc") {
		t.Fatal("expected process to be active by default")
	}
	if got := c.State("svc"); got != PauseStateActive {
		t.Fatalf("expected active, got %s", got)
	}
}

func TestPauseResume_PauseAndResume(t *testing.T) {
	c := NewPauseResumeController()

	if err := c.Pause("svc"); err != nil {
		t.Fatalf("unexpected error pausing: %v", err)
	}
	if !c.IsPaused("svc") {
		t.Fatal("expected process to be paused")
	}

	if err := c.Resume("svc"); err != nil {
		t.Fatalf("unexpected error resuming: %v", err)
	}
	if c.IsPaused("svc") {
		t.Fatal("expected process to be active after resume")
	}
}

func TestPauseResume_DoublePauseErrors(t *testing.T) {
	c := NewPauseResumeController()
	_ = c.Pause("svc")
	if err := c.Pause("svc"); err == nil {
		t.Fatal("expected error when pausing already-paused process")
	}
}

func TestPauseResume_ResumeNotPausedErrors(t *testing.T) {
	c := NewPauseResumeController()
	if err := c.Resume("svc"); err == nil {
		t.Fatal("expected error when resuming non-paused process")
	}
}

func TestPauseResume_All(t *testing.T) {
	c := NewPauseResumeController()
	_ = c.Pause("a")
	_ = c.Pause("b")

	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a"] != PauseStatePaused || all["b"] != PauseStatePaused {
		t.Fatal("expected both processes to be paused")
	}
}

func TestPauseResume_AllIsCopy(t *testing.T) {
	c := NewPauseResumeController()
	_ = c.Pause("svc")

	all := c.All()
	all["svc"] = PauseStateActive

	if !c.IsPaused("svc") {
		t.Fatal("modifying All() result should not affect controller state")
	}
}

func TestPauseResume_Remove(t *testing.T) {
	c := NewPauseResumeController()
	_ = c.Pause("svc")
	c.Remove("svc")

	if c.IsPaused("svc") {
		t.Fatal("expected process to be active after removal")
	}
	if len(c.All()) != 0 {
		t.Fatal("expected empty state after removal")
	}
}

func TestPauseState_String(t *testing.T) {
	if PauseStateActive.String() != "active" {
		t.Errorf("expected 'active', got %s", PauseStateActive.String())
	}
	if PauseStatePaused.String() != "paused" {
		t.Errorf("expected 'paused', got %s", PauseStatePaused.String())
	}
}
