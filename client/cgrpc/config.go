package cgrpc

import (
	"context"
	"time"

	"github.com/xslasd/goxf/application"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xslasd/goxf/registry/balancer"
	"google.golang.org/grpc"
)

type Config struct {
	Name                 string
	Address              string
	CallTimeout          time.Duration //连接超时
	ReadTimeout          time.Duration
	SlowThreshold        time.Duration //慢阈值
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration
	PermitWithoutStream  bool
	Balancer             string
}

type clientOption struct {
	config        *Config
	context       context.Context
	dialOptions   []grpc.DialOption
	confPrefix    string
	confName      string
	enableConsole bool
	enableTrace   bool
	enableMetric  bool
	syncECode     bool
}

type Option func(*clientOption)

func WithSyncECode(syncECode bool) Option {
	return func(o *clientOption) {
		o.syncECode = syncECode
	}
}

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
		CallTimeout:          60,
		ReadTimeout:          60,
		SlowThreshold:        600,
		DialKeepAliveTime:    30,
		DialKeepAliveTimeout: 20,
		PermitWithoutStream:  true,
		Balancer:             balancer.NameSmoothWeightRoundRobin,
	}
}

func defaultOptions() *clientOption {
	return &clientOption{
		config:  defaultConfig(),
		context: context.Background(),
		dialOptions: []grpc.DialOption{
			//禁用 TLS 连接
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			//grpc.WithBlock(),
		},
		confPrefix:    "client.grpc",
		confName:      "default",
		enableConsole: application.GetEnableConsole(),
		enableTrace:   application.GetEnableTrace(),
		enableMetric:  application.GetEnableMetric(),
		syncECode:     false,
	}
}
