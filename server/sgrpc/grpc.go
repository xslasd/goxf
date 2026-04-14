package sgrpc

import (
	"context"
	"net"

	ecodesrv "github.com/xslasd/goxf/ecode/grpc"
	pb "github.com/xslasd/goxf/ecode/grpc/proto"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	listener net.Listener
	opts     *srvOption
}

func (s *Server) Init() error {
	return nil
}
func (s *Server) IsRegister() bool {
	return s.opts.enableRegister
}

func (s *Server) Serve() error {
	// ecode grpc api
	pb.RegisterCodesServer(s.Server, &ecodesrv.Codes{})
	// The grpc process will block until done.
	err := s.Server.Serve(s.listener)
	return err
}

func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	log.Warnf("grpc.%s: %s closed.", s.opts.config.Name, s.opts.config.Addr)
	return nil
}

func (s *Server) Info() *server.ServiceInfo {
	return &server.ServiceInfo{
		Name:    s.opts.config.Name,
		Address: s.opts.config.Addr,
		Scheme:  "grpc",
		Weight:  s.opts.config.Weight,
	}
}
