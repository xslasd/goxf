package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CounterVecOpts Opts

// Build ...
func (opts CounterVecOpts) Build() *CounterVec {
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels)
	prometheus.MustRegister(vec)
	return &CounterVec{
		CounterVec: vec,
	}
}

// NewCounterVec ...
func NewCounterVec(name, help string, labels []string) *CounterVec {
	cfg := DefaultConfig()
	return CounterVecOpts{
		Namespace: cfg.NameSpace,
		Subsystem: cfg.SubSystem,
		Name:      name,
		Help:      help,
		Labels:    labels,
	}.Build()
}

type CounterVec struct {
	*prometheus.CounterVec
}

// Inc ...
func (counter *CounterVec) Inc(labels ...string) {
	counter.WithLabelValues(labels...).Inc()
}

// Add ...
func (counter *CounterVec) Add(v float64, labels ...string) {
	counter.WithLabelValues(labels...).Add(v)
}
