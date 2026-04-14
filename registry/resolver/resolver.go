package resolver

import (
	"context"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func RegisterBuilder(scheme string, reg registry.Registry) {
	resolver.Register(&EtcdBuilder{
		scheme: scheme,
		reg:    reg,
	})
}

type EtcdBuilder struct {
	scheme string
	reg    registry.Registry
}

func (e EtcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	endpoints, err := e.reg.Watch(context.Background(), target.URL.Path)
	if err != nil {
		return nil, err
	}
	var stop = make(chan struct{})
	go func() {
		for {
			select {
			case endpoint := <-endpoints:
				var state = resolver.State{
					Addresses: make([]resolver.Address, 0),
				}
				for _, node := range endpoint.Nodes {
					var address resolver.Address
					address.Addr = node.Address
					address.ServerName = node.Name
					address.Attributes = attributes.New(application.KeyServiceInfo, node.Weight)
					state.Addresses = append(state.Addresses, address)
				}
				_ = cc.UpdateState(state)
			case <-stop:
				return
			}
		}
	}()
	return &EtcdResolver{
		stop: stop,
	}, nil
}

func (e EtcdBuilder) Scheme() string {
	return e.scheme
}

type EtcdResolver struct {
	stop chan struct{}
}

func (e EtcdResolver) ResolveNow(options resolver.ResolveNowOptions) {
	//fmt.Println("ResolveNow", options)
}

func (e EtcdResolver) Close() {
	e.stop <- struct{}{}
}
