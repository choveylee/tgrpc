package tgrpc

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
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
	startTime := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(startTime)
	service, shortMethod := splitMethodName(method)

	logFormatter(ctx, service, shortMethod, duration, req, reply, err)

	observeGrpcClientLatency(ctx, startTime, "unary", service, shortMethod, status.Code(err).String())

	return err
}
