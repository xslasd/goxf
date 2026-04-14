package queue

import "github.com/xslasd/goxf/application"

type Config struct {
	QueueName string
	WorkerNum int
}

type options struct {
	config       *Config
	confPrefix   string
	confName     string
	enableMetric bool
}

type Option func(*options)

func WithEnableMetric(enable bool) Option {
	return func(o *options) {
		o.enableMetric = enable
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

func WithWorkerNum(num int) Option {
	return func(o *options) {
		o.config.WorkerNum = num
	}
}

func defaultConfig() *Config {
	return &Config{
		QueueName: "default",
		WorkerNum: 1,
	}
}

func defaultOptions() *options {
	return &options{
		config:       defaultConfig(),
		confPrefix:   "job.queue",
		confName:     "default",
		enableMetric: application.GetEnableMetric(),
	}
}
