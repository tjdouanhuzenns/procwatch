package supervisor

import (
	"testing"
)

func TestCheckpoint_SaveAndGet(t *testing.T) {
	cs := NewCheckpointStore()
	err := cs.Save("svc", "boot", map[string]string{"step": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp, ok := cs.Get("svc", "boot")
	if !ok {
		t.Fatal("expected checkpoint to exist")
	}
	if cp.Metadata["step"] != "1" {
		t.Errorf("expected step=1, got %s", cp.Metadata["step"])
	}
}

func TestCheckpoint_GetMissing(t *testing.T) {
	cs := NewCheckpointStore()
	_, ok := cs.Get("svc", "missing")
	if ok {
		t.Fatal("expected missing checkpoint to return false")
	}
}

func TestCheckpoint_EmptyProcessErrors(t *testing.T) {
	cs := NewCheckpointStore()
	if err := cs.Save("", "boot", nil); err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestCheckpoint_EmptyNameErrors(t *testing.T) {
	cs := NewCheckpointStore()
	if err := cs.Save("svc", "", nil); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCheckpoint_OverwriteUpdates(t *testing.T) {
	cs := NewCheckpointStore()
	_ = cs.Save("svc", "boot", map[string]string{"step": "1"})
	_ = cs.Save("svc", "boot", map[string]string{"step": "2"})
	cp, _ := cs.Get("svc", "boot")
	if cp.Metadata["step"] != "2" {
		t.Errorf("expected overwritten step=2, got %s", cp.Metadata["step"])
	}
}

func TestCheckpoint_ForProcess(t *testing.T) {
	cs := NewCheckpointStore()
	_ = cs.Save("svc", "boot", nil)
	_ = cs.Save("svc", "ready", nil)
	_ = cs.Save("other", "init", nil)
	results := cs.ForProcess("svc")
	if len(results) != 2 {
		t.Errorf("expected 2 checkpoints, got %d", len(results))
	}
}

func TestCheckpoint_Delete(t *testing.T) {
	cs := NewCheckpointStore()
	_ = cs.Save("svc", "boot", nil)
	cs.Delete("svc", "boot")
	_, ok := cs.Get("svc", "boot")
	if ok {
		t.Fatal("expected checkpoint to be deleted")
	}
}

func TestCheckpoint_Clear(t *testing.T) {
	cs := NewCheckpointStore()
	_ = cs.Save("svc", "boot", nil)
	_ = cs.Save("svc", "ready", nil)
	cs.Clear("svc")
	if got := cs.ForProcess("svc"); len(got) != 0 {
		t.Errorf("expected 0 after clear, got %d", len(got))
	}
}

func TestCheckpoint_TimestampIsSet(t *testing.T) {
	cs := NewCheckpointStore()
	_ = cs.Save("svc", "boot", nil)
	cp, _ := cs.Get("svc", "boot")
	if cp.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}
