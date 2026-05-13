package process

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/user/procwatch/internal/logger"
)

// State represents the current state of a managed process.
type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateRestarting
	StateFailed
)

// Process manages the lifecycle of a single supervised process.
type Process struct {
	name    string
	command string
	args    []string
	env     []string

	log     *logger.Logger
	restart *RestartTracker

	mu    sync.Mutex
	state State
	cmd   *exec.Cmd

	doneCh chan struct{}
}

// New creates a new Process with the given configuration.
func New(name, command string, args, env []string, log *logger.Logger, restart *RestartTracker) *Process {
	return &Process{
		name:    name,
		command: command,
		args:    args,
		env:     env,
		log:     log,
		restart: restart,
		doneCh:  make(chan struct{}),
	}
}

// Start launches the process and supervises it according to the restart policy.
func (p *Process) Start(ctx context.Context) {
	go p.supervise(ctx)
}

// Stop signals the process to stop and waits for it to exit.
func (p *Process) Stop() {
	p.mu.Lock()
	cmd := p.cmd
	p.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	<-p.doneCh
}

// State returns the current state of the process.
func (p *Process) CurrentState() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *Process) supervise(ctx context.Context) {
	defer close(p.doneCh)

	for {
		p.setState(StateStarting)
		cmd := exec.CommandContext(ctx, p.command, p.args...)
		if len(p.env) > 0 {
			cmd.Env = p.env
		}
		cmd.Stdout = p.log.Writer(p.name)
		cmd.Stderr = p.log.Writer(p.name)

		p.mu.Lock()
		p.cmd = cmd
		p.mu.Unlock()

		if err := cmd.Start(); err != nil {
			p.log.Error(p.name, "failed to start", err)
			p.setState(StateFailed)
			return
		}

		p.setState(StateRunning)
		p.log.Info(p.name, "process started", cmd.Process.Pid)

		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		p.log.Info(p.name, "process exited", exitCode)

		if ctx.Err() != nil {
			p.setState(StateStopped)
			return
		}

		if !p.restart.ShouldRestart(exitCode) {
			p.setState(StateFailed)
			return
		}

		backoff := p.restart.NextBackoff()
		p.setState(StateRestarting)
		p.log.Info(p.name, "restarting after backoff", backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			p.setState(StateStopped)
			return
		}
	}
}

func (p *Process) setState(s State) {
	p.mu.Lock()
	p.state = s
	p.mu.Unlock()
}
