package golanglru

import (
	"time"

	"github.com/xslasd/goxf/metric"
)

// metricCache 为启用了 Prometheus 监控时的包装层
type metricCache[K comparable, V any] struct {
	Cache[K, V]
	name string
	algo string
}

func wrapMetric[K comparable, V any](c Cache[K, V], name string, algo string) Cache[K, V] {
	return &metricCache[K, V]{
		Cache: c,
		name:  name,
		algo:  algo,
	}
}

func (m *metricCache[K, V]) Get(key K) (V, bool) {
	beg := time.Now()
	val, ok := m.Cache.Get(key)
	code := "miss"
	if ok {
		code = "hit"
	}
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Get", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Get", m.algo, code)
	return val, ok
}

func (m *metricCache[K, V]) Set(key K, value V) {
	beg := time.Now()
	m.Cache.Set(key, value)
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Set", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Set", m.algo, "ok")
}

func (m *metricCache[K, V]) Add(key K, value V) bool {
	beg := time.Now()
	evicted := m.Cache.Add(key, value)
	code := "ok"
	if evicted {
		code = "evicted"
	}
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Add", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Add", m.algo, code)
	return evicted
}

func (m *metricCache[K, V]) Contains(key K) bool {
	beg := time.Now()
	ok := m.Cache.Contains(key)
	code := "miss"
	if ok {
		code = "hit"
	}
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Contains", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Contains", m.algo, code)
	return ok
}

func (m *metricCache[K, V]) Peek(key K) (V, bool) {
	beg := time.Now()
	val, ok := m.Cache.Peek(key)
	code := "miss"
	if ok {
		code = "hit"
	}
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Peek", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Peek", m.algo, code)
	return val, ok
}

func (m *metricCache[K, V]) Remove(key K) bool {
	beg := time.Now()
	present := m.Cache.Remove(key)
	code := "miss"
	if present {
		code = "hit"
	}
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Remove", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Remove", m.algo, code)
	return present
}

func (m *metricCache[K, V]) Purge() {
	beg := time.Now()
	m.Cache.Purge()
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Purge", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Purge", m.algo, "ok")
}

func (m *metricCache[K, V]) Resize(size int) int {
	beg := time.Now()
	evicted := m.Cache.Resize(size)
	metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.CacheType, m.name, "Resize", m.algo)
	metric.ClientHandleCounter.Inc(metric.CacheType, m.name, "Resize", m.algo, "ok")
	return evicted
}
