package supervisor

import (
	"testing"
)

func TestDependencyGraph_AddAndDeps(t *testing.T) {
	g := NewDependencyGraph()
	if err := g.Add("worker", []string{"db", "cache"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	deps := g.Deps("worker")
	if len(deps) != 2 || deps[0] != "db" || deps[1] != "cache" {
		t.Fatalf("unexpected deps: %v", deps)
	}
}

func TestDependencyGraph_EmptyNameErrors(t *testing.T) {
	g := NewDependencyGraph()
	if err := g.Add("", []string{"db"}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDependencyGraph_CycleDetected(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("a", []string{"b"})
	_ = g.Add("b", []string{"c"})
	if err := g.Add("c", []string{"a"}); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestDependencyGraph_NoCycleOnValidDAG(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("a", []string{"b"})
	_ = g.Add("b", []string{"c"})
	if err := g.Add("c", []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDependencyGraph_ReadyToStart_AllRunning(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("worker", []string{"db", "cache"})
	running := map[string]bool{"db": true, "cache": true}
	if !g.ReadyToStart("worker", running) {
		t.Fatal("expected ready")
	}
}

func TestDependencyGraph_ReadyToStart_MissingDep(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("worker", []string{"db", "cache"})
	running := map[string]bool{"db": true}
	if g.ReadyToStart("worker", running) {
		t.Fatal("expected not ready")
	}
}

func TestDependencyGraph_ReadyToStart_NoDeps(t *testing.T) {
	g := NewDependencyGraph()
	if !g.ReadyToStart("standalone", map[string]bool{}) {
		t.Fatal("process with no deps should always be ready")
	}
}

func TestDependencyGraph_Remove(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("worker", []string{"db"})
	g.Remove("worker")
	if deps := g.Deps("worker"); len(deps) != 0 {
		t.Fatalf("expected no deps after remove, got %v", deps)
	}
}

func TestDependencyGraph_CycleRollback(t *testing.T) {
	// Ensure original edges are preserved after a rejected cycle.
	g := NewDependencyGraph()
	_ = g.Add("a", []string{"b"})
	_ = g.Add("b", []string{}) // valid so far
	// Introduce a cycle via b -> a, should fail.
	_ = g.Add("b", []string{"a"})
	// After rejection, b should still have its previous empty deps.
	if deps := g.Deps("b"); len(deps) != 0 {
		t.Fatalf("rollback failed, got deps: %v", deps)
	}
}
