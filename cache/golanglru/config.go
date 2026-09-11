package golanglru

import (
	"strings"
	"time"

	"github.com/xslasd/goxf/application"
)

// Algorithm 表示缓存替换算法类型
type Algorithm string

const (
	AlgorithmLRU Algorithm = "lru"
	Algorithm2Q  Algorithm = "2q"
)

// Normalize 格式化并校验算法类型，默认返回 AlgorithmLRU
func (a Algorithm) Normalize() Algorithm {
	switch strings.ToLower(strings.TrimSpace(string(a))) {
	case string(Algorithm2Q):
		return Algorithm2Q
	default:
		return AlgorithmLRU
	}
}

// Config 为 golanglru 本地缓存配置
type Config struct {
	Name       string        `json:"name" yaml:"name" toml:"name"`             // 本地缓存名称
	Size       int           `json:"size" yaml:"size" toml:"size"`             // 缓存容量大小
	Expiration   time.Duration `json:"expiration" yaml:"expiration" toml:"expiration"`     // 失效时间（<=0 表示不过期）
	Algorithm    Algorithm     `json:"algorithm" yaml:"algorithm" toml:"algorithm"`       // 缓存算法，支持 lru / 2q，默认为 lru
	EnableMetric bool          `json:"enableMetric" yaml:"enableMetric" toml:"enableMetric"` // 是否开启监控埋点
}

// EvictCallback 当缓存条目被淘汰或过期清理时的回调函数
type EvictCallback[K comparable, V any] func(key K, value V)

// options 包含初始化缓存所需的全部参数与依赖
type options[K comparable, V any] struct {
	config       *Config
	confPrefix   string
	confName     string
	onEvict      EvictCallback[K, V]
	enableMetric bool
}

// Option 定义配置函数项
type Option[K comparable, V any] func(*options[K, V])

// WithConfPrefix 设置配置前缀（默认 "cache.golanglru"）
func WithConfPrefix[K comparable, V any](prefix string) Option[K, V] {
	return func(o *options[K, V]) {
		o.confPrefix = prefix
	}
}

// WithConfName 设置配置名称（对应配置文件中后缀段，默认 "default"）
func WithConfName[K comparable, V any](name string) Option[K, V] {
	return func(o *options[K, V]) {
		o.confName = name
	}
}

// WithName 设置本地缓存名称
func WithName[K comparable, V any](name string) Option[K, V] {
	return func(o *options[K, V]) {
		o.config.Name = name
	}
}

// WithSize 设置缓存容量大小
func WithSize[K comparable, V any](size int) Option[K, V] {
	return func(o *options[K, V]) {
		o.config.Size = size
	}
}

// WithExpiration 设置缓存条目失效时间
func WithExpiration[K comparable, V any](expiration time.Duration) Option[K, V] {
	return func(o *options[K, V]) {
		o.config.Expiration = expiration
	}
}

// WithTTL 设置缓存失效时间的快捷别名
func WithTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return WithExpiration[K, V](ttl)
}

// WithAlgorithm 设置缓存替换算法（支持 LRU / 2Q）
func WithAlgorithm[K comparable, V any](algo Algorithm) Option[K, V] {
	return func(o *options[K, V]) {
		o.config.Algorithm = algo
	}
}

// WithOnEvict 设置淘汰/过期回调函数
func WithOnEvict[K comparable, V any](onEvict EvictCallback[K, V]) Option[K, V] {
	return func(o *options[K, V]) {
		o.onEvict = onEvict
	}
}

// WithEnableMetric 设置是否启用监控
func WithEnableMetric[K comparable, V any](enable bool) Option[K, V] {
	return func(o *options[K, V]) {
		o.enableMetric = enable
		o.config.EnableMetric = enable
	}
}

func defaultConfig() *Config {
	return &Config{
		Name:         "default",
		Size:         1024,
		Expiration:   0,
		Algorithm:    AlgorithmLRU,
		EnableMetric: application.GetEnableMetric(),
	}
}

func defaultOptions[K comparable, V any]() *options[K, V] {
	return &options[K, V]{
		config:       defaultConfig(),
		confPrefix:   "cache.golanglru",
		confName:     "default",
		enableMetric: application.GetEnableMetric(),
	}
}
