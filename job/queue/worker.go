package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
)

type Worker[T any] struct {
	broker  Broker
	handler Handler[T]
	opts    *options
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func (w *Worker[T]) run() error {
	w.ctx, w.cancel = context.WithCancel(context.Background())
	for i := 0; i < w.opts.config.WorkerNum; i++ {
		w.wg.Add(1)
		go w.work(i)
	}
	log.Infof("job queue worker[%s] start ok, workers: %d", w.opts.config.QueueName, w.opts.config.WorkerNum)
	return nil
}

func (w *Worker[T]) work(id int) {
	defer w.wg.Done()
	queueName := w.opts.config.QueueName
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			payload, err := w.broker.Dequeue(w.ctx, queueName)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				log.Errorf("worker[%d] dequeue error: %v", id, err)
				continue
			}

			if payload == nil {
				continue
			}

			w.handle(payload)
		}
	}
}

func (w *Worker[T]) handle(payload any) {
	beg := time.Now()
	name := w.opts.config.QueueName
	code := "ok"
	defer func() {
		if err := recover(); err != nil {
			log.Error("queue job panic", log.String("queue", name), log.Any("err", err))
			code = "panic"
		}
		log.Debug("queue job finish", log.String("queue", name), log.FieldCost(time.Since(beg)))
		if w.opts.enableMetric {
			metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.JobType, "handle", name)
			metric.ServerHandleCounter.Inc(metric.JobType, name, "handle", code)
		}
	}()

	var data T
	if p, ok := payload.(T); ok {
		data = p
	} else if pb, ok := payload.([]byte); ok {
		if err := json.Unmarshal(pb, &data); err != nil {
			log.Warn("queue job unmarshal error", log.String("queue", name), log.FieldErr(err))
			code = "error"
			return
		}
	} else {
		log.Warn("queue job invalid payload type", log.String("queue", name), log.Any("type", fmt.Sprintf("%T", payload)))
		code = "error"
		return
	}

	err := w.handler(w.ctx, data)
	if err != nil {
		log.Warn("queue job handler error", log.String("queue", name), log.FieldErr(err))
		code = "error"
	}
}

func (w *Worker[T]) stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	log.Infof("job queue worker[%s] stop ok", w.opts.config.QueueName)
}
