package sgin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
)

// Server ...
type Server struct {
	*gin.Engine
	server *http.Server
	opts   *srvOption
}

func (s *Server) Init() error {
	if s.opts.enableConsole {
		for _, route := range s.Engine.Routes() {
			str := fmt.Sprintf("%s %s", color.BlueString(route.Method), color.YellowString(route.Path))
			log.Debugf("gin.%s: %-30s --> %s", s.opts.confName, str, color.GreenString(route.Handler))
		}
		log.Debugf("gin.%s: cors option allowedOrigins[%s]", s.opts.confName, strings.Join(s.opts.corsOptions.AllowedOrigins, ","))
		log.Debugf("gin.%s: cors option allowedMethods[%s]", s.opts.confName, strings.Join(s.opts.corsOptions.AllowedMethods, ","))
		log.Debugf("gin.%s: cors option allowedHeaders[%s]", s.opts.confName, strings.Join(s.opts.corsOptions.AllowedHeaders, ","))
		log.Debugf("gin.%s: cors option allowCredentials[%t]", s.opts.confName, s.opts.corsOptions.AllowCredentials)
	}
	return nil
}

func (s *Server) IsRegister() bool {
	return s.opts.enableRegister
}

// Serve implements server.Server interface.
func (s *Server) Serve() error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Warnf("gin.%s: %s closed.", s.opts.confName, s.opts.config.Addr)
		return nil
	}
	return err
}

// Stop implements server.Server interface
// it will stop gin server gracefully
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Info() *server.ServiceInfo {
	return &server.ServiceInfo{
		Name:    s.opts.confName,
		Address: s.opts.config.Addr,
		Scheme:  "gin",
	}
}
