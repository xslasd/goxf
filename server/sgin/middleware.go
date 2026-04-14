package sgin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/xslasd/goxf/api"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/i18n"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/metric"
	"github.com/xslasd/goxf/tracer"
)

func getRoute(c *gin.Context, routeFn UseRouteFunc) string {
	if routeFn != nil {
		if nr := routeFn(c); nr != "" {
			return nr
		}
	}
	return c.Request.Method + ":" + c.Request.URL.Path
}

func traceServerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span, ctx := tracer.FromIncomingContext(
			c.Request.Context(),
			c.Request.Method+" "+c.Request.URL.Path,
			"RestFul",
			"http",
		)
		c.Request = c.Request.WithContext(ctx)
		defer span.Finish()
		c.Next()
	}
}

func metricServerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		c.Next()
		metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.HTTPType, c.Request.Method+"."+c.Request.URL.Path)
		metric.ServerHandleCounter.Inc(metric.HTTPType, c.Request.Method+"."+c.Request.URL.Path, http.StatusText(c.Writer.Status()))
		return
	}
}

func defaultMiddleware(name string, slowQueryThresholdInMilli int64, routeFn UseRouteFunc, timeoutEvent TimeoutEventFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		defer func() {
			du := time.Since(beg)
			route := getRoute(c, routeFn)

			isSlow := false
			event := "normal"
			if slowQueryThresholdInMilli > 0 && du > time.Duration(slowQueryThresholdInMilli)*time.Millisecond {
				event = "slow"
				isSlow = true
			}

			log.Info("access",
				log.String("name", name),
				log.String("route", route),
				log.FieldCost(du),
				log.FieldEvent(event),
			)

			if isSlow && timeoutEvent != nil {
				timeoutEvent(c, route, du)
			}
		}()
		c.Next()
	}
}

func debugMiddleware(name string, slowQueryThresholdInMilli int64, routeFn UseRouteFunc, timeoutEvent TimeoutEventFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		defer func() {
			du := time.Since(beg)
			route := getRoute(c, routeFn)

			isSlow := false
			event := color.GreenString("normal")
			if slowQueryThresholdInMilli > 0 && du > time.Duration(slowQueryThresholdInMilli)*time.Millisecond {
				event = color.RedString("slow")
				isSlow = true
			}

			log.Debugf("%-10s: %s %-20s --> %-60s %-18v [%s]",
				name,
				color.GreenString(strconv.Itoa(c.Writer.Status())),
				c.ClientIP(),
				color.BlueString(route),
				du,
				event,
			)

			if isSlow && timeoutEvent != nil {
				timeoutEvent(c, route, du)
			}
		}()
		c.Next()
	}
}

type JWTVerifyFunc func(c *gin.Context, token string) error

func DefaultJSONRes(c *gin.Context, data any, err error, options ...api.ResOption) {
	c.JSON(api.BaseRes(data, err, options...))
}

func DefaultJSONWithI18nRes(c *gin.Context, data any, err error) {
	options := []api.ResOption{
		api.WithI18nHandler(I18nECodeFunc(c)),
	}
	DefaultJSONRes(c, data, err, options...)
}

// I18nKey gin.Context中存储 i18n.ILanguage 实例的键
var I18nKey = "i18n"

// I18nECodePrefixKey i18n.ILanguage.T 调用时的键前缀
var I18nECodePrefixKey = "ecode."

// I18nECodeFunc 从 gin.Context 中获取 i18n.ILanguage 实例，并返回 i18n.ECode 转换函数
func I18nECodeFunc(c *gin.Context) api.I18nECodeHandler {
	language, ok := c.Value(I18nKey).(i18n.ILanguage)
	return func(code ecode.ECodes) string {
		var message string
		if ok {
			vals := code.Values()
			key := I18nECodePrefixKey + strconv.Itoa(code.Code())

			if len(vals) > 0 {
				message = language.Tf(key, vals)
			} else {
				message = language.T(key)
			}
		}
		if message != "" {
			return message
		}
		return code.Message()
	}
}

// TokenHeaderKey token 请求头中的键名
var TokenHeaderKey = "Authorization"

// TokenQueryKey token 查询参数的键名
var TokenQueryKey = "token"

func JWTAuthMiddleware(verifyFunc JWTVerifyFunc, options ...api.ResOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(TokenHeaderKey)

		if token != "" {
			if strings.HasPrefix(token, "Bearer ") {
				token = token[7:]
			} else if idx := strings.IndexByte(token, ' '); idx != -1 {
				// 兼容自定义前缀的 Authorization (如: "Token xxx")
				token = token[idx+1:]
			}
		}

		// 无 Authorization 头部时，降级尝试从 Query 或 Form 提取
		if token == "" {
			if token = c.Query(TokenQueryKey); token == "" {
				token = c.PostForm(TokenQueryKey)
			}
		}

		if token == "" || verifyFunc == nil {
			c.JSON(api.BaseRes(nil, ecode.Unauthorized, options...))
			c.Abort()
			return
		}

		if err := verifyFunc(c, token); err != nil {
			c.JSON(api.BaseRes(nil, err, options...))
			c.Abort()
			return
		}
		c.Next()
	}
}
