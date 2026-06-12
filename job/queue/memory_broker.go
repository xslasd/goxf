package queue

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type delayedTask struct {
	queueName string
	msg       *Message
	executeAt time.Time
}

type taskHeap []delayedTask

func (h taskHeap) Len() int           { return len(h) }
func (h taskHeap) Less(i, j int) bool { return h[i].executeAt.Before(h[j].executeAt) }
func (h taskHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *taskHeap) Push(x any) {
	*h = append(*h, x.(delayedTask))
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MemoryBroker is an in-memory implementation of the Broker interface based on Go channels.
// It uses a min-heap and a single sweeper goroutine to efficiently manage delayed tasks.
type MemoryBroker struct {
	queues map[string]chan any
	mu     sync.RWMutex

	dHeap  taskHeap
	dMu    sync.Mutex
	wakeup chan struct{}
	ctx    context.Context
	cancel context.CancelFunc

	deletedTasks map[string]struct{}
	delMu        sync.Mutex
}

// NewMemoryBroker creates a new in-memory broker.
func NewMemoryBroker() *MemoryBroker {
	ctx, cancel := context.WithCancel(context.Background())
	m := &MemoryBroker{
		queues:       make(map[string]chan any),
		dHeap:        make(taskHeap, 0),
		wakeup:       make(chan struct{}, 1),
		ctx:          ctx,
		cancel:       cancel,
		deletedTasks: make(map[string]struct{}),
	}
	heap.Init(&m.dHeap)
	go m.runSweeper()
	return m
}

// getOrCreateQueue safely gets or creates the underlying channel for a queue.
func (m *MemoryBroker) getOrCreateQueue(queueName string) chan any {
	m.mu.RLock()
	ch, exists := m.queues[queueName]
	m.mu.RUnlock()

	if exists {
		return ch
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if ch, exists = m.queues[queueName]; exists {
		return ch
	}
	
	ch = make(chan any, 1000)
	m.queues[queueName] = ch
	return ch
}

// Dequeue receives a message on the specified queue or blocks until one is available.
func (m *MemoryBroker) Dequeue(ctx context.Context, queueName string) (*Message, error) {
	ch := m.getOrCreateQueue(queueName)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msgVal := <-ch:
			msg, ok := msgVal.(*Message)
			if !ok {
				continue
			}
			m.delMu.Lock()
			_, deleted := m.deletedTasks[msg.ID]
			if deleted {
				delete(m.deletedTasks, msg.ID)
				m.delMu.Unlock()
				continue
			}
			m.delMu.Unlock()
			return msg, nil
		}
	}
}

// Enqueue publishes a message to the specified queue immediately.
func (m *MemoryBroker) Enqueue(ctx context.Context, queueName string, msg *Message) error {
	ch := m.getOrCreateQueue(queueName)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- msg:
		return nil
	}
}

// EnqueueAfter schedules a message for delivery after the specified delay.
func (m *MemoryBroker) EnqueueAfter(ctx context.Context, queueName string, msg *Message, delay time.Duration) error {
	task := delayedTask{
		queueName: queueName,
		msg:       msg,
		executeAt: time.Now().Add(delay),
	}

	m.dMu.Lock()
	heap.Push(&m.dHeap, task)
	// Check if the newly added task is the earliest
	isEarliest := m.dHeap[0].executeAt.Equal(task.executeAt)
	m.dMu.Unlock()

	if isEarliest {
		select {
		case m.wakeup <- struct{}{}:
		default:
		}
	}
	return nil
}

func (m *MemoryBroker) runSweeper() {
	timer := time.NewTimer(0)
	<-timer.C // Stop initially

	for {
		m.dMu.Lock()
		var delay time.Duration
		hasTasks := m.dHeap.Len() > 0
		if hasTasks {
			now := time.Now()
			top := m.dHeap[0]
			if top.executeAt.After(now) {
				delay = top.executeAt.Sub(now)
			} else {
				// Due task, pop and enqueue
				task := heap.Pop(&m.dHeap).(delayedTask)
				m.dMu.Unlock()

				m.delMu.Lock()
				_, deleted := m.deletedTasks[task.msg.ID]
				if deleted {
					delete(m.deletedTasks, task.msg.ID)
					m.delMu.Unlock()
					continue
				}
				m.delMu.Unlock()

				_ = m.Enqueue(context.Background(), task.queueName, task.msg)
				continue
			}
		}
		m.dMu.Unlock()

		if hasTasks {
			timer.Reset(delay)
			select {
			case <-m.ctx.Done():
				timer.Stop()
				return
			case <-m.wakeup:
				timer.Stop()
			case <-timer.C:
			}
		} else {
			select {
			case <-m.ctx.Done():
				return
			case <-m.wakeup:
			}
		}
	}
}

// Delete cancels a delayed task or a task waiting in the queue.
func (m *MemoryBroker) Delete(ctx context.Context, queueName string, id string) error {
	m.delMu.Lock()
	m.deletedTasks[id] = struct{}{}
	m.delMu.Unlock()
	return nil
}

// Close stops the background sweeper for delayed tasks.
func (m *MemoryBroker) Close() {
	m.cancel()
}
