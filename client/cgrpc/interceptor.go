package cgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/opentracing/opentracing-go/ext"
	tracerlog "github.com/opentracing/opentracing-go/log"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
	"github.com/xslasd/goxf/tracer"
	"github.com/xslasd/goxf/utils/xstring"
	"google.golang.org/grpc"
)

func callTimeOutClientInterceptor(callTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			ctx, cancel = context.WithTimeout(ctx, callTimeout)
			defer cancel()
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}

// @Description: metric统计-RPCUnary
// @param name 服务名称
// @return grpc.UnaryClientInterceptor
func metricUnaryClientInterceptor(name string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		_, code := ecode.FromGRPCError(err)
		peer := cc.Target()
		metric.ClientHandleCounter.Inc(metric.GRPCUnaryType, name, method, peer, code.Error())
		metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.GRPCUnaryType, name, method, peer)
		return err
	}
}

// @Description: metric统计-GRPCStream
// @param name 服务名称
// @return grpc.StreamClientInterceptor
func metricStreamClientInterceptor(name string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		beg := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		_, code := ecode.FromGRPCError(err)
		srv := cc.Target()
		metric.ClientHandleCounter.Inc(metric.GRPCStreamType, name, method, srv, code.Error())
		metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.GRPCStreamType, name, method, srv)
		return clientStream, err
	}
}

// traceUnaryClientInterceptor ...
func traceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, ctx := tracer.FromOutgoingContext(ctx, method, "client", "grpc")
		defer span.Finish()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			s, c := ecode.FromGRPCError(err)
			span.SetTag("response_status", s)
			span.SetTag("response_code", c.Code())
			ext.Error.Set(span, true)
			span.LogFields(tracerlog.String("event", "error"), tracerlog.String("message", c.Message()))
		}
		return err
	}
}

// traceStreamClientInterceptor ...
func traceStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		span, ctx := tracer.FromOutgoingContext(ctx, method, "client", "grpc")
		defer span.Finish()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			s, c := ecode.FromGRPCError(err)
			span.SetTag("response_status", s)
			span.SetTag("response_code", c.Code())
			ext.Error.Set(span, true)
			span.LogFields(tracerlog.String("event", "error"), tracerlog.String("message", c.Message()))
		}
		return clientStream, err
	}
}

// debugUnaryClientInterceptor ...
func debugUnaryClientInterceptor(addr string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		prefix := fmt.Sprintf("%s-->%s", addr, method)
		log.Debugf("%s => %s\n", color.GreenString(prefix), color.GreenString("req: "+xstring.Json(req)))
		err := invoker(ctx, method, req, reply, cc, opts...)
		prefix2 := fmt.Sprintf("%s<--%s", addr, method)
		if err != nil {
			log.Debugf("%s => %s\n", color.RedString(prefix2), color.RedString("Err: "+err.Error()))
		} else {
			log.Debugf("%s => %s\n", color.GreenString(prefix2), color.GreenString("reply: "+xstring.Json(reply)))
		}
		return err
	}
}
func debugStreamClientInterceptor(addr string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		prefix := fmt.Sprintf("%s-->%s", addr, method)
		log.Debugf("%s => %s\n", color.GreenString(prefix), color.GreenString("req: grpc.Streamer"))
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		prefix2 := fmt.Sprintf("%s<--%s", addr, method)
		if err != nil {
			log.Debugf("%s => %s\n", color.RedString(prefix2), color.RedString("Err: "+err.Error()))
		} else {
			log.Debugf("%s => %s\n", color.GreenString(prefix2), color.GreenString("reply: ...."))
		}
		return clientStream, err
	}
}

// loggerUnaryClientInterceptor gRPC客户端日志中间件
func loggerUnaryClientInterceptor(name string, slowThreshold time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		du := time.Since(beg)
		srv := cc.Target()
		event := "normal"
		if slowThreshold > 0 && du > slowThreshold {
			event = "slow"
		}
		if err != nil {
			// 只记录系统级别错误
			log.Error(
				"access",
				log.FieldName(name),
				log.FieldMethod(method),
				log.FieldCost(du),
				log.FieldAddr(srv),
				log.FieldEvent(event),
				log.Any("req", xstring.Json(req)),
				log.FieldErr(err),
			)
		} else {
			log.Info(
				"access",
				log.FieldName(name),
				log.FieldMethod(method),
				log.FieldCost(du),
				log.FieldAddr(srv),
				log.FieldEvent(event),
				log.Any("req", xstring.Json(req)),
				log.Any("reply", xstring.Json(reply)),
			)
		}
		return err
	}
}
