package sgrpc

import (
	"github.com/xslasd/goxf/application"
	"google.golang.org/grpc"
)

type Config struct {
	Name                      string
	Addr                      string
	Network                   string
	Weight                    int
	SlowQueryThresholdInMilli int64
}

type srvOption struct {
	config             *Config
	enableConsole      bool
	enableRegister     bool
	enableTrace        bool
	enableMetric       bool
	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
	confPrefix         string
	confName           string
}

type Option func(*srvOption)

func WithConfPrefix(prefix string) Option {
	return func(s *srvOption) {
		s.confPrefix = prefix
	}
}

func WithConfName(name string) Option {
	return func(s *srvOption) {
		s.confName = name
	}
}

func WithGRPCServerOption(opt grpc.ServerOption) Option {
	return func(s *srvOption) {
		s.serverOptions = append(s.serverOptions, opt)
	}
}

func WithStreamInterceptor(opt grpc.StreamServerInterceptor) Option {
	return func(c *srvOption) {
		c.streamInterceptors = append(c.streamInterceptors, opt)
	}
}

func WithUnaryInterceptor(opt grpc.UnaryServerInterceptor) Option {
	return func(c *srvOption) {
		c.unaryInterceptors = append(c.unaryInterceptors, opt)
	}
}

func WithEnableRegister(enableRegister bool) Option {
	return func(o *srvOption) {
		o.enableRegister = enableRegister
	}
}

// defaultConfig represents a default config which can be set in config file.
// User can get a default config through this func and redefine it,
// then use WithConfig func to pass it to serverOption.
func defaultConfig() *Config {
	return &Config{
		Addr:                      ":9092",
		Network:                   "tcp4",
		SlowQueryThresholdInMilli: 500,
	}
}

// defaultServerOption represents default config option of grpc server
// User should construct config base on With series functions.
func defaultServerOption() *srvOption {
	return &srvOption{
		config:             defaultConfig(),
		enableConsole:      application.GetEnableConsole(),
		enableRegister:     application.GetEnableRegister(),
		enableTrace:        application.GetEnableTrace(),
		enableMetric:       application.GetEnableMetric(),
		serverOptions:      []grpc.ServerOption{},
		streamInterceptors: []grpc.StreamServerInterceptor{},
		unaryInterceptors:  []grpc.UnaryServerInterceptor{},
		confPrefix:         "server.grpc",
		confName:           "default",
	}
}
