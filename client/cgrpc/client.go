package cgrpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func newGRPCClient(ctx context.Context, config *Config, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialOptions = append(
		dialOptions,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.DialKeepAliveTime,    // send pings every 10 seconds if there is no activity
			Timeout:             config.DialKeepAliveTimeout, // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: config.PermitWithoutStream,  // send pings even without active streams
		}),
	)
	return grpc.DialContext(ctx, config.Address, dialOptions...)
}
