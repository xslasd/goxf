package cgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/ecode"
	pb "github.com/xslasd/goxf/ecode/grpc/proto"
	"github.com/xslasd/goxf/log"

	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/registry/balancer"
	"google.golang.org/grpc"
)

func NewClient(opts ...Option) (*grpc.ClientConn, error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		return nil, err
	}

	opt.config.CallTimeout = time.Second * opt.config.CallTimeout
	opt.config.ReadTimeout = time.Second * opt.config.ReadTimeout
	opt.config.SlowThreshold = time.Millisecond * opt.config.SlowThreshold
	opt.config.DialKeepAliveTime = time.Second * opt.config.DialKeepAliveTime
	opt.config.DialKeepAliveTimeout = time.Second * opt.config.DialKeepAliveTimeout

	if opt.config.Balancer == "" {
		opt.config.Balancer = balancer.NameSmoothWeightRoundRobin
	}
	b := fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, opt.config.Balancer)
	opt.dialOptions = append(opt.dialOptions,
		grpc.WithDefaultServiceConfig(b),
		grpc.WithChainUnaryInterceptor(callTimeOutClientInterceptor(opt.config.CallTimeout)),
	)

	if opt.enableConsole {
		opt.dialOptions = append(opt.dialOptions,
			grpc.WithChainUnaryInterceptor(debugUnaryClientInterceptor(opt.config.Address)),
			grpc.WithStreamInterceptor(debugStreamClientInterceptor(opt.config.Address)),
		)
	} else {
		opt.dialOptions = append(opt.dialOptions,
			grpc.WithChainUnaryInterceptor(loggerUnaryClientInterceptor(opt.config.Name, opt.config.SlowThreshold)),
		)
	}
	if opt.enableMetric {
		opt.dialOptions = append(opt.dialOptions,
			grpc.WithChainUnaryInterceptor(metricUnaryClientInterceptor(opt.config.Name)),
			grpc.WithStreamInterceptor(metricStreamClientInterceptor(opt.config.Name)),
		)
	}
	if opt.enableTrace {
		opt.dialOptions = append(opt.dialOptions,
			grpc.WithChainUnaryInterceptor(traceUnaryClientInterceptor()),
			grpc.WithStreamInterceptor(traceStreamClientInterceptor()),
		)
	}
	log.Infof("cgRPC.%s ：new gRPC client[%s] dialing...", opt.config.Name, opt.config.Address)
	cc, err := newGRPCClient(opt.context, opt.config, opt.dialOptions...)
	if err != nil {
		return nil, err
	}

	log.Infof("cgRPC.%s ：new gRPC client[%s] dial ok.", opt.config.Name, opt.config.Address)
	if opt.syncECode {
		cli := pb.NewCodesClient(cc)
		res, err := cli.GetCodes(context.TODO(), &pb.Empty{})
		if err != nil {
			log.Error("GetCodes failed", log.FieldErr(err))
		}
		if res != nil {
			cm := make(map[int]string)
			for _, item := range res.Rows {
				cm[int(item.Code)] = item.Message
			}
			ecode.Register(cm)
		}
	}

	return cc, nil
}
