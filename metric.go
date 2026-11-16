package tgrpc

import (
	"context"
	"sync"
	"time"

	"github.com/choveylee/tlog"
	"github.com/choveylee/tmetric"
)

var (
	grpcClientLatency, grpcClientLatencyInitErr = tmetric.NewHistogramVec(
		"grpc_client_latency",
		"histogram of grpc client request latency (milliseconds)",
		[]string{"type", "service", "method", "code"},
	)
	grpcClientLatencyWarnOnce sync.Once
)

// observeGrpcClientLatency records the measured RPC duration in milliseconds.
func observeGrpcClientLatency(ctx context.Context, duration time.Duration, rpcType, service, method, code string) {
	if grpcClientLatency == nil {
		grpcClientLatencyWarnOnce.Do(func() {
			if grpcClientLatencyInitErr != nil {
				tlog.W(ctx).Err(grpcClientLatencyInitErr).Msg("Initialize gRPC client latency metric error")

				return
			}

			tlog.W(ctx).Msg("gRPC client latency metric is unavailable")
		})

		return
	}

	if err := grpcClientLatency.Observe(float64(duration)/float64(time.Millisecond), rpcType, service, method, code); err != nil {
		tlog.W(ctx).Err(err).Msg("Record gRPC client latency metric error")
	}
}
