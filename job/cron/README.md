# 定时任务模块 (job/cron)

`job/cron` 是 `goxf` 框架提供的定时任务组件。它封装了优秀的定时任务库 `github.com/robfig/cron/v3`，并将其无缝接入了框架的配置系统、监控指标（Metric）以及服务优雅退出生命周期中。

## 核心功能

- **生命周期托管**：定时任务在初始化时会自动注册到框架的 `AfterStop` 生命周期 Hook 中。在服务收到退出信号时，调度器会自动停止，防止新任务的触发。
- **配置系统联动**：支持将 Spec 定时表达式配置在 YAML 配置文件中，做到无需重新编译即可动态调整执行频率。
- **监控指标集成**：内置了任务耗时（Histogram）及执行结果计数器（Counter）的 Metric 监控，方便追踪任务健康状态。

## 配置项说明

模块支持通过框架配置系统自动加载任务执行频率。每个定时任务通过 `confName`（默认为 `"default"`) 来做标识，对应的配置项结构体为 `Config`：

| 配置项 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `Spec` | `string` | `""` | Cron 定时表达式，如 `@every 5s` 或 `*/5 * * * * *`。 |

默认的配置文件路径前缀为 `job.cron.<confName>`。例如在 `app.yaml` 中配置：

```yaml
job:
  cron:
    task1:
      Spec: "@every 10s" # 每10秒执行一次
    task2:
      Spec: "0 0 12 * * ?" # 每天中午12点执行
```

## 快速上手

下面展示了两种注册定时任务的方法：
1. **代码内直接指定频率**（常用于测试或固定不可变频率的任务）；
2. **通过配置文件加载频率**（生产环境推荐的做法）。

```go
package main

import (
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/job/cron"
	"github.com/xslasd/goxf/log"
)

func main() {
	// 初始化服务
	srv := goxf.NewService()

	// 方式 1：从代码内直接指定 Spec 定时规则
	err := cron.NewCron(func() {
		fmt.Printf("[Task1] 执行时间: %s\n", time.Now().Format("15:04:05"))
	}, cron.WithConfName("task1"), cron.WithSpec("@every 1s")) // 配合 WithSpec 直接定义
	if err != nil {
		panic(err)
	}

	// 方式 2：不指定 Spec，自动从配置文件中加载 job.cron.task2.Spec 规则
	err = cron.NewCron(func() {
		fmt.Printf("[Task2] 执行时间: %s\n", time.Now().Format("15:04:05"))
	}, cron.WithConfName("task2"))
	if err != nil {
		panic(err)
	}

	// 启动服务（Cron 在 NewCron 调用后即在后台开始工作，通过 Run 阻塞进程监听退出信号）
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

## API 选项说明

在创建定时任务时，通过 `NewCron` 函数的 Option 参数进行定制：

```go
func NewCron(cmd func(), opts ...Option) error
```

### 选项列表：
- `WithSpec(spec string)`: 直接指定定时表达式（代码内最高优先级，将覆盖配置文件的配置）。
- `WithConfName(name string)`: 指定该定时任务的名称（默认 `"default"`），这决定了它加载配置文件中的哪个 Key（`job.cron.<confName>`）。
- `WithConfPrefix(prefix string)`: 自定义配置文件前缀（默认 `"job.cron"`）。
- `WithEnableMetric(enable bool)`: 是否为此定时任务开启 Metric 监控指标（默认与全局 Metric 开关一致）。
- `WithCronOption(opts ...cron.Option)`: 向底层 `robfig/cron/v3` 调度器传递原始的设置（如时区设置 `cron.WithLocation` 等）。

---

> [!NOTE]
> 默认情况下，底层调度器启用了秒级精度（`cron.WithSeconds()`）。若需要使用标准的五位 Cron 表达式，您可以通过 `WithCronOption` 进行覆盖调整。
