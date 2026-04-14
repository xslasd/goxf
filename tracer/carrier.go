package tracer

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerlog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"
)

// MetadataReaderWriter ...
type MetadataReaderWriter struct {
	metadata.MD
}

// Set ...
func (w MetadataReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

// ForeachKey ...
func (w MetadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// MetadataInjector ...
func MetadataInjector(ctx context.Context, md metadata.MD) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return metadata.NewOutgoingContext(ctx, md)
	}
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, MetadataReaderWriter{MD: md})
	if err != nil {
		span.LogFields(tracerlog.String("event", "inject failed"), tracerlog.Error(err))
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// FromIncomingContext ... 服务器端使用
func FromIncomingContext(ctx context.Context, method, spanKind, component string) (opentracing.Span, context.Context) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, MetadataReaderWriter{MD: md})
	opts := make([]opentracing.StartSpanOption, 0)
	if err == nil {
		opts = append(opts, ext.RPCServerOption(sc))
	}
	opts = append(opts, SpanKindTag(spanKind), ComponentTag(component))
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		method,
		opts...,
	)
	return span, ctx
}

// FromOutgoingContext ...客户端使用
func FromOutgoingContext(ctx context.Context, method, spanKind, component string) (opentracing.Span, context.Context) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		method,
		SpanKindTag(spanKind),
		ComponentTag(component),
	)
	return span, MetadataInjector(ctx, md)
}
