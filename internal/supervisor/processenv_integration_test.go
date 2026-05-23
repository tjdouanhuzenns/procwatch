package supervisor

import (
	"fmt"
	"sync"
	"testing"
)

func TestProcessEnv_ConcurrentWrites(t *testing.T) {
	s := NewProcessEnvStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("proc-%d", i)
			_ = s.Set(name, map[string]string{"IDX": fmt.Sprintf("%d", i)})
		}(i)
	}
	wg.Wait()
	all := s.All()
	if len(all) != 50 {
		t.Errorf("expected 50 entries, got %d", len(all))
	}
}

func TestProcessEnv_ConcurrentReadWrite(t *testing.T) {
	s := NewProcessEnvStore()
	_ = s.Set("shared", map[string]string{"K": "initial"})
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			_ = s.Set("shared", map[string]string{"K": fmt.Sprintf("%d", i)})
		}(i)
		go func() {
			defer wg.Done()
			_, _ = s.Get("shared")
		}()
	}
	wg.Wait()
}
