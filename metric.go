/**
 * @Author: lidonglin
 * @Description:
 * @File:  metric.go
 * @Version: 1.0.0
 * @Date: 2023/12/10 15:24
 */

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
