package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/client/credis"
	"github.com/xslasd/goxf/job/queue"
	"github.com/xslasd/goxf/log"
)

func main() {
	srv := goxf.NewService()

	// 初始化 Redis 客户端
	redisClient, err := credis.NewClient()
	if err != nil {
		panic(err)
	}

	// 将 Redis 封装为 Broker
	queueBroker := credis.NewQueueBroker(redisClient, 5*time.Second)

	// 注册一个异步任务消费者 (Worker)
	// 监听独立配置的队列: job.queue.order_process.queue_name
	worker, err := queue.NewWorker(queueBroker, func(ctx context.Context, payload any) error {
		// 断言出底层的数据格式
		var payloadStr string
		if pBytes, ok := payload.([]byte); ok {
			payloadStr = string(pBytes)
		} else {
			payloadStr = fmt.Sprintf("%v", payload)
		}

		fmt.Printf("Processing order: %s at %v\n", payloadStr, time.Now())
		time.Sleep(time.Duration(2) * time.Second) // 模拟业务处理
		return nil
	}, queue.WithConfName("order_process"))
	if err != nil {
		panic(err)
	}

	// 示例：借助 worker 实例进行生产投递，内部已处理好 QueueName、日志和指标
	go func() {
		for {
			time.Sleep(3 * time.Second)
			_ = worker.Enqueue(context.Background(), []byte("ORDER_12345"))
		}
	}()

	// 运行服务
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
