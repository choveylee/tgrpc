package tgrpc

import (
	"context"
	"sync"
	"time"

	"github.com/choveylee/tlog"
	"github.com/choveylee/tmetric"
)

var (
	grpcClientLatency         *tmetric.HistogramVec
	grpcClientLatencyInitErr  error
	grpcClientLatencyWarnOnce sync.Once
)

func init() {
	grpcClientLatency, grpcClientLatencyInitErr = tmetric.NewHistogramVec(
		"grpc_client_latency",
		"histogram of grpc client request latency (milliseconds)",
		[]string{"type", "service", "method", "code"},
	)
}

func observeGrpcClientLatency(ctx context.Context, startTime time.Time, rpcType, service, method, code string) {
	if grpcClientLatency == nil {
		grpcClientLatencyWarnOnce.Do(func() {
			if grpcClientLatencyInitErr != nil {
				tlog.W(ctx).Err(grpcClientLatencyInitErr).Msg("The gRPC client latency metric is disabled because collector initialization failed")
				return
			}

			tlog.W(ctx).Msg("The gRPC client latency metric is unavailable")
		})
		return
	}

	if err := grpcClientLatency.Observe(tmetric.SinceMS(startTime), rpcType, service, method, code); err != nil {
		tlog.W(ctx).Err(err).Msg("Failed to record the gRPC client latency metric")
	}
}
