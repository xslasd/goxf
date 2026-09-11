package golanglru

import (
	"errors"
	"fmt"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/utils/xcast"
)

// NewCache 创建支持指定泛型类型的本地缓存实例
func NewCache[K comparable, V any](opts ...Option[K, V]) (Cache[K, V], error) {
	application.CheckStartupGoxf()
	opt := defaultOptions[K, V]()
	for _, o := range opts {
		o(opt)
	}

	type rawConfig struct {
		Name         string `mapstructure:"name"`
		Size         int    `mapstructure:"size"`
		Expiration   any    `mapstructure:"expiration"`
		Algorithm    string `mapstructure:"algorithm"`
		EnableMetric *bool  `mapstructure:"enableMetric"`
	}

	key := opt.confPrefix + "." + opt.confName
	var raw rawConfig
	if err := conf.UnmarshalKey(key, &raw); err != nil {
		if errors.Is(err, conf.ErrInvalidKey) {
			log.Warn(fmt.Sprintf("cache golanglru [%s] not found in config, using defaults", key))
		} else {
			return nil, err
		}
	} else {
		if raw.Name != "" {
			opt.config.Name = raw.Name
		}
		if raw.Size > 0 {
			opt.config.Size = raw.Size
		}
		if raw.Expiration != nil {
			dur, err := xcast.ToDurationE(raw.Expiration)
			if err == nil {
				opt.config.Expiration = dur
			}
		}
		if raw.Algorithm != "" {
			opt.config.Algorithm = Algorithm(raw.Algorithm)
		}
		if raw.EnableMetric != nil {
			opt.config.EnableMetric = *raw.EnableMetric
			opt.enableMetric = *raw.EnableMetric
		}
	}

	// 允许通过 opts 显式覆盖配置文件中的值
	for _, o := range opts {
		o(opt)
	}

	if opt.config.Name == "" {
		opt.config.Name = opt.confName
	}
	if opt.config.Size <= 0 {
		opt.config.Size = 1024
	}

	// 规范化算法配置，默认为 LRU
	algo := opt.config.Algorithm.Normalize()
	opt.config.Algorithm = algo

	var (
		c   Cache[K, V]
		err error
	)

	switch algo {
	case Algorithm2Q:
		c, err = newTwoQueue[K, V](opt.config.Size, opt.config.Expiration, opt.onEvict)
		if err != nil {
			return nil, err
		}
	case AlgorithmLRU:
		fallthrough
	default:
		if opt.config.Expiration > 0 {
			c = newExpirableLRU[K, V](opt.config.Size, opt.config.Expiration, opt.onEvict)
		} else {
			c, err = newStandardLRU[K, V](opt.config.Size, opt.onEvict)
			if err != nil {
				return nil, err
			}
		}
	}

	if opt.enableMetric || opt.config.EnableMetric {
		c = wrapMetric(c, opt.config.Name, string(opt.config.Algorithm))
	}

	// 注册生命周期退出阶段的资源回收
	hooks.Register(hooks.Stage_AfterStop, func() {
		c.Close()
	})

	log.Info(fmt.Sprintf("start golanglru cache [%s] ok. algorithm: %s, size: %d, expiration: %v, metric: %v",
		opt.config.Name, opt.config.Algorithm, opt.config.Size, opt.config.Expiration, opt.enableMetric || opt.config.EnableMetric))

	return c, nil
}

// New 创建常用的 string -> any 泛型缓存实例
func New(opts ...Option[string, any]) (Cache[string, any], error) {
	return NewCache[string, any](opts...)
}
