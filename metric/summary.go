package metric

import "github.com/prometheus/client_golang/prometheus"

// SummaryVecOpts ...
type SummaryVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

// Build ...
func (opts SummaryVecOpts) Build() *summaryVec {
	vec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels)
	prometheus.MustRegister(vec)
	return &summaryVec{
		SummaryVec: vec,
	}
}

// NewSummaryVec ...
func NewSummaryVec(name, help string, labels []string) *summaryVec {
	config := DefaultConfig()
	return SummaryVecOpts{
		Namespace: config.NameSpace,
		Subsystem: config.SubSystem,
		Name:      name,
		Help:      help,
		Labels:    labels,
	}.Build()
}

type summaryVec struct {
	*prometheus.SummaryVec
}

// Observe ...
func (summary *summaryVec) Observe(v float64, labels ...string) {
	summary.WithLabelValues(labels...).Observe(v)
}
