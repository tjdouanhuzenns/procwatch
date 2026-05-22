package supervisor

import (
	"sync"
	"testing"
)

// TestDependencyGraph_ConcurrentAccess verifies that concurrent reads and
// writes do not race (run with -race).
func TestDependencyGraph_ConcurrentAccess(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("a", []string{"b"})
	_ = g.Add("b", []string{})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			running := map[string]bool{"b": true}
			_ = g.ReadyToStart("a", running)
			_ = g.Deps("a")
		}(i)
	}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Adding a new independent node should not race.
			name := "node"
			_ = g.Add(name, []string{})
		}(i)
	}
	wg.Wait()
}

// TestDependencyGraph_ChainReady verifies a multi-hop dependency chain.
func TestDependencyGraph_ChainReady(t *testing.T) {
	g := NewDependencyGraph()
	_ = g.Add("c", []string{"b"})
	_ = g.Add("b", []string{"a"})
	_ = g.Add("a", []string{})

	running := map[string]bool{}
	if g.ReadyToStart("b", running) {
		t.Fatal("b should not be ready when a is not running")
	}
	running["a"] = true
	if !g.ReadyToStart("b", running) {
		t.Fatal("b should be ready when a is running")
	}
	if g.ReadyToStart("c", running) {
		t.Fatal("c should not be ready when b is not running")
	}
	running["b"] = true
	if !g.ReadyToStart("c", running) {
		t.Fatal("c should be ready when both a and b are running")
	}
}
