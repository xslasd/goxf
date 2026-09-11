package golanglru

// Cache 定义通用缓存统一接口
type Cache[K comparable, V any] interface {
	// Set 写入或更新缓存条目
	Set(key K, value V)

	// Add 写入或更新缓存条目，并返回是否发生容量驱逐淘汰
	Add(key K, value V) (evicted bool)

	// Get 读取缓存条目，若不存在或已失效则返回 false
	Get(key K) (value V, ok bool)

	// Contains 判断指定 key 是否存在且未失效
	Contains(key K) bool

	// Peek 读取缓存条目但不调整其在淘汰链表中的顺序/访问频次
	Peek(key K) (value V, ok bool)

	// Remove 删除指定 key 的条目，返回之前是否存在该条目
	Remove(key K) (present bool)

	// Keys 返回当前所有未失效的 key 列表
	Keys() []K

	// Values 返回当前所有未失效的 value 列表
	Values() []V

	// Len 返回当前有效缓存条目总数
	Len() int

	// Resize 动态调整缓存容量，返回被驱逐的条目数
	Resize(size int) (evicted int)

	// Purge 清空全部缓存条目
	Purge()

	// Close 释放并关闭缓存资源（如停止后台清理协程）
	Close()
}
