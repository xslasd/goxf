package balancer

import (
	"errors"
	"sync"

	"github.com/smallnest/weighted"
	"github.com/xslasd/goxf/application"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const (
	// NameSmoothWeightRoundRobin ...
	NameSmoothWeightRoundRobin = "swr"
)

func init() {
	balancer.Register(
		base.NewBalancerBuilder(NameSmoothWeightRoundRobin, &swrPickerBuilder{}, base.Config{HealthCheck: true}),
	)
}

type swrPickerBuilder struct{}

// Build ...
func (s swrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	return newSWRPicker(info)
}

type swrPicker struct {
	readySCs map[balancer.SubConn]base.SubConnInfo
	mu       sync.Mutex
	next     int
	buckets  *weighted.SW
	// todo : Wait for the implementation
	// routeBuckets map[string]*weighted.SW
}

func newSWRPicker(info base.PickerBuildInfo) *swrPicker {
	picker := &swrPicker{
		buckets:  &weighted.SW{},
		readySCs: info.ReadySCs,
		// routeBuckets: map[string]*weighted.SW{},
	}
	picker.parseBuildInfo(info)
	return picker
}

// Pick ...
func (p *swrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var buckets = p.buckets
	sub, ok := buckets.Next().(balancer.SubConn)
	if ok {
		return balancer.PickResult{SubConn: sub}, nil
	}
	return balancer.PickResult{}, errors.New("pick failed")
}

func (p *swrPicker) parseBuildInfo(info base.PickerBuildInfo) {
	for subConn, item := range info.ReadySCs {
		var weight = 1
		attributes := item.Address.Attributes
		if attributes != nil {
			attr := attributes.Value(application.KeyServiceInfo)
			if wInt, ok := attr.(int); ok {
				//log.Debug("get weight from ServiceInfo for pick", log.String("name", serviceInfo.Name), log.String("address", serviceInfo.Address), log.Int("weight", serviceInfo.Weight))
				weight = wInt
			}
		}
		p.buckets.Add(subConn, weight)
	}
}
