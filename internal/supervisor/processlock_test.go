package supervisor

import (
	"sync"
	"testing"
)

func TestProcessLock_UnlockedByDefault(t *testing.T) {
	pl := NewProcessLock()
	if pl.IsLocked("web") {
		t.Fatal("expected process to be unlocked initially")
	}
	if h := pl.Holder("web"); h != "" {
		t.Fatalf("expected empty holder, got %q", h)
	}
}

func TestProcessLock_TryLockSucceeds(t *testing.T) {
	pl := NewProcessLock()
	if err := pl.TryLock("web", "restart"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pl.IsLocked("web") {
		t.Fatal("expected process to be locked")
	}
	if h := pl.Holder("web"); h != "restart" {
		t.Fatalf("expected holder %q, got %q", "restart", h)
	}
}

func TestProcessLock_TryLockFailsWhenHeld(t *testing.T) {
	pl := NewProcessLock()
	_ = pl.TryLock("web", "restart")
	err := pl.TryLock("web", "stop")
	if err == nil {
		t.Fatal("expected error when lock already held")
	}
}

func TestProcessLock_UnlockReleasesLock(t *testing.T) {
	pl := NewProcessLock()
	_ = pl.TryLock("web", "restart")
	if err := pl.Unlock("web"); err != nil {
		t.Fatalf("unexpected error on unlock: %v", err)
	}
	if pl.IsLocked("web") {
		t.Fatal("expected process to be unlocked after Unlock")
	}
}

func TestProcessLock_UnlockNotHeldErrors(t *testing.T) {
	pl := NewProcessLock()
	if err := pl.Unlock("web"); err == nil {
		t.Fatal("expected error when unlocking a process that is not locked")
	}
}

func TestProcessLock_EmptyProcessErrors(t *testing.T) {
	pl := NewProcessLock()
	if err := pl.TryLock("", "restart"); err == nil {
		t.Fatal("expected error for empty process name")
	}
	if err := pl.Unlock(""); err == nil {
		t.Fatal("expected error for empty process name on unlock")
	}
}

func TestProcessLock_EmptyOperationErrors(t *testing.T) {
	pl := NewProcessLock()
	if err := pl.TryLock("web", ""); err == nil {
		t.Fatal("expected error for empty operation")
	}
}

func TestProcessLock_IndependentPerProcess(t *testing.T) {
	pl := NewProcessLock()
	_ = pl.TryLock("web", "restart")
	if err := pl.TryLock("worker", "restart"); err != nil {
		t.Fatalf("expected independent lock for different process, got: %v", err)
	}
}

func TestProcessLock_ConcurrentAccess(t *testing.T) {
	pl := NewProcessLock()
	var wg sync.WaitGroup
	successes := make(chan struct{}, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := pl.TryLock("web", "op"); err == nil {
				successes <- struct{}{}
				_ = pl.Unlock("web")
			}
		}()
	}
	wg.Wait()
	close(successes)
	count := 0
	for range successes {
		count++
	}
	if count == 0 {
		t.Fatal("expected at least one goroutine to acquire the lock")
	}
}
