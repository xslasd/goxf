package metric

const (
	GRPCUnaryType  = "GRPCUnary"
	GRPCStreamType = "GRPCStream"
	HTTPType       = "HTTP"
	RedisType      = "Redis"
	JobType        = "Job"
)

var (
	config = DefaultConfig()
	// ServerHandleCounter ...
	ServerHandleCounter = NewCounterVec(
		"server_handle_total",
		"the total count of server handle",
		[]string{"type", "method", "code"},
	)

	// ServerHandleHistogram ...
	ServerHandleHistogram = NewHistogramVec(
		"server_handle_seconds",
		"the histogram of server handle duration",
		[]string{"type", "method"},
	)

	// ClientHandleCounter ...
	ClientHandleCounter = NewCounterVec(
		"client_handle_total",
		"the total count of client invoke",
		[]string{"type", "name", "method", "peer", "code"},
	)

	// ClientHandleHistogram ...
	ClientHandleHistogram = NewHistogramVec(
		"client_handle_seconds",
		"the histogram of client invoke duration",
		[]string{"type", "name", "method", "peer"},
	)
)
