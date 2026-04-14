package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/xslasd/goxf/application"
)

type Config struct {
	Spec string
}

type options struct {
	config       *Config
	confPrefix   string
	confName     string
	cronOpts     []cron.Option
	enableMetric bool
}

type Option func(*options)

func WithCronOption(opts ...cron.Option) Option {
	return func(o *options) {
		o.cronOpts = append(o.cronOpts, opts...)
	}
}

func WithConfPrefix(prefix string) Option {
	return func(o *options) {
		o.confPrefix = prefix
	}
}

func WithConfName(name string) Option {
	return func(o *options) {
		o.confName = name
	}
}

func WithSpec(spec string) Option {
	return func(o *options) {
		o.config.Spec = spec
	}
}

func WithEnableMetric(enableMetric bool) Option {
	return func(o *options) {
		o.enableMetric = enableMetric
	}
}

func defaultConfig() *Config {
	return &Config{}
}

func defaultOptions() *options {
	return &options{
		config:       defaultConfig(),
		confPrefix:   "job.cron",
		confName:     "default",
		enableMetric: application.GetEnableMetric(),
	}
}
