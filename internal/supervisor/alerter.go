package supervisor

import (
	"context"
	"fmt"
	"io"
)

// AlertRule defines which state transitions should trigger an alert.
type AlertRule struct {
	// TriggerStates lists the NewStatus values that should fire an alert.
	TriggerStates []string
}

// Alerter listens on a Watcher's event channel and writes structured alerts.
type Alerter struct {
	watcher *Watcher
	rule    AlertRule
	out     io.Writer
}

// NewAlerter creates an Alerter that writes to out when a WatchEvent matches rule.
func NewAlerter(watcher *Watcher, rule AlertRule, out io.Writer) *Alerter {
	return &Alerter{watcher: watcher, rule: rule, out: out}
}

// Run consumes events until ctx is cancelled or the events channel closes.
func (a *Alerter) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-a.watcher.Events():
			if !ok {
				return
			}
			if a.matches(ev) {
				a.emit(ev)
			}
		}
	}
}

func (a *Alerter) matches(ev WatchEvent) bool {
	for _, s := range a.rule.TriggerStates {
		if s == ev.NewStatus {
			return true
		}
	}
	return false
}

func (a *Alerter) emit(ev WatchEvent) error {
	_, err := fmt.Fprintf(
		a.out,
		`{"alert":true,"process":%q,"old_status":%q,"new_status":%q,"at":%q}`+"\n",
		ev.ProcessName, ev.OldStatus, ev.NewStatus, ev.At.UTC().Format("2006-01-02T15:04:05Z"),
	)
	return err
}
