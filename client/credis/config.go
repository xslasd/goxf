package credis

import (
	"context"
	"crypto/tls"
	"github.com/xslasd/goxf/application"
	"time"
)

type Config struct {
	Network    string
	Addrs      []string
	Username   string
	Password   string
	DB         int
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration
	CertFile           string
	KeyFile            string
	CaCert             string
}

type clientOption struct {
	config        *Config
	context       context.Context
	tls           *tls.Config
	confPrefix    string
	confName      string
	enableConsole bool
	enableTrace   bool
	enableMetric  bool
}

type Option func(*clientOption)

func Context(ctx context.Context) Option {
	return func(o *clientOption) {
		o.context = ctx
	}
}

func WithConfPrefix(prefix string) Option {
	return func(o *clientOption) {
		o.confPrefix = prefix
	}
}

func WithConfName(name string) Option {
	return func(o *clientOption) {
		o.confName = name
	}
}

func WithEnableConsole(enableConsole bool) Option {
	return func(o *clientOption) {
		o.enableConsole = enableConsole
	}
}

func WithEnableMetric(enableMetric bool) Option {
	return func(o *clientOption) {
		o.enableMetric = enableMetric
	}
}

func WithEnableTrace(enableTrace bool) Option {
	return func(o *clientOption) {
		o.enableTrace = enableTrace
	}
}

func defaultConfig() *Config {
	return &Config{
		DialTimeout:  3 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		DB:           0,
	}
}

func defaultOptions() *clientOption {
	return &clientOption{
		config:        defaultConfig(),
		context:       context.Background(),
		confPrefix:    "client.redis",
		confName:      "default",
		enableConsole: application.GetEnableConsole(),
		enableTrace:   application.GetEnableTrace(),
		enableMetric:  application.GetEnableMetric(),
	}
}
