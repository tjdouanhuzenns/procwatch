package supervisor

import (
	"sync"
	"testing"
)

func TestHealthReporter_ConcurrentWrites(t *testing.T) {
	r := NewHealthReporter()
	var wg sync.WaitGroup

	processes := []string{"alpha", "beta", "gamma"}
	statuses := []HealthStatus{HealthHealthy, HealthDegraded, HealthUnhealthy}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := processes[i%len(processes)]
			status := statuses[i%len(statuses)]
			r.Record(name, status, "concurrent")
		}(i)
	}

	wg.Wait()

	all := r.All()
	if len(all) != len(processes) {
		t.Errorf("expected %d reports, got %d", len(processes), len(all))
	}
}

func TestHealthReporter_ConcurrentReadWrite(t *testing.T) {
	r := NewHealthReporter()
	r.Record("svc", HealthHealthy, "initial")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			r.Record("svc", HealthDegraded, "write")
		}()
		go func() {
			defer wg.Done()
			r.Get("svc")
		}()
	}
	wg.Wait()

	rep, ok := r.Get("svc")
	if !ok {
		t.Fatal("expected report to still exist after concurrent access")
	}
	if rep.Process != "svc" {
		t.Errorf("unexpected process name: %q", rep.Process)
	}
}
