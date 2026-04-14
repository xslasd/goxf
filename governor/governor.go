package governor

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xslasd/goxf/api"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"github.com/xslasd/goxf/server/sgin"
)

func CreateGovernor(prof, metric bool) server.Server {
	gs, err := sgin.NewGinServer(
		sgin.WithConfName("governor"),
		sgin.WithEnableMetric(false),
		sgin.WithEnableTrace(false),
		sgin.WithEnableRegister(false),
	)
	if err != nil {
		log.Panic("create governor fail", log.FieldErr(err))
	}
	if prof {
		pprof.Register(gs.Engine)
	}
	if metric {
		gs.GET("/metrics", func(c *gin.Context) {
			promhttp.Handler().ServeHTTP(c.Writer, c.Request)
		})
	}

	gs.GET("/info", func(c *gin.Context) {
		code := ecode.OK
		c.JSON(http.StatusOK, api.ResData{
			Code: code.Code(),
			Msg:  code.Message(),
			Data: info(),
		})
	})

	gs.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	gs.GET("/ecodes", func(c *gin.Context) {
		cm := ecode.GetAllCodes()
		var codes Codes
		for k, v := range cm {
			c := new(CodeInfo)
			c.Code, c.Message = k, v
			codes = append(codes, c)
		}
		sort.Sort(codes)
		code := ecode.OK
		c.JSON(http.StatusOK, api.ResData{
			Code: code.Code(),
			Msg:  code.Message(),
			Data: codes,
		})
	})

	gs.POST("/ecodes/register", func(c *gin.Context) {
		req := new(CodeMap)
		if err := c.ShouldBind(req); err != nil {
			code := ecode.RequestErr
			c.JSON(http.StatusBadRequest, api.ResData{
				Code: code.Code(),
				Msg:  code.Message(),
			})
			return
		}
		rm := make(map[int]string)
		for _, c := range req.Codes {
			rm[c.Code] = c.Message
		}
		ecode.Register(rm)
		code := ecode.OK
		c.JSON(http.StatusOK, api.ResData{
			Code: code.Code(),
			Msg:  code.Message(),
		})
	})
	return gs
}

func info() *APPInfo {
	app := new(APPInfo)
	app.APPID = application.GetAppId()
	app.ServiceName = application.GetServiceName()
	app.Version = application.GetVersion()
	app.BuildUser = application.GetBuildUser()

	if application.GetBuildTime() != "" {
		time, err := strconv.ParseInt(application.GetBuildTime(), 10, 64)
		if err == nil {
			app.BuildTime = time
		}
	}

	app.StartTime = application.GetStartTime().UnixMilli()
	return app
}

type Codes []*CodeInfo

func (c Codes) Len() int           { return len(c) }
func (c Codes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Codes) Less(i, j int) bool { return c[i].Code < c[j].Code }
