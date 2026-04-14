package queue

import (
	"context"
	"time"
)

type Broker interface {
	Dequeue(ctx context.Context, queueName string) (any, error)
	Enqueue(ctx context.Context, queueName string, payload any) error
	EnqueueAfter(ctx context.Context, queueName string, payload any, delay time.Duration) error
}

type Handler[T any] func(ctx context.Context, payload T) error
