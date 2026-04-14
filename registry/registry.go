package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xslasd/goxf/server"
)

type Registry interface {
	Register(ctx context.Context, srv *server.ServiceInfo) error
	UnRegister(ctx context.Context, srv *server.ServiceInfo) error
	ListServices(ctx context.Context, name string) ([]*server.ServiceInfo, error)
	Watch(ctx context.Context, name string) (chan *Endpoints, error)
	Kind() string
}

// GetServiceKey ..
func GetServiceKey(prefix string, s *server.ServiceInfo) string {
	key := fmt.Sprintf("/%s/%s/%s://%s", prefix, s.Name, s.Scheme, s.Address)
	return strings.ToLower(key)
}

// GetServiceValue ..
func GetServiceValue(s *server.ServiceInfo) string {
	val, _ := json.Marshal(s)
	return string(val)
}
