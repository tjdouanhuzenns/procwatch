package supervisor

import (
	"fmt"
	"sync"
	"testing"
)

func TestRunbook_ConcurrentWrites(t *testing.T) {
	rb := NewRunbook(500)
	const goroutines = 20
	const perGoroutine = 25

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				rb.Record(fmt.Sprintf("proc-%d", id), "event", "concurrent note")
			}
		}(i)
	}
	wg.Wait()

	if rb.Len() != goroutines*perGoroutine {
		t.Errorf("expected %d entries, got %d", goroutines*perGoroutine, rb.Len())
	}
}

func TestRunbook_ConcurrentReadWrite(t *testing.T) {
	rb := NewRunbook(100)
	const writers = 5
	const readers = 5
	const ops = 20

	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				rb.Record(fmt.Sprintf("proc-%d", id), "write", "note")
			}
		}(i)
	}

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				_ = rb.All()
				_ = rb.ForProcess(fmt.Sprintf("proc-%d", id))
				_ = rb.Len()
			}
		}(i)
	}

	wg.Wait()
}
