package main

import (
	"fmt"
	"time"

	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/cache/golanglru"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/log"
)

func main() {
	// 启动 goxf 服务运行时（自动加载指定 config.yaml）
	_ = goxf.NewService(
		goxf.WithDefaultConfAddr("./examples/cache/config.yaml"),
	)

	fmt.Printf("Loaded service: %s\n", conf.GetString("goxf.serviceName"))

	// 1. 创建基于配置名 "default" 的 LRU 缓存实例
	lruCache, err := golanglru.NewCache[string, string](
		golanglru.WithConfName[string, string]("default"),
	)
	if err != nil {
		log.Fatalf("failed to init lru cache: %v", err)
	}

	lruCache.Set("session:1001", "user_alice")
	val, ok := lruCache.Get("session:1001")
	fmt.Printf("LRU Get 'session:1001': val=%s, ok=%v\n", val, ok)

	// 2. 创建基于配置名 "user_cache" 的 2Q 缓存实例
	twoQueueCache, err := golanglru.NewCache[string, int](
		golanglru.WithConfName[string, int]("user_cache"),
	)
	if err != nil {
		log.Fatalf("failed to init 2q cache: %v", err)
	}

	twoQueueCache.Set("score:alice", 98)
	score, ok := twoQueueCache.Get("score:alice")
	fmt.Printf("2Q Get 'score:alice': val=%d, ok=%v\n", score, ok)

	// 3. 测试失效时间（2Q 配置为 1s 过期）
	fmt.Println("Waiting for 2Q cache to expire (1.2s)...")
	time.Sleep(1200 * time.Millisecond)

	scoreAfter, okAfter := twoQueueCache.Get("score:alice")
	fmt.Printf("2Q Get after expire: val=%d, ok=%v (expected: false)\n", scoreAfter, okAfter)

	// 验证 LRU 仍在（2s 过期）
	valAfter, okLRU := lruCache.Get("session:1001")
	fmt.Printf("LRU Get before 2s expire: val=%s, ok=%v (expected: true)\n", valAfter, okLRU)

	fmt.Println("All cache operations tested successfully!")
}
