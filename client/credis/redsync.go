package credis

//  Redsync 的 Redis 适配器，核心场景是：实现 Redis 分布式锁
// 是一个用于实现分布式锁的库，它支持多个 Redis 实例，从而提高锁的可用性和性能。

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	redsync "github.com/go-redsync/redsync/v4/redis"
)

type pool struct {
	delegate redis.Cmdable
}

func (p *pool) Get(ctx context.Context) (redsync.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return &conn{p.delegate, ctx}, nil
}

func NewPool(delegate redis.Cmdable) redsync.Pool {
	return &pool{delegate}
}

type conn struct {
	delegate redis.Cmdable
	ctx      context.Context
}

func (c *conn) ScriptLoad(script *redsync.Script) error {
	hash, err := c.delegate.ScriptLoad(c.ctx, script.Src).Result()
	if err != nil {
		return err
	}
	script.Hash = hash
	return nil
}

func (c *conn) Get(name string) (string, error) {
	value, err := c.delegate.Get(c.ctx, name).Result()
	return value, noErrNil(err)
}

func (c *conn) Set(name string, value string) (bool, error) {
	reply, err := c.delegate.Set(c.ctx, name, value, 0).Result()
	return reply == "OK", err
}

func (c *conn) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	return c.delegate.SetNX(c.ctx, name, value, expiry).Result()
}

func (c *conn) PTTL(name string) (time.Duration, error) {
	return c.delegate.PTTL(c.ctx, name).Result()
}

func (c *conn) Eval(script *redsync.Script, keysAndArgs ...interface{}) (interface{}, error) {
	keys := make([]string, script.KeyCount)
	args := keysAndArgs

	if script.KeyCount > 0 {
		for i := 0; i < script.KeyCount; i++ {
			keys[i] = keysAndArgs[i].(string)
		}
		args = keysAndArgs[script.KeyCount:]
	}

	v, err := c.delegate.EvalSha(c.ctx, script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = c.delegate.Eval(c.ctx, script.Src, keys, args...).Result()
	}
	return v, noErrNil(err)
}

func (c *conn) Close() error {
	// Not needed for this library
	return nil
}

func noErrNil(err error) error {
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}
