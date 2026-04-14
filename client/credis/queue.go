package credis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// QueueBroker implements the job/queue.Broker interface for Redis
type QueueBroker struct {
	client      redis.Cmdable
	waitTimeout time.Duration

	// Delayed task support
	sweeperOnce   sync.Once
	activeQueues  sync.Map // map[string]struct{}
	sweeperCtx    context.Context
	sweeperCancel context.CancelFunc
}

// NewQueueBroker creates a new Redis-based queue broker.
func NewQueueBroker(client redis.Cmdable, waitTimeout time.Duration) *QueueBroker {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueBroker{
		client:        client,
		waitTimeout:   waitTimeout,
		sweeperCtx:    ctx,
		sweeperCancel: cancel,
	}
}

func (q *QueueBroker) Dequeue(ctx context.Context, queueName string) (any, error) {
	// Periodic check to ensure sweeper is running if we have delayed tasks
	q.sweeperOnce.Do(func() {
		go q.runSweeper()
	})

	res, err := q.client.BLPop(ctx, q.waitTimeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if len(res) < 2 {
		return nil, nil
	}

	return []byte(res[1]), nil
}

// Enqueue pushes a message to the end of the Redis list queue.
func (q *QueueBroker) Enqueue(ctx context.Context, queueName string, payload any) error {
	return q.client.RPush(ctx, queueName, payload).Err()
}

// EnqueueAfter schedules a message for future delivery using Redis ZSET.
func (q *QueueBroker) EnqueueAfter(ctx context.Context, queueName string, payload any, delay time.Duration) error {
	q.sweeperOnce.Do(func() {
		go q.runSweeper()
	})
	q.activeQueues.Store(queueName, struct{}{})

	delayedKey := fmt.Sprintf("%s:delayed", queueName)
	at := time.Now().Add(delay).UnixMilli()
	return q.client.ZAdd(ctx, delayedKey, &redis.Z{
		Score:  float64(at),
		Member: payload,
	}).Err()
}

func (q *QueueBroker) runSweeper() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	luaScript := `
local expired = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
if #expired > 0 then
    for i, msg in ipairs(expired) do
        redis.call('RPUSH', KEYS[2], msg)
    end
    redis.call('ZREM', KEYS[1], unpack(expired))
end
return #expired`

	for {
		select {
		case <-q.sweeperCtx.Done():
			return
		case <-ticker.C:
			now := time.Now().UnixMilli()
			q.activeQueues.Range(func(key, value any) bool {
				queueName := key.(string)
				delayedKey := fmt.Sprintf("%s:delayed", queueName)
				
				// Execute Lua script to move expired tasks atomically
				_ = q.client.Eval(q.sweeperCtx, luaScript, []string{delayedKey, queueName}, now).Err()
				return true
			})
		}
	}
}

// Close stops the background sweeper.
func (q *QueueBroker) Close() {
	if q.sweeperCancel != nil {
		q.sweeperCancel()
	}
}
