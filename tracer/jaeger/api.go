package jaeger

import (
	"io"

	"github.com/xslasd/goxf/log"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"
	"github.com/xslasd/goxf/conf"
)

func NewTrace(opts ...Option) (opentracing.Tracer, io.Closer, error) {
	config := DefaultConfig()
	for _, o := range opts {
		o(config)
	}
	key := config.confPrefix + "." + config.confName
	if err := conf.UnmarshalKey(key, config); err != nil {
		return nil, nil, err
	}
	if config.enableConsole {
		logger, err := log.GetJaegerLogger()
		if err != nil {
			return nil, nil, err
		}
		config.options = append(config.options, jconfig.Logger(logger))
	}
	cfg := jconfig.Configuration{
		ServiceName: config.serviceName,
		Sampler: &jconfig.SamplerConfig{
			Type:                     config.SamplerType,
			Param:                    config.SamplerParam,
			SamplingServerURL:        config.SamplingServerURL,
			SamplingRefreshInterval:  config.SamplingRefreshInterval,
			MaxOperations:            config.MaxOperations,
			OperationNameLateBinding: config.OperationNameLateBinding,
			Options:                  config.samplerOptions,
		},
		Reporter: &jconfig.ReporterConfig{
			QueueSize:                  config.QueueSize,
			LogSpans:                   config.LogSpans,
			DisableAttemptReconnecting: config.DisableAttemptReconnecting,
			AttemptReconnectInterval:   config.AttemptReconnectInterval,
			CollectorEndpoint:          config.CollectorEndpoint,
			LocalAgentHostPort:         config.Addr,
			User:                       config.User,
			Password:                   config.Password,
			HTTPHeaders:                config.reporterHTTPHeaders,
		},
		RPCMetrics: config.EnableRPCMetrics,
		Headers: &jaeger.HeadersConfig{
			JaegerDebugHeader:        config.JaegerDebugHeader,
			JaegerBaggageHeader:      config.JaegerBaggageHeader,
			TraceContextHeaderName:   config.TraceContextHeaderName,
			TraceBaggageHeaderPrefix: config.TraceBaggageHeaderPrefix,
		},
		Tags: config.tags,
	}
	return cfg.NewTracer(config.options...)
}
