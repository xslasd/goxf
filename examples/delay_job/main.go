package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/job/queue"
	"github.com/xslasd/goxf/log"
)

func main() {
	srv := goxf.NewService()

	// 1. 使用内存 Broker 进行快速测试
	memBroker := queue.NewMemoryBroker()

	// 2. 注册消费者
	worker, err := queue.NewWorker(memBroker, func(ctx context.Context, payload string) error {
		fmt.Printf("[%s] 消费成功: %s\n", time.Now().Format("15:04:05"), payload)
		return nil
	}, queue.WithConfName("delay_test"))

	if err != nil {
		panic(err)
	}

	// 3. 投递延迟任务
	fmt.Printf("[%s] 投递延迟任务: 5秒后执行...\n", time.Now().Format("15:04:05"))
	
	err = worker.EnqueueAfter(context.Background(), "Hello Delayed World!", 5*time.Second)
	if err != nil {
		log.Errorf("EnqueueAfter failed: %v", err)
	}

	// 投递一个对比任务（立即执行）
	fmt.Printf("[%s] 投递普通任务: 立即执行...\n", time.Now().Format("15:04:05"))
	_ = worker.Enqueue(context.Background(), "Hello Instant World!")

	// 运行服务
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
