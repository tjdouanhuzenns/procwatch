package supervisor

import (
	"context"
	"sync"

	"github.com/user/procwatch/internal/config"
	"github.com/user/procwatch/internal/logger"
	"github.com/user/procwatch/internal/process"
)

// Supervisor manages a set of supervised processes.
type Supervisor struct {
	cfg     *config.Config
	log     *logger.Logger
	procs   []*process.Process
	mu      sync.Mutex
}

// New creates a new Supervisor from the given config and logger.
func New(cfg *config.Config, log *logger.Logger) *Supervisor {
	return &Supervisor{
		cfg: cfg,
		log: log,
	}
}

// Start launches all configured processes and blocks until ctx is cancelled.
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	for _, pc := range s.cfg.Processes {
		p := process.New(pc, s.log)
		s.procs = append(s.procs, p)
	}
	procs := s.procs
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, p := range procs {
		wg.Add(1)
		go func(p *process.Process) {
			defer wg.Done()
			p.Run(ctx)
		}(p)
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

// Processes returns a snapshot of the supervised processes.
func (s *Supervisor) Processes() []*process.Process {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*process.Process, len(s.procs))
	copy(out, s.procs)
	return out
}
