package golanglru

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type twoQueueItem[V any] struct {
	val      V
	expireAt time.Time
}

type twoQueueWrapper[K comparable, V any] struct {
	cache      *lru.TwoQueueCache[K, *twoQueueItem[V]]
	expiration time.Duration
	onEvict    EvictCallback[K, V]
	mu         sync.Mutex
	stopCh     chan struct{}
	once       sync.Once
}

func newTwoQueue[K comparable, V any](size int, expiration time.Duration, onEvict EvictCallback[K, V]) (Cache[K, V], error) {
	c, err := lru.New2Q[K, *twoQueueItem[V]](size)
	if err != nil {
		return nil, err
	}

	w := &twoQueueWrapper[K, V]{
		cache:      c,
		expiration: expiration,
		onEvict:    onEvict,
		stopCh:     make(chan struct{}),
	}

	// 若启用了失效时间，启动周期性后台清理，防止冷数据常驻
	if expiration > 0 {
		go w.startCleaner()
	}

	return w, nil
}

func (w *twoQueueWrapper[K, V]) startCleaner() {
	interval := w.expiration / 2
	if interval < 500*time.Millisecond {
		interval = 500 * time.Millisecond
	} else if interval > time.Minute {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.cleanExpired()
		}
	}
}

func (w *twoQueueWrapper[K, V]) cleanExpired() {
	w.mu.Lock()
	defer w.mu.Unlock()

	keys := w.cache.Keys()
	now := time.Now()
	for _, k := range keys {
		item, ok := w.cache.Peek(k)
		if ok && !item.expireAt.IsZero() && now.After(item.expireAt) {
			w.cache.Remove(k)
			if w.onEvict != nil {
				w.onEvict(k, item.val)
			}
		}
	}
}

func (w *twoQueueWrapper[K, V]) Set(key K, value V) {
	w.Add(key, value)
}

func (w *twoQueueWrapper[K, V]) Add(key K, value V) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	var expireAt time.Time
	if w.expiration > 0 {
		expireAt = time.Now().Add(w.expiration)
	}

	item := &twoQueueItem[V]{
		val:      value,
		expireAt: expireAt,
	}

	preLen := w.cache.Len()
	w.cache.Add(key, item)
	postLen := w.cache.Len()

	// 简要判定容量驱逐
	return preLen > 0 && postLen <= preLen
}

func (w *twoQueueWrapper[K, V]) Get(key K) (V, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var zero V
	item, ok := w.cache.Get(key)
	if !ok {
		return zero, false
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		w.cache.Remove(key)
		if w.onEvict != nil {
			w.onEvict(key, item.val)
		}
		return zero, false
	}

	return item.val, true
}

func (w *twoQueueWrapper[K, V]) Contains(key K) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	item, ok := w.cache.Peek(key)
	if !ok {
		return false
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		w.cache.Remove(key)
		if w.onEvict != nil {
			w.onEvict(key, item.val)
		}
		return false
	}

	return true
}

func (w *twoQueueWrapper[K, V]) Peek(key K) (V, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var zero V
	item, ok := w.cache.Peek(key)
	if !ok {
		return zero, false
	}

	if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
		w.cache.Remove(key)
		if w.onEvict != nil {
			w.onEvict(key, item.val)
		}
		return zero, false
	}

	return item.val, true
}

func (w *twoQueueWrapper[K, V]) Remove(key K) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	item, ok := w.cache.Peek(key)
	if !ok {
		return false
	}

	w.cache.Remove(key)
	if w.onEvict != nil {
		w.onEvict(key, item.val)
	}
	return true
}

func (w *twoQueueWrapper[K, V]) Keys() []K {
	w.mu.Lock()
	defer w.mu.Unlock()

	keys := w.cache.Keys()
	now := time.Now()
	validKeys := make([]K, 0, len(keys))
	for _, k := range keys {
		item, ok := w.cache.Peek(k)
		if ok {
			if !item.expireAt.IsZero() && now.After(item.expireAt) {
				w.cache.Remove(k)
				if w.onEvict != nil {
					w.onEvict(k, item.val)
				}
				continue
			}
			validKeys = append(validKeys, k)
		}
	}
	return validKeys
}

func (w *twoQueueWrapper[K, V]) Values() []V {
	w.mu.Lock()
	defer w.mu.Unlock()

	keys := w.cache.Keys()
	now := time.Now()
	validVals := make([]V, 0, len(keys))
	for _, k := range keys {
		item, ok := w.cache.Peek(k)
		if ok {
			if !item.expireAt.IsZero() && now.After(item.expireAt) {
				w.cache.Remove(k)
				if w.onEvict != nil {
					w.onEvict(k, item.val)
				}
				continue
			}
			validVals = append(validVals, item.val)
		}
	}
	return validVals
}

func (w *twoQueueWrapper[K, V]) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.cache.Len()
}

func (w *twoQueueWrapper[K, V]) Resize(size int) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.cache.Resize(size)
}

func (w *twoQueueWrapper[K, V]) Purge() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.cache.Purge()
}

func (w *twoQueueWrapper[K, V]) Close() {
	w.once.Do(func() {
		close(w.stopCh)
	})
	w.Purge()
}
