package supervisor

import (
	"errors"
	"sync"
	"time"
)

// Priority levels for process restart scheduling.
const (
	PriorityLow    = 0
	PriorityNormal = 1
	PriorityHigh   = 2
	PriorityCritical = 3
)

// RestartRequest represents a queued restart for a managed process.
type RestartRequest struct {
	Process   string
	Priority  int
	Scheduled time.Time
	Reason    string
}

// priorityEntry is an internal heap node.
type priorityEntry struct {
	req   RestartRequest
	index int
}

// entryHeap implements heap.Interface for priority-ordered restart requests.
type entryHeap []*priorityEntry

func (h entryHeap) Len() int { return len(h) }

func (h entryHeap) Less(i, j int) bool {
	// Higher priority wins; break ties by earliest scheduled time.
	if h[i].req.Priority != h[j].req.Priority {
		return h[i].req.Priority > h[j].req.Priority
	}
	return h[i].req.Scheduled.Before(h[j].req.Scheduled)
}

func (h entryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *entryHeap) Push(x any) {
	e := x.(*priorityEntry)
	e.index = len(*h)
	*h = append(*h, e)
}

func (h *entryHeap) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	e.index = -1
	return e
}

// PriorityQueue schedules process restart requests ordered by priority and
// scheduled time. It is safe for concurrent use.
type PriorityQueue struct {
	mu   sync.Mutex
	heap entryHeap
	cap  int
}

// NewPriorityQueue creates a PriorityQueue with the given maximum capacity.
// A capacity of 0 means unlimited.
func NewPriorityQueue(capacity int) *PriorityQueue {
	return &PriorityQueue{cap: capacity}
}

// Push adds a restart request to the queue. Returns an error if the queue is
// at capacity.
func (q *PriorityQueue) Push(req RestartRequest) error {
	if req.Process == "" {
		return errors.New("priorityqueue: process name must not be empty")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.cap > 0 && len(q.heap) >= q.cap {
		return errors.New("priorityqueue: queue is full")
	}
	if req.Scheduled.IsZero() {
		req.Scheduled = time.Now()
	}
	heapPush(&q.heap, &priorityEntry{req: req})
	return nil
}

// Pop removes and returns the highest-priority request. Returns false if the
// queue is empty.
func (q *PriorityQueue) Pop() (RestartRequest, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.heap) == 0 {
		return RestartRequest{}, false
	}
	e := heapPop(&q.heap).(*priorityEntry)
	return e.req, true
}

// Len returns the number of pending restart requests.
func (q *PriorityQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.heap)
}

// Drain returns all pending requests in priority order and empties the queue.
func (q *PriorityQueue) Drain() []RestartRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]RestartRequest, 0, len(q.heap))
	for len(q.heap) > 0 {
		e := heapPop(&q.heap).(*priorityEntry)
		out = append(out, e.req)
	}
	return out
}

// --- minimal heap operations (avoids importing container/heap to keep deps clean) ---

func heapPush(h *entryHeap, e *priorityEntry) {
	h.Push(e)
	siftUp(h, len(*h)-1)
}

func heapPop(h *entryHeap) any {
	n := len(*h) - 1
	h.Swap(0, n)
	siftDown(h, 0, n)
	return h.Pop()
}

func siftUp(h *entryHeap, i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.Less(i, parent) {
			break
		}
		h.Swap(i, parent)
		i = parent
	}
}

func siftDown(h *entryHeap, i, n int) {
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		j := left
		if right := left + 1; right < n && h.Less(right, left) {
			j = right
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
}
