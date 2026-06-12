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
	broker        Broker
	handler       Handler[T]
	opts          *options
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	activeTasks   map[string]context.CancelFunc
	activeTasksMu sync.Mutex
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
			msg, err := w.broker.Dequeue(w.ctx, queueName)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				log.Errorf("worker[%d] dequeue error: %v", id, err)
				continue
			}

			if msg == nil {
				continue
			}

			w.handle(msg)
		}
	}
}

func (w *Worker[T]) handle(msg *Message) {
	beg := time.Now()
	name := w.opts.config.QueueName
	code := "ok"

	taskCtx, taskCancel := context.WithCancel(w.ctx)
	w.activeTasksMu.Lock()
	w.activeTasks[msg.ID] = taskCancel
	w.activeTasksMu.Unlock()

	defer func() {
		w.activeTasksMu.Lock()
		delete(w.activeTasks, msg.ID)
		w.activeTasksMu.Unlock()
		taskCancel()

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
	if p, ok := msg.Payload.(T); ok {
		data = p
	} else {
		// Attempt to marshal and unmarshal back to type T (e.g. from JSON map[string]any to struct)
		pb, err := json.Marshal(msg.Payload)
		if err != nil {
			log.Warn("queue job invalid payload type", log.String("queue", name), log.Any("type", fmt.Sprintf("%T", msg.Payload)))
			code = "error"
			return
		}
		if err := json.Unmarshal(pb, &data); err != nil {
			log.Warn("queue job unmarshal error", log.String("queue", name), log.FieldErr(err))
			code = "error"
			return
		}
	}

	err := w.handler(taskCtx, data)
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
