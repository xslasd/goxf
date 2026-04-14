package metric

import "github.com/prometheus/client_golang/prometheus"

type GaugeVecOpts Opts

// Build ...
func (opts GaugeVecOpts) Build() *gaugeVec {
	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels)
	prometheus.MustRegister(vec)
	return &gaugeVec{
		GaugeVec: vec,
	}
}

// NewGaugeVec ...
func NewGaugeVec(name, help string, labels []string) *gaugeVec {
	config := DefaultConfig()
	return GaugeVecOpts{
		Namespace: config.NameSpace,
		Subsystem: config.SubSystem,
		Name:      name,
		Help:      help,
		Labels:    labels,
	}.Build()
}

type gaugeVec struct {
	*prometheus.GaugeVec
}

// Inc ...
func (gv *gaugeVec) Inc(labels ...string) {
	gv.WithLabelValues(labels...).Inc()
}

// Add ...
func (gv *gaugeVec) Add(v float64, labels ...string) {
	gv.WithLabelValues(labels...).Add(v)
}

// Set ...
func (gv *gaugeVec) Set(v float64, labels ...string) {
	gv.WithLabelValues(labels...).Set(v)
}
