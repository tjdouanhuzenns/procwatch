package supervisor

import (
	"testing"
	"time"
)

func TestProcessTimeout_SetAndGet(t *testing.T) {
	pt := NewProcessTimeout()
	if err := pt.Set("web", 5*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := pt.Get("web")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if d != 5*time.Second {
		t.Errorf("expected 5s, got %v", d)
	}
}

func TestProcessTimeout_GetMissing(t *testing.T) {
	pt := NewProcessTimeout()
	_, ok := pt.Get("missing")
	if ok {
		t.Fatal("expected missing entry to return false")
	}
}

func TestProcessTimeout_SetEmptyNameErrors(t *testing.T) {
	pt := NewProcessTimeout()
	if err := pt.Set("", time.Second); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessTimeout_SetNonPositiveDurationErrors(t *testing.T) {
	pt := NewProcessTimeout()
	if err := pt.Set("web", 0); err == nil {
		t.Fatal("expected error for zero duration")
	}
	if err := pt.Set("web", -time.Second); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestProcessTimeout_Remove(t *testing.T) {
	pt := NewProcessTimeout()
	_ = pt.Set("web", time.Second)
	pt.Remove("web")
	_, ok := pt.Get("web")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessTimeout_All(t *testing.T) {
	pt := NewProcessTimeout()
	_ = pt.Set("web", 2*time.Second)
	_ = pt.Set("worker", 10*time.Second)
	all := pt.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["web"] != 2*time.Second {
		t.Errorf("unexpected value for web: %v", all["web"])
	}
}

func TestProcessTimeout_Exceeded(t *testing.T) {
	pt := NewProcessTimeout()
	_ = pt.Set("web", 5*time.Second)

	if pt.Exceeded("web", 3*time.Second) {
		t.Error("expected not exceeded at 3s")
	}
	if !pt.Exceeded("web", 5*time.Second) {
		t.Error("expected exceeded at exactly 5s")
	}
	if !pt.Exceeded("web", 7*time.Second) {
		t.Error("expected exceeded at 7s")
	}
}

func TestProcessTimeout_ExceededMissingProcess(t *testing.T) {
	pt := NewProcessTimeout()
	if pt.Exceeded("ghost", 100*time.Second) {
		t.Error("expected false for unconfigured process")
	}
}

func TestProcessTimeout_OverwriteUpdates(t *testing.T) {
	pt := NewProcessTimeout()
	_ = pt.Set("web", 2*time.Second)
	_ = pt.Set("web", 8*time.Second)
	d, _ := pt.Get("web")
	if d != 8*time.Second {
		t.Errorf("expected 8s after overwrite, got %v", d)
	}
}
