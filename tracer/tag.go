package tracer

import "github.com/opentracing/opentracing-go"

// CustomTag ...
func CustomTag(key string, val any) opentracing.Tag {
	return opentracing.Tag{
		Key:   key,
		Value: val,
	}
}

// TagComponent ...
func ComponentTag(component string) opentracing.Tag {
	return opentracing.Tag{
		Key:   "component",
		Value: component,
	}
}

// TagSpanKind ...
func SpanKindTag(kind string) opentracing.Tag {
	return opentracing.Tag{
		Key:   "span.kind",
		Value: kind,
	}
}

// TagSpanURL ...
func SpanURLTag(url string) opentracing.Tag {
	return opentracing.Tag{
		Key:   "span.url",
		Value: url,
	}
}
