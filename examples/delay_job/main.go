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

	// 2. 注册消费者，支持监听 Context 取消信号
	worker, err := queue.NewWorker(memBroker, func(ctx context.Context, payload string) error {
		fmt.Printf("[%s] [Worker] 开始执行任务: %s\n", time.Now().Format("15:04:05"), payload)
		
		// 模拟耗时任务并监听取消
		select {
		case <-time.After(3 * time.Second):
			fmt.Printf("[%s] [Worker] 任务执行完成: %s\n", time.Now().Format("15:04:05"), payload)
		case <-ctx.Done():
			fmt.Printf("[%s] [Worker] 任务执行被取消: %s (原因: %v)\n", time.Now().Format("15:04:05"), payload, ctx.Err())
		}
		return nil
	}, queue.WithConfName("delay_test"))

	if err != nil {
		panic(err)
	}

	// 3. 演示情况 1：取消尚未执行的延迟任务
	fmt.Printf("[%s] 投递延迟任务: 5秒后执行...\n", time.Now().Format("15:04:05"))
	delayTaskID, err := worker.EnqueueAfter(context.Background(), "延迟任务内容", 5*time.Second)
	if err != nil {
		log.Errorf("EnqueueAfter failed: %v", err)
	}

	// 2秒后取消该延迟任务（它还有3秒才到期，应该会被成功取消且不打印“开始执行”）
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] [Cancel] 正在取消延迟任务: ID=%s\n", time.Now().Format("15:04:05"), delayTaskID)
		_ = worker.Cancel(context.Background(), delayTaskID)
	}()

	// 4. 演示情况 2：取消正在执行的任务
	go func() {
		// 等待延迟任务取消演示完成
		time.Sleep(6 * time.Second)

		fmt.Printf("[%s] 投递长任务（立即执行）...\n", time.Now().Format("15:04:05"))
		longTaskID, err := worker.Enqueue(context.Background(), "长任务内容")
		if err != nil {
			log.Errorf("Enqueue failed: %v", err)
			return
		}

		// 1秒后取消这个正在执行的长任务
		time.Sleep(1 * time.Second)
		fmt.Printf("[%s] [Cancel] 正在取消正在运行的任务: ID=%s\n", time.Now().Format("15:04:05"), longTaskID)
		_ = worker.Cancel(context.Background(), longTaskID)
	}()

	// 运行服务
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
