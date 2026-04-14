package sgrpc

import (
	"net"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"google.golang.org/grpc"
)

// NewGRPCServer return a Server pointer according to the Options,
// note that the priority is: With series functions > config file > default
func NewGRPCServer(opts ...Option) (*Server, error) {
	application.CheckStartupGoxf()
	opt := defaultServerOption()
	for _, o := range opts {
		o(opt)
	}

	// get config content with default config prefix and name when opt.config is nil
	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		log.Panic("unmarshal config err", log.FieldErr(err), log.FieldKey(key), log.FieldValueAny(opt.config))
	}

	opt.config.Addr = server.BuildAddress(opt.config.Addr)
	// 增加拦截器选项
	opt.unaryInterceptors = append(opt.unaryInterceptors, defaultUnaryServerInterceptor(opt.config.Name, opt.config.SlowQueryThresholdInMilli))
	opt.streamInterceptors = append(opt.streamInterceptors, defaultStreamServerInterceptor(opt.config.Name, opt.config.SlowQueryThresholdInMilli))

	if opt.enableMetric {
		opt.unaryInterceptors = append(opt.unaryInterceptors, prometheusUnaryServerInterceptor)
		opt.streamInterceptors = append(opt.streamInterceptors, prometheusStreamServerInterceptor)
	}
	if opt.enableTrace {
		opt.unaryInterceptors = append(opt.unaryInterceptors, traceUnaryServerInterceptor)
		opt.streamInterceptors = append(opt.streamInterceptors, traceStreamServerInterceptor)
	}
	opt.serverOptions = append(opt.serverOptions,
		grpc.ChainUnaryInterceptor(opt.unaryInterceptors...),
		grpc.ChainStreamInterceptor(opt.streamInterceptors...),
	)
	srv := grpc.NewServer(opt.serverOptions...)
	listener, err := net.Listen(opt.config.Network, opt.config.Addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		Server:   srv,
		listener: listener,
		opts:     opt,
	}, nil
}
