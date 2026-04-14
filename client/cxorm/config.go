package cxorm

import "github.com/xslasd/goxf/application"

type Config struct {
	Driver      string
	DataSources []string
	IsShowSQL   bool
}

type clientOption struct {
	config       *Config
	confPrefix   string
	confName     string
	isShowSQL    bool
	enableTrace  bool
	enableMetric bool
}

type Option func(*clientOption)

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
		Driver: "mysql",
	}
}

func defaultOptions() *clientOption {
	return &clientOption{
		config:       defaultConfig(),
		confPrefix:   "client.xorm",
		confName:     "default",
		enableTrace:  application.GetEnableTrace(),
		enableMetric: application.GetEnableMetric(),
	}
}
