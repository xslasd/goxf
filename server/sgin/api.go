package sgin

import (
	"net/http"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"github.com/xslasd/goxf/server/cors"
)

func NewGinServer(opts ...Option) (*Server, error) {
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

	if opt.enableConsole {
		opt.middleware = append(opt.middleware, debugMiddleware(opt.confName, opt.config.SlowQueryThresholdInMilli, opt.setRouteFn, opt.timeoutEvent))
	} else {
		opt.middleware = append(opt.middleware, defaultMiddleware(opt.confName, opt.config.SlowQueryThresholdInMilli, opt.setRouteFn, opt.timeoutEvent))
	}

	if opt.enableTrace {
		opt.middleware = append(opt.middleware, traceServerMiddleware())
	}

	if opt.enableMetric {
		opt.middleware = append(opt.middleware, metricServerMiddleware())
	}
	if len(opt.config.AllowedOrigins) > 0 {
		opt.corsOptions.AllowedOrigins = opt.config.AllowedOrigins
	}
	opt.config.Addr = server.BuildAddress(opt.config.Addr)
	gin.SetMode(opt.mode)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debugf("%-15s %-60s --> %s (%d handlers)", color.BlueString(httpMethod), color.YellowString(absolutePath), color.GreenString(handlerName), nuHandlers)
	}
	engine := gin.New()
	corsHandel := cors.New(opt.corsOptions)
	engine.Use(func(ctx *gin.Context) {
		corsHandel.HandlerFunc(ctx.Writer, ctx.Request)
	})
	engine.Use(opt.middleware...)
	return &Server{
		Engine: engine,
		opts:   opt,
		server: &http.Server{
			Addr:    opt.config.Addr,
			Handler: engine,
		},
	}, nil
}
