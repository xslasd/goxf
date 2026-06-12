package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/job/queue"
	"github.com/xslasd/goxf/log"
)

// 定义一个我们自定义的结构体，用于内存队列传递
type CleanTask struct {
	TaskID    int
	TargetDir string
	CreatedAt time.Time
}

func main() {
	srv := goxf.NewService()

	// 1. 实例化我们的基于 Go Channel 的纯内存 Broker
	memBroker := queue.NewMemoryBroker()

	// 2. 注册基于内存 Broker 的 Worker (显式指定 T 为 CleanTask)
	worker, err := queue.NewWorker(memBroker, func(ctx context.Context, task CleanTask) error {
		// 3. 在 Handler 中直接使用强类型，无需手动断言！
		fmt.Printf("[Worker] 收到清理任务: ID=%d, Dir=%s, 延迟=%v\n",
			task.TaskID, task.TargetDir, time.Since(task.CreatedAt))
		time.Sleep(1 * time.Second) // 模拟处理耗时
		return nil
	}, queue.WithWorkerNum(1))

	if err != nil {
		panic(err)
	}

	// 4. 模拟在其他地方（比如 HTTP API 或者 Cron）不断往内存队列里投递结构体数据
	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(2 * time.Second)
			task := CleanTask{
				TaskID:    i,
				TargetDir: fmt.Sprintf("/tmp/cache_%d", i),
				CreatedAt: time.Now(),
			}

			// 对于内存 Broker，我们提供了一个配套的 Enqueue 辅助方法用于直接投递
			// 直接通过刚刚返回的 worker 实例将 Struct 送入队列中，这是最标准的写法
			_, err := worker.Enqueue(context.Background(), task)
			if err != nil {
				log.Errorf("Enqueue failed: %v", err)
			}
			fmt.Printf("[Producer] 已投递任务 ID=%d\n", i)
		}
	}()

	// 运行服务，阻塞直至退出
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
