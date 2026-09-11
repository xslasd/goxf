package golanglru

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/log"
)

func TestMain(m *testing.M) {
	application.NewRuntime("test-app", "test-service", false, false, false, false, false)
	log.SetLogger(log.NewLogger())
	os.Exit(m.Run())
}

// 测试 LRU 模式默认无过期行为
func TestLRUCacheBasic(t *testing.T) {
	c, err := NewCache[string, int](
		WithName[string, int]("test-lru-basic"),
		WithSize[string, int](3),
		WithAlgorithm[string, int](AlgorithmLRU),
	)
	assert.NoError(t, err)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	val, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	assert.True(t, c.Contains("b"))
	assert.Equal(t, 3, c.Len())

	// 测试 Peek 不影响访问
	peekVal, ok := c.Peek("c")
	assert.True(t, ok)
	assert.Equal(t, 3, peekVal)

	// 测试超出容量自动淘汰最久未使用的
	c.Set("d", 4)
	assert.Equal(t, 3, c.Len())
	// b 是最久未被访问的条目（因为前面访问了 a）
	assert.False(t, c.Contains("b"))
	assert.True(t, c.Contains("d"))

	// 测试 Remove
	assert.True(t, c.Remove("a"))
	assert.False(t, c.Contains("a"))

	// 测试 Purge
	c.Purge()
	assert.Equal(t, 0, c.Len())
}

// 测试 LRU 模式配合 TTL 失效时间
func TestLRUCacheExpiration(t *testing.T) {
	var evictedKey string
	var evictedVal string
	var mu sync.Mutex

	c, err := NewCache[string, string](
		WithName[string, string]("test-lru-ttl"),
		WithSize[string, string](10),
		WithExpiration[string, string](100*time.Millisecond),
		WithAlgorithm[string, string](AlgorithmLRU),
		WithOnEvict[string, string](func(k string, v string) {
			mu.Lock()
			evictedKey = k
			evictedVal = v
			mu.Unlock()
		}),
	)
	assert.NoError(t, err)
	defer c.Close()

	c.Set("token", "abc-123")
	val, ok := c.Get("token")
	assert.True(t, ok)
	assert.Equal(t, "abc-123", val)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	_, ok = c.Get("token")
	assert.False(t, ok, "key 应当已过期")
	assert.False(t, c.Contains("token"))

	mu.Lock()
	_ = evictedKey
	_ = evictedVal
	mu.Unlock()
}

// 测试 2Q 算法基础操作
func Test2QCacheBasic(t *testing.T) {
	c, err := NewCache[string, string](
		WithName[string, string]("test-2q-basic"),
		WithSize[string, string](5),
		WithAlgorithm[string, string](Algorithm2Q),
	)
	assert.NoError(t, err)
	defer c.Close()

	c.Set("k1", "v1")
	c.Set("k2", "v2")

	v, ok := c.Get("k1")
	assert.True(t, ok)
	assert.Equal(t, "v1", v)

	assert.True(t, c.Contains("k2"))
	assert.False(t, c.Contains("k3"))

	keys := c.Keys()
	assert.Len(t, keys, 2)

	assert.True(t, c.Remove("k1"))
	assert.False(t, c.Contains("k1"))
	assert.Equal(t, 1, c.Len())
}

// 测试 2Q 算法配合 TTL 失效时间
func Test2QCacheExpiration(t *testing.T) {
	var evictedKeys []string
	var mu sync.Mutex

	c, err := NewCache[string, int](
		WithName[string, int]("test-2q-ttl"),
		WithSize[string, int](10),
		WithExpiration[string, int](100*time.Millisecond),
		WithAlgorithm[string, int](Algorithm2Q),
		WithOnEvict[string, int](func(k string, v int) {
			mu.Lock()
			evictedKeys = append(evictedKeys, k)
			mu.Unlock()
		}),
	)
	assert.NoError(t, err)
	defer c.Close()

	c.Set("num1", 100)
	c.Set("num2", 200)

	val, ok := c.Get("num1")
	assert.True(t, ok)
	assert.Equal(t, 100, val)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 验证 Get 返回 false 并清理
	_, ok = c.Get("num1")
	assert.False(t, ok, "num1 应当已过期")
	assert.False(t, c.Contains("num2"), "num2 应当已过期")

	// 检查 Keys / Values 过滤
	keys := c.Keys()
	assert.Empty(t, keys)

	mu.Lock()
	assert.Contains(t, evictedKeys, "num1")
	mu.Unlock()
}

// 测试并发安全性
func TestConcurrent(t *testing.T) {
	c, err := New(
		WithName[string, any]("test-concurrent"),
		WithSize[string, any](100),
		WithExpiration[string, any](200*time.Millisecond),
		WithAlgorithm[string, any](Algorithm2Q),
	)
	assert.NoError(t, err)
	defer c.Close()

	var wg sync.WaitGroup
	workers := 20
	iterations := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("k-%d-%d", workerID, j%10)
				c.Set(key, j)
				_, _ = c.Get(key)
				_ = c.Contains(key)
				if j%5 == 0 {
					_ = c.Remove(key)
				}
			}
		}(i)
	}

	wg.Wait()
}

// 测试默认算法回退与大小写
func TestAlgorithmDefaults(t *testing.T) {
	assert.Equal(t, AlgorithmLRU, Algorithm("").Normalize())
	assert.Equal(t, AlgorithmLRU, Algorithm("unknown").Normalize())
	assert.Equal(t, AlgorithmLRU, Algorithm("LRU").Normalize())
	assert.Equal(t, Algorithm2Q, Algorithm("2Q").Normalize())
	assert.Equal(t, Algorithm2Q, Algorithm("2q").Normalize())
}

// 测试开启 Metric 监控埋点功能
func TestEnableMetric(t *testing.T) {
	c, err := NewCache[string, string](
		WithName[string, string]("test-metric-cache"),
		WithSize[string, string](10),
		WithEnableMetric[string, string](true),
	)
	assert.NoError(t, err)
	defer c.Close()

	// 验证底层被正确包装为 metricCache
	_, isMetric := c.(*metricCache[string, string])
	assert.True(t, isMetric)

	// 验证各项监控埋点调用正常运行
	c.Set("k1", "v1")
	c.Add("k2", "v2")

	val, ok := c.Get("k1")
	assert.True(t, ok)
	assert.Equal(t, "v1", val)

	_, ok = c.Get("not_exist")
	assert.False(t, ok)

	assert.True(t, c.Contains("k2"))
	assert.False(t, c.Contains("not_exist"))

	pVal, pOk := c.Peek("k1")
	assert.True(t, pOk)
	assert.Equal(t, "v1", pVal)

	assert.True(t, c.Remove("k1"))
	assert.False(t, c.Remove("not_exist"))

	c.Resize(20)
	c.Purge()
	assert.Equal(t, 0, c.Len())
}
