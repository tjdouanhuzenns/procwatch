package supervisor

import (
	"context"
	"sync"
	"time"
)

// NotifyEvent represents a notification triggered by a process state change.
type NotifyEvent struct {
	Process   string
	State     string
	Message   string
	Timestamp time.Time
}

// NotifyHandler is a function that receives a notification event.
type NotifyHandler func(ctx context.Context, event NotifyEvent) error

// Notifier dispatches state-change notifications to registered handlers.
type Notifier struct {
	mu       sync.RWMutex
	handlers []NotifyHandler
}

// NewNotifier creates a new Notifier with no handlers.
func NewNotifier() *Notifier {
	return &Notifier{}
}

// Register adds a handler to the notifier.
func (n *Notifier) Register(h NotifyHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers = append(n.handlers, h)
}

// Notify dispatches an event to all registered handlers.
// Errors from individual handlers are collected but do not stop dispatch.
func (n *Notifier) Notify(ctx context.Context, event NotifyEvent) []error {
	n.mu.RLock()
	handlers := make([]NotifyHandler, len(n.handlers))
	copy(handlers, n.handlers)
	n.mu.RUnlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	var errs []error
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Len returns the number of registered handlers.
func (n *Notifier) Len() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.handlers)
}
