package main

import (
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/job/cron"
	"github.com/xslasd/goxf/log"
)

func main() {
	srv := goxf.NewService()

	// 任务1：从代码直接指定 spec。调用后立即在后台运行，并由框架托管生命周期。
	err := cron.NewCron(func() {
		fmt.Printf("Task1 running at %s\n", time.Now().Format("15:04:05"))
	}, cron.WithConfName("task1"), cron.WithSpec("@every 1s"))
	if err != nil {
		panic(err)
	}

	// 任务2：假设从配置文件读取 job.cron.task2.spec
	err = cron.NewCron(func() {
		fmt.Printf("Task2 running at %s\n", time.Now().Format("15:04:05"))
	}, cron.WithConfName("task2"))
	if err != nil {
		panic(err)
	}

	// 运行服务。虽然 cron 已经在运行，但仍需要调用 Run 阻塞主进程并监听信号
	if err := srv.Run(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
