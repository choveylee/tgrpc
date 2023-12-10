/**
 * @Author: lidonglin
 * @Description:
 * @File:  grpc_interceptor.go
 * @Version: 1.0.0
 * @Date: 2023/12/10 16:29
 */

package tgrpc

import (
	"context"
	"strings"
	"time"

	"github.com/choveylee/tmetric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func splitMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}

	return "unknown", "unknown"
}

func logClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	//ctx = log.WithLogTrace(ctx)

	startTime := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(startTime)

	var service string

	service, method = splitMethodName(method)

	logFormatter(ctx, service, method, duration, req, reply, err)

	grpcClientLatency.Observe(tmetric.SinceMS(startTime), "unary", service, method, status.Code(err).String())

	return err
}

func metaDataClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.MD{}

	// md.Set("app-id", cfg.DefaultString(cfg.DefaultRcraiAppName, ""))

	return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
}
