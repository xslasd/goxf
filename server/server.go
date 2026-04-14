package server

import (
	"context"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/utils/xnet"
	"strings"
)

type Server interface {
	IsRegister() bool
	Init() error
	Serve() error
	Stop(ctx context.Context) error
	Info() *ServiceInfo
}

type ServiceInfo struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Scheme   string            `json:"scheme"`
	Metadata map[string]string `json:"metadata"`
	Weight   int               `json:"weight"`
}

func BuildAddress(addr string) string {
	ad := strings.Split(addr, ":")
	if ad[0] != "" {
		return addr
	}
	ip, err := xnet.GetOutboundIP()
	if err != nil {
		log.Error("get Outbound IP error", log.FieldErr(err))
		return addr
	}
	ad[0] = ip.String()
	return strings.Join(ad, ":")
}
