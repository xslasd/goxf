package queue

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
)

func NewWorker[T any](broker Broker, handler Handler[T], opts ...Option) (*Worker[T], error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		if errors.Is(err, conf.ErrInvalidKey) {
			log.Warn("job queue use default config!")
		} else {
			return nil, err
		}
	}

	if opt.config.QueueName == "" {
		opt.config.QueueName = opt.confName
	}

	w := &Worker[T]{
		broker:      broker,
		handler:     handler,
		opts:        opt,
		activeTasks: make(map[string]context.CancelFunc),
	}

	if handler != nil {
		if err := w.run(); err != nil {
			return nil, err
		}
	}

	hooks.Register(hooks.Stage_AfterStop, func() {
		w.stop()
	})

	return w, nil
}

// Enqueue pushes a message to the queue utilizing the underlying Broker.
func (w *Worker[T]) Enqueue(ctx context.Context, payload T) (string, error) {
	taskID := uuid.New().String()
	err := w.EnqueueWithID(ctx, taskID, payload)
	return taskID, err
}

// EnqueueWithID pushes a message to the queue utilizing the underlying Broker with a custom task ID.
func (w *Worker[T]) EnqueueWithID(ctx context.Context, taskID string, payload T) error {
	name := w.opts.config.QueueName
	beg := time.Now()
	msg := &Message{
		ID:      taskID,
		Payload: payload,
	}
	err := w.broker.Enqueue(ctx, name, msg)
	code := "ok"
	if err != nil {
		code = "error"
	}

	if w.opts.enableMetric {
		metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.JobType, name, "Enqueue", "")
		metric.ClientHandleCounter.Inc(metric.JobType, name, "Enqueue", "", code)
	}
	return err
}

// EnqueueAfter schedules a message for future delivery.
func (w *Worker[T]) EnqueueAfter(ctx context.Context, payload T, delay time.Duration) (string, error) {
	taskID := uuid.New().String()
	err := w.EnqueueAfterWithID(ctx, taskID, payload, delay)
	return taskID, err
}

// EnqueueAfterWithID schedules a message for future delivery with a custom task ID.
func (w *Worker[T]) EnqueueAfterWithID(ctx context.Context, taskID string, payload T, delay time.Duration) error {
	name := w.opts.config.QueueName
	beg := time.Now()
	msg := &Message{
		ID:      taskID,
		Payload: payload,
	}
	err := w.broker.EnqueueAfter(ctx, name, msg, delay)
	code := "ok"
	if err != nil {
		code = "error"
	}

	if w.opts.enableMetric {
		metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.JobType, name, "EnqueueAfter", "")
		metric.ClientHandleCounter.Inc(metric.JobType, name, "EnqueueAfter", "", code)
	}
	return err
}

// Cancel cancels a task, either if it is still delayed or if it is currently executing.
func (w *Worker[T]) Cancel(ctx context.Context, taskID string) error {
	// 1. Try to cancel currently executing task
	w.activeTasksMu.Lock()
	cancel, ok := w.activeTasks[taskID]
	w.activeTasksMu.Unlock()
	if ok {
		cancel()
	}

	// 2. Try to cancel/delete it from the underlying broker (delayed or queued)
	queueName := w.opts.config.QueueName
	return w.broker.Delete(ctx, queueName, taskID)
}
