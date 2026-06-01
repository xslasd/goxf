package sgin

import (
	"time"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/server/cors"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Addr                      string
	SlowQueryThresholdInMilli int64
	AllowedOrigins            []string
}

type srvOption struct {
	config         *Config
	enableConsole  bool
	enableRegister bool
	enableTrace    bool
	enableMetric   bool
	middleware     []gin.HandlerFunc
	confPrefix     string
	confName       string
	mode           string
	printRoute     bool

	//设置自定义路由
	setRouteFn UseRouteFunc
	//设置超时事件
	timeoutEvent TimeoutEventFunc

	corsOptions cors.Options
}

type UseRouteFunc func(c *gin.Context) string
type TimeoutEventFunc func(c *gin.Context, route string, cost time.Duration)

func WithUseTimeRoute(routeFn UseRouteFunc) Option {
	return func(o *srvOption) {
		o.setRouteFn = routeFn
	}
}
func WithTimeoutEvent(timeoutEvent TimeoutEventFunc) Option {
	return func(o *srvOption) {
		o.timeoutEvent = timeoutEvent
	}
}

type Option func(*srvOption)

func WithEnableMetric(enableMetric bool) Option {
	return func(o *srvOption) {
		o.enableMetric = enableMetric
	}
}

func WithEnableConsole(enableConsole bool) Option {
	return func(o *srvOption) {
		o.enableConsole = enableConsole
	}
}

func WithEnableTrace(enableTrace bool) Option {
	return func(o *srvOption) {
		o.enableTrace = enableTrace
	}
}

func WithEnableRegister(enableRegister bool) Option {
	return func(o *srvOption) {
		o.enableRegister = enableRegister
	}
}

func WithConfPrefix(prefix string) Option {
	return func(s *srvOption) {
		s.confPrefix = prefix
	}
}

func WithConfName(name string) Option {
	return func(s *srvOption) {
		s.confName = name
	}
}

func WithMiddleware(handle gin.HandlerFunc) Option {
	return func(s *srvOption) {
		s.middleware = append(s.middleware, handle)
	}
}
func WithCorsOptions(option cors.Options) Option {
	return func(s *srvOption) {
		s.corsOptions = option
	}
}
func WithPrintRoute(isPrint bool) Option {
	return func(o *srvOption) {
		o.printRoute = isPrint
	}
}

// DefaultConfig ...
func defaultConfig() *Config {
	return &Config{
		Addr:                      "0.0.0.0:8080",
		SlowQueryThresholdInMilli: 500, // 500ms
	}
}

func defaultServerOption() *srvOption {
	return &srvOption{
		config:         defaultConfig(),
		enableConsole:  application.GetEnableConsole(),
		enableTrace:    application.GetEnableTrace(),
		enableMetric:   application.GetEnableMetric(),
		enableRegister: application.GetEnableRegister(),
		confPrefix:     "server.gin",
		confName:       "default",
		middleware:     []gin.HandlerFunc{},
		mode:           gin.ReleaseMode,
		printRoute:     true,
		corsOptions:    cors.DefaultOptions(),
	}
}
