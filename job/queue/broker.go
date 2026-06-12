package queue

import (
	"context"
	"time"
)

// Message 包装了流转于任务队列中的消息实体
type Message struct {
	ID      string // 任务的唯一标识 ID (通常由系统生成)
	Payload any    // 任务的实际数据载荷
}

// Broker 定义了底层消息队列引擎的通用抽象接口
type Broker interface {
	// Dequeue 从指定队列中取出一个任务消息。若队列中无可用任务，会阻塞等待直到新任务到来或 Context 被取消。
	Dequeue(ctx context.Context, queueName string) (*Message, error)

	// Enqueue 将一条任务消息立即发布/推送到指定的队列中。
	Enqueue(ctx context.Context, queueName string, msg *Message) error

	// EnqueueAfter 调度一个延迟任务消息，使其在指定的延迟时间 delay 过去之后，才会被投递到普通队列被消费。
	EnqueueAfter(ctx context.Context, queueName string, msg *Message, delay time.Duration) error

	// Delete 根据任务 ID，将尚在延迟队列中等待或已在队列通道中排队但未执行的任务取消/标记删除。
	Delete(ctx context.Context, queueName string, id string) error
}

// Handler 定义了处理强类型任务数据 T 的回调执行函数。
// 当任务被执行时，如果该任务被外部取消，传入的 ctx 会收到 Done() 信号，Handler 需感知并及时退出。
type Handler[T any] func(ctx context.Context, payload T) error
