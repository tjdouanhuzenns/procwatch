package supervisor

import (
	"sync"
	"testing"
)

func TestProcessWeight_ConcurrentWrites(t *testing.T) {
	pw := NewProcessWeight()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = pw.Set("worker", n)
		}(i)
	}
	wg.Wait()
	_, ok := pw.Get("worker")
	if !ok {
		t.Fatal("expected weight to be present after concurrent writes")
	}
}

func TestProcessWeight_ConcurrentReadWrite(t *testing.T) {
	pw := NewProcessWeight()
	_ = pw.Set("svc", 5)
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			_ = pw.Set("svc", n)
		}(i)
		go func() {
			defer wg.Done()
			pw.Get("svc")
		}()
	}
	wg.Wait()
}
