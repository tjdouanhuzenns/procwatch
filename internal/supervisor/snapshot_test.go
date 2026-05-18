package supervisor

import (
	"testing"
	"time"
)

func TestSnapshotStore_SaveAndGet(t *testing.T) {
	store := NewSnapshotStore()

	snap := ProcessSnapshot{Name: "web", State: "running", Restarts: 0}
	store.Save(snap)

	got, ok := store.Get("web")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if got.State != "running" {
		t.Errorf("expected state 'running', got %q", got.State)
	}
}

func TestSnapshotStore_SaveReturnsTrueOnChange(t *testing.T) {
	store := NewSnapshotStore()

	snap := ProcessSnapshot{Name: "web", State: "running", Restarts: 0}
	if !store.Save(snap) {
		t.Error("first save should report changed")
	}

	// same state and restarts — no change
	if store.Save(snap) {
		t.Error("identical save should not report changed")
	}

	// state changes
	snap.State = "stopped"
	if !store.Save(snap) {
		t.Error("state change should report changed")
	}
}

func TestSnapshotStore_SaveReturnsTrueOnRestartIncrement(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(ProcessSnapshot{Name: "svc", State: "running", Restarts: 1})

	if !store.Save(ProcessSnapshot{Name: "svc", State: "running", Restarts: 2}) {
		t.Error("restart increment should report changed")
	}
}

func TestSnapshotStore_TimestampIsSet(t *testing.T) {
	store := NewSnapshotStore()
	before := time.Now()
	store.Save(ProcessSnapshot{Name: "svc", State: "running"})
	after := time.Now()

	snap, _ := store.Get("svc")
	if snap.Timestamp.Before(before) || snap.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", snap.Timestamp, before, after)
	}
}

func TestSnapshotStore_GetMissing(t *testing.T) {
	store := NewSnapshotStore()
	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("expected missing snapshot to return false")
	}
}

func TestSnapshotStore_All(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(ProcessSnapshot{Name: "a", State: "running"})
	store.Save(ProcessSnapshot{Name: "b", State: "stopped"})

	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(all))
	}
}

func TestSnapshotStore_Remove(t *testing.T) {
	store := NewSnapshotStore()
	store.Save(ProcessSnapshot{Name: "svc", State: "running"})
	store.Remove("svc")

	_, ok := store.Get("svc")
	if ok {
		t.Error("expected snapshot to be removed")
	}
}
