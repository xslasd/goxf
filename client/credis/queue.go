package credis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xslasd/goxf/job/queue"
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

func (q *QueueBroker) Dequeue(ctx context.Context, queueName string) (*queue.Message, error) {
	// Periodic check to ensure sweeper is running if we have delayed tasks
	q.sweeperOnce.Do(func() {
		go q.runSweeper()
	})

	for {
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

		var msg queue.Message
		if err := json.Unmarshal([]byte(res[1]), &msg); err != nil {
			return nil, err
		}

		// Check if the task has been cancelled
		deletedKey := fmt.Sprintf("%s:deleted", queueName)
		isDeleted, err := q.client.SIsMember(ctx, deletedKey, msg.ID).Result()
		if err == nil && isDeleted {
			q.client.SRem(ctx, deletedKey, msg.ID)
			continue
		}

		return &msg, nil
	}
}

// Enqueue pushes a message to the end of the Redis list queue.
func (q *QueueBroker) Enqueue(ctx context.Context, queueName string, msg *queue.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, queueName, data).Err()
}

// EnqueueAfter schedules a message for future delivery using Redis ZSET.
func (q *QueueBroker) EnqueueAfter(ctx context.Context, queueName string, msg *queue.Message, delay time.Duration) error {
	q.sweeperOnce.Do(func() {
		go q.runSweeper()
	})
	q.activeQueues.Store(queueName, struct{}{})

	delayedKey := fmt.Sprintf("%s:delayed", queueName)
	at := time.Now().Add(delay).UnixMilli()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.client.ZAdd(ctx, delayedKey, &redis.Z{
		Score:  float64(at),
		Member: data,
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

// Delete marks a message ID as deleted in Redis.
func (q *QueueBroker) Delete(ctx context.Context, queueName string, id string) error {
	deletedKey := fmt.Sprintf("%s:deleted", queueName)
	pipe := q.client.TxPipeline()
	pipe.SAdd(ctx, deletedKey, id)
	pipe.Expire(ctx, deletedKey, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// Close stops the background sweeper.
func (q *QueueBroker) Close() {
	if q.sweeperCancel != nil {
		q.sweeperCancel()
	}
}
