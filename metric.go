package tgrpc

import (
	"github.com/choveylee/tmetric"
)

var (
	grpcClientLatency, _ = tmetric.NewHistogramVec(
		"grpc_client_latency",
		"histogram of grpc client request latency (milliseconds)",
		[]string{"type", "service", "method", "code"},
	)
)
