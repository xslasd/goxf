package cetcd

import (
	"context"
	"crypto/tls"
	"github.com/xslasd/goxf/application"
	"google.golang.org/grpc"
	"time"
)

type Config struct {
	Addrs                []string
	DialTimeout          time.Duration
	AutoSyncInterval     time.Duration
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration
	MaxCallSendMsgSize   int
	MaxCallRecvMsgSize   int
	Username             string
	Password             string
	RejectOldCluster     bool
	PermitWithoutStream  bool
	CertFile             string
	KeyFile              string
	CaCert               string
}

type clientOption struct {
	config        *Config
	context       context.Context
	dialOptions   []grpc.DialOption
	tls           *tls.Config
	confPrefix    string
	confName      string
	enableConsole bool
	enableTrace   bool
	enableMetric  bool
}

type Option func(*clientOption)

func WithContext(ctx context.Context) Option {
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

func WithDialOptions(opt grpc.DialOption) Option {
	return func(o *clientOption) {
		o.dialOptions = append(o.dialOptions, opt)
	}
}

func defaultConfig() *Config {
	return &Config{
		DialTimeout:          3 * time.Second,
		DialKeepAliveTime:    20 * time.Second,
		DialKeepAliveTimeout: 10 * time.Second,
		PermitWithoutStream:  true,
	}
}

func defaultOptions() *clientOption {
	return &clientOption{
		config:  defaultConfig(),
		context: context.Background(),
		dialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
		confPrefix:    "client.etcd",
		confName:      "default",
		enableConsole: application.GetEnableConsole(),
		enableTrace:   application.GetEnableTrace(),
		enableMetric:  application.GetEnableMetric(),
	}
}
