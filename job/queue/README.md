# 任务队列模块 (job/queue)

`job/queue` 是 `goxf` 框架提供的轻量级、强类型后台任务队列组件。它对底层队列引擎（Broker）进行了高度抽象，支持强类型的任务处理（Handler）、并发处理（Worker）以及延迟任务投递。

## 核心概念

- **Broker**: 队列引擎的接口抽象，负责底层消息的入队 (`Enqueue`)、出队 (`Dequeue`) 和延迟入队 (`EnqueueAfter`)。
- **Handler[T]**: 任务的具体执行函数。使用 Go 泛型，您无需在处理函数中手动进行 `interface{}` 类型断言，即可直接拿到强类型 `T` 类型的任务数据。
- **Worker[T]**: 任务队列的实际工作节点。它将底层的 `Broker` 和您注册的 `Handler` 关联起来，控制并发拉取和执行任务，并参与服务的优雅退出生命周期。

## 配置项

模块默认使用配置系统加载配置，对应的配置结构体为 `Config`：

| 配置项 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `QueueName` | `string` | `"default"` | 队列的名称。如果不配置，默认会使用选项中的 `confName`。 |
| `WorkerNum` | `int` | `1` | 并发执行任务的协程数（并发数）。 |

### 配置文件结构

默认加载的配置前缀为 `job.queue.<confName>`，例如在 `app.yaml` 中配置：

```yaml
job:
  queue:
    default:
      QueueName: "my_task_queue"
      WorkerNum: 4 # 开启4个并发协程处理任务
```

## 快速上手

这里以纯内存队列（基于 Channel）为例，展示如何定义任务结构、创建 Worker 并投递任务。

### 1. 定义任务结构和处理器

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/job/queue"
)

// Define your task payload struct
type EmailTask struct {
	EmailID   string
	Recipient string
	Content   string
}

func main() {
	// 初始化 goxf 服务
	srv := goxf.NewService()

	// 1. 实例化基于 Go Channel 的内存 Broker
	memBroker := queue.NewMemoryBroker()

	// 2. 注册并启动 Worker (显式指定泛型参数为 EmailTask)
	worker, err := queue.NewWorker(memBroker, func(ctx context.Context, task EmailTask) error {
		// 直接处理强类型数据，无需 interface{} 转换
		fmt.Printf("[Worker] 开始发送邮件: ID=%s to %s\n", task.EmailID, task.Recipient)
		time.Sleep(500 * time.Millisecond) // 模拟处理耗时
		return nil
	}, queue.WithWorkerNum(2)) // 开启 2 个并发协程
    
	if err != nil {
		panic(err)
	}

	// 3. 模拟生产任务
	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(1 * time.Second)
			task := EmailTask{
				EmailID:   fmt.Sprintf("msg-%d", i),
				Recipient: fmt.Sprintf("user%d@example.com", i),
				Content:   "Hello from goxf!",
			}
            
			// 投递即时任务
			err := worker.Enqueue(context.Background(), task)
			if err != nil {
				fmt.Printf("Enqueue err: %v\n", err)
			}
		}
	}()

	// 启动服务，当进程接收到退出信号时，Worker 会自动优雅停止并等待进行中的任务执行完毕
	if err := srv.Run(); err != nil {
		panic(err)
	}
}
```

## 接口与方法说明

### 1. 创建 Worker

```go
func NewWorker[T any](broker Broker, handler Handler[T], opts ...Option) (*Worker[T], error)
```
- `broker`: 实现了 `Broker` 接口的队列引擎（例如：`MemoryBroker` 或自定义 Redis/RabbitMQ 等 Broker）。
- `handler`: 任务回调函数，类型为 `func(ctx context.Context, payload T) error`。
- `opts`: 支持以下配置选项：
  - `WithWorkerNum(num int)`: 设置并发处理任务协程数。
  - `WithConfPrefix(prefix string)`: 自定义配置文件前缀（默认 `job.queue`）。
  - `WithConfName(name string)`: 自定义配置名（默认 `default`）。
  - `WithEnableMetric(enable bool)`: 是否启用 Metric 统计。

> [!NOTE]
> 在创建 Worker 时，会自动在框架的 `AfterStop` 阶段注册优雅退出 Hook，服务退出时将阻塞并等待所有正在执行的 Handler 任务运行完毕后再关闭。

### 2. 投递即时任务

```go
func (w *Worker[T]) Enqueue(ctx context.Context, payload T) (string, error)
func (w *Worker[T]) EnqueueWithID(ctx context.Context, taskID string, payload T) error
```
- `Enqueue`：将类型为 `T` 的 payload 立即投递到当前 Worker 所绑定的队列中，并返回自动生成的任务唯一 ID（`taskID`）。
- `EnqueueWithID`：允许调用方传入自定义的 `taskID` 并立即投递任务（可用于与业务系统订单 ID 等绑定，便于后续取消或防重）。

### 3. 投递延迟任务

```go
func (w *Worker[T]) EnqueueAfter(ctx context.Context, payload T, delay time.Duration) (string, error)
func (w *Worker[T]) EnqueueAfterWithID(ctx context.Context, taskID string, payload T, delay time.Duration) error
```
- `EnqueueAfter`：将类型为 `T` 的 payload 投递到当前队列中，并在 `delay` 时间之后才会被消费执行。返回自动生成的任务唯一 ID。
- `EnqueueAfterWithID`：允许调用方传入自定义的 `taskID` 投递延迟任务。
> [!NOTE]
> 基于内存的 `MemoryBroker` 内部利用最小堆 (`min-heap`) 和单协程触发器 (`sweeper`) 实现了精准的高性能内存延迟任务派发。

### 4. 取消执行任务

```go
func (w *Worker[T]) Cancel(ctx context.Context, taskID string) error
```
支持在以下两种场景下取消任务：
1. **延迟任务尚未开始执行**：自动在底层的 Broker 中将其标记删除或移出，出队或被 sweeper 扫描时会被直接过滤拦截；
2. **任务目前正在运行**：向当前任务所独立绑定的 Context 派发取消信号，从而让正在运行的 Handler 能够通过 `ctx.Done()` 中断执行。
