package sgrpc

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	tracerlog "github.com/opentracing/opentracing-go/log"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
	"github.com/xslasd/goxf/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func prometheusUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	startTime := time.Now()
	resp, err := handler(ctx, req)
	_, code := ecode.FromGRPCError(err)
	metric.ServerHandleHistogram.Observe(time.Since(startTime).Seconds(), metric.GRPCUnaryType, info.FullMethod)
	metric.ServerHandleCounter.Inc(metric.GRPCUnaryType, info.FullMethod, code.Error())
	return resp, err
}

func prometheusStreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	startTime := time.Now()
	err := handler(srv, ss)
	_, code := ecode.FromGRPCError(err)
	metric.ServerHandleHistogram.Observe(time.Since(startTime).Seconds(), metric.GRPCStreamType, info.FullMethod)
	metric.ServerHandleCounter.Inc(metric.GRPCStreamType, info.FullMethod, code.Error())
	return err
}

func traceUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	span, ctx := tracer.FromIncomingContext(ctx, info.FullMethod, "gRPC", "server.unary")
	defer span.Finish()

	resp, err := handler(ctx, req)
	if err != nil {
		s, c := ecode.FromGRPCError(err)
		span.SetTag("response_status", s)
		span.SetTag("response_code", c.Code())
		ext.Error.Set(span, true)
		span.LogFields(tracerlog.String("event", "error"), tracerlog.String("message", c.Message()))
	}
	return resp, err
}

type contextedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context ...
func (css contextedServerStream) Context() context.Context {
	return css.ctx
}

func traceStreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	span, ctx := tracer.FromIncomingContext(ss.Context(), info.FullMethod, "server.stream", "gRPC")
	defer span.Finish()
	return handler(srv, contextedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	})
}

func defaultUnaryServerInterceptor(name string, slowQueryThresholdInMilli int64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var beg = time.Now()
		defer func() {
			logFields := []log.Field{
				log.String("name", name),
				log.Any("gRPC interceptor type", "unary"),
				log.FieldMethod(info.FullMethod),
			}
			err = defaultInterceptorHandler(ctx, beg, slowQueryThresholdInMilli, logFields)
		}()
		return handler(ctx, req)
	}
}

func defaultStreamServerInterceptor(name string, slowQueryThresholdInMilli int64) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var beg = time.Now()
		defer func() {
			logFields := []log.Field{
				log.String("name", name),
				log.Any("gRPC interceptor type", "stream"),
				log.FieldMethod(info.FullMethod),
			}
			err = defaultInterceptorHandler(stream.Context(), beg, slowQueryThresholdInMilli, logFields)
		}()
		return handler(srv, stream)
	}
}

func defaultInterceptorHandler(ctx context.Context, beg time.Time, slowQueryThresholdInMilli int64, logFields []log.Field) error {
	var err error
	var event = "normal"
	var fields = make([]log.Field, 0)
	du := time.Since(beg)
	if slowQueryThresholdInMilli > 0 {
		if du > time.Duration(slowQueryThresholdInMilli)*time.Millisecond {
			event = "slow"
		}
	}
	if logFields != nil {
		fields = append(fields, logFields...)
	}
	fields = append(fields,
		log.FieldCost(du),
		log.FieldEvent(event),
	)
	for key, val := range getPeer(ctx) {
		fields = append(fields, log.Any(key, val))
	}
	if rec := recover(); rec != nil {
		switch recT := rec.(type) {
		case error:
			err = recT
		default:
			err = fmt.Errorf("%v", recT)
		}
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, true)]
		fields = append(fields, log.FieldErr(err))
		fields = append(fields, log.FieldStack(stack))
		log.Error("access", fields...)
		return err
	}
	log.Info("access", fields...)
	return err
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClientIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

func getPeer(ctx context.Context) map[string]string {
	var peerMeta = make(map[string]string)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val, ok := md["aid"]; ok {
			peerMeta["aid"] = strings.Join(val, ";")
		}
		var clientIP string
		if val, ok := md["client-ip"]; ok {
			clientIP = strings.Join(val, ";")
		} else {
			ip, err := getClientIP(ctx)
			if err == nil {
				clientIP = ip
			}
		}
		peerMeta["clientIP"] = clientIP
		if val, ok := md["client-host"]; ok {
			peerMeta["host"] = strings.Join(val, ";")
		}
	}
	return peerMeta
}
