package queue

import (
	"context"
	"errors"
	"time"

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
		broker:  broker,
		handler: handler,
		opts:    opt,
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
func (w *Worker[T]) Enqueue(ctx context.Context, payload T) error {
	name := w.opts.config.QueueName
	beg := time.Now()
	err := w.broker.Enqueue(ctx, name, payload)
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
func (w *Worker[T]) EnqueueAfter(ctx context.Context, payload T, delay time.Duration) error {
	name := w.opts.config.QueueName
	beg := time.Now()
	err := w.broker.EnqueueAfter(ctx, name, payload, delay)
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
