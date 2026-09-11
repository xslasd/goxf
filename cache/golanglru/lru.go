package golanglru

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

// expirableLRUWrapper 包装带失效时间的 LRU 缓存
type expirableLRUWrapper[K comparable, V any] struct {
	cache *expirable.LRU[K, V]
}

func newExpirableLRU[K comparable, V any](size int, expiration time.Duration, onEvict EvictCallback[K, V]) Cache[K, V] {
	var evictCb expirable.EvictCallback[K, V]
	if onEvict != nil {
		evictCb = expirable.EvictCallback[K, V](onEvict)
	}

	c := expirable.NewLRU[K, V](size, evictCb, expiration)
	return &expirableLRUWrapper[K, V]{
		cache: c,
	}
}

func (w *expirableLRUWrapper[K, V]) Set(key K, value V) {
	w.cache.Add(key, value)
}

func (w *expirableLRUWrapper[K, V]) Add(key K, value V) bool {
	return w.cache.Add(key, value)
}

func (w *expirableLRUWrapper[K, V]) Get(key K) (V, bool) {
	return w.cache.Get(key)
}

func (w *expirableLRUWrapper[K, V]) Contains(key K) bool {
	return w.cache.Contains(key)
}

func (w *expirableLRUWrapper[K, V]) Peek(key K) (V, bool) {
	return w.cache.Peek(key)
}

func (w *expirableLRUWrapper[K, V]) Remove(key K) bool {
	return w.cache.Remove(key)
}

func (w *expirableLRUWrapper[K, V]) Keys() []K {
	return w.cache.Keys()
}

func (w *expirableLRUWrapper[K, V]) Values() []V {
	return w.cache.Values()
}

func (w *expirableLRUWrapper[K, V]) Len() int {
	return w.cache.Len()
}

func (w *expirableLRUWrapper[K, V]) Resize(size int) int {
	return w.cache.Resize(size)
}

func (w *expirableLRUWrapper[K, V]) Purge() {
	w.cache.Purge()
}

func (w *expirableLRUWrapper[K, V]) Close() {
	w.cache.Purge()
}

// standardLRUWrapper 包装标准无过期淘汰的 LRU 缓存
type standardLRUWrapper[K comparable, V any] struct {
	cache *lru.Cache[K, V]
}

func newStandardLRU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (Cache[K, V], error) {
	var cb func(key K, value V)
	if onEvict != nil {
		cb = func(key K, value V) {
			onEvict(key, value)
		}
	}

	c, err := lru.NewWithEvict[K, V](size, cb)
	if err != nil {
		return nil, err
	}

	return &standardLRUWrapper[K, V]{
		cache: c,
	}, nil
}

func (w *standardLRUWrapper[K, V]) Set(key K, value V) {
	w.cache.Add(key, value)
}

func (w *standardLRUWrapper[K, V]) Add(key K, value V) bool {
	return w.cache.Add(key, value)
}

func (w *standardLRUWrapper[K, V]) Get(key K) (V, bool) {
	return w.cache.Get(key)
}

func (w *standardLRUWrapper[K, V]) Contains(key K) bool {
	return w.cache.Contains(key)
}

func (w *standardLRUWrapper[K, V]) Peek(key K) (V, bool) {
	return w.cache.Peek(key)
}

func (w *standardLRUWrapper[K, V]) Remove(key K) bool {
	return w.cache.Remove(key)
}

func (w *standardLRUWrapper[K, V]) Keys() []K {
	return w.cache.Keys()
}

func (w *standardLRUWrapper[K, V]) Values() []V {
	return w.cache.Values()
}

func (w *standardLRUWrapper[K, V]) Len() int {
	return w.cache.Len()
}

func (w *standardLRUWrapper[K, V]) Resize(size int) int {
	return w.cache.Resize(size)
}

func (w *standardLRUWrapper[K, V]) Purge() {
	w.cache.Purge()
}

func (w *standardLRUWrapper[K, V]) Close() {
	w.cache.Purge()
}
