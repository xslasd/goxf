package jaeger

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"
	"github.com/xslasd/goxf/application"
)

type Config struct {
	SamplerType              string
	SamplerParam             float64
	SamplingServerURL        string
	SamplingRefreshInterval  time.Duration
	MaxOperations            int
	OperationNameLateBinding bool

	QueueSize                  int
	BufferFlushInterval        time.Duration
	LogSpans                   bool
	DisableAttemptReconnecting bool
	AttemptReconnectInterval   time.Duration
	CollectorEndpoint          string
	Addr                       string
	User                       string
	Password                   string

	JaegerDebugHeader        string
	JaegerBaggageHeader      string
	TraceContextHeaderName   string
	TraceBaggageHeaderPrefix string

	EnableRPCMetrics bool

	reporterHTTPHeaders map[string]string
	samplerOptions      []jaeger.SamplerOption
	serviceName         string
	tags                []opentracing.Tag
	options             []jconfig.Option
	confPrefix          string
	confName            string
	enableConsole       bool
}

type Option func(*Config)

func WithSamplerOptions(opt jaeger.SamplerOption) Option {
	return func(o *Config) {
		o.samplerOptions = append(o.samplerOptions, opt)
	}
}

func WithReporterHTTPHeaders(header map[string]string) Option {
	return func(o *Config) {
		o.reporterHTTPHeaders = header
	}
}

func WithServiceName(serviceName string) Option {
	return func(o *Config) {
		o.serviceName = serviceName
	}
}

func WithConfPrefix(prefix string) Option {
	return func(o *Config) {
		o.confPrefix = prefix
	}
}
func WithConfName(name string) Option {
	return func(o *Config) {
		o.confName = name
	}
}

func WithTags(tag opentracing.Tag) Option {
	return func(o *Config) {
		o.tags = append(o.tags, tag)
	}
}

func WithOptions(opt jconfig.Option) Option {
	return func(o *Config) {
		o.options = append(o.options, opt)
	}
}

func WithEnableConsole(enableConsole bool) Option {
	return func(o *Config) {
		o.enableConsole = enableConsole
	}
}

func DefaultConfig() *Config {
	return &Config{
		SamplerType:              "const",
		SamplerParam:             1,
		LogSpans:                 true,
		EnableRPCMetrics:         true,
		TraceBaggageHeaderPrefix: "ctx-",
		TraceContextHeaderName:   application.TraceContextHeaderName,
		samplerOptions:           make([]jaeger.SamplerOption, 0),
		tags:                     make([]opentracing.Tag, 0),
		options:                  make([]jconfig.Option, 0),
		confPrefix:               "tracer.jaeger",
		confName:                 "default",
		enableConsole:            application.GetEnableConsole(),
		serviceName:              application.GetServiceName(),
	}
}
