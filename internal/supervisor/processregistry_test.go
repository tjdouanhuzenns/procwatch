package supervisor

import (
	"sync"
	"testing"
)

func TestProcessRegistry_RegisterAndGet(t *testing.T) {
	r := NewProcessRegistry()

	if err := r.Register("web", "/usr/bin/web", map[string]string{"env": "prod"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e, ok := r.Get("web")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Name != "web" || e.Command != "/usr/bin/web" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.Tags["env"] != "prod" {
		t.Errorf("expected tag env=prod, got %q", e.Tags["env"])
	}
}

func TestProcessRegistry_GetMissing(t *testing.T) {
	r := NewProcessRegistry()
	_, ok := r.Get("missing")
	if ok {
		t.Fatal("expected no entry for unknown process")
	}
}

func TestProcessRegistry_RegisterEmptyNameErrors(t *testing.T) {
	r := NewProcessRegistry()
	if err := r.Register("", "/bin/foo", nil); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessRegistry_RegisterEmptyCommandErrors(t *testing.T) {
	r := NewProcessRegistry()
	if err := r.Register("svc", "", nil); err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestProcessRegistry_Deregister(t *testing.T) {
	r := NewProcessRegistry()
	_ = r.Register("worker", "/bin/worker", nil)
	r.Deregister("worker")
	if _, ok := r.Get("worker"); ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestProcessRegistry_DeregisterUnknownIsNoop(t *testing.T) {
	r := NewProcessRegistry()
	r.Deregister("ghost") // must not panic
}

func TestProcessRegistry_All(t *testing.T) {
	r := NewProcessRegistry()
	_ = r.Register("a", "/bin/a", nil)
	_ = r.Register("b", "/bin/b", nil)

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestProcessRegistry_TagsAreCopied(t *testing.T) {
	r := NewProcessRegistry()
	tags := map[string]string{"k": "v"}
	_ = r.Register("svc", "/bin/svc", tags)
	tags["k"] = "mutated"

	e, _ := r.Get("svc")
	if e.Tags["k"] != "v" {
		t.Errorf("tags were not copied; got %q", e.Tags["k"])
	}
}

func TestProcessRegistry_ConcurrentAccess(t *testing.T) {
	r := NewProcessRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("proc-%d", i)
			_ = r.Register(name, "/bin/x", nil)
			_, _ = r.Get(name)
			_ = r.All()
		}(i)
	}
	wg.Wait()
}
