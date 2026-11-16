package tgrpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// observedUnaryCall marks a call measured by GrpcClient.Invoke. An individual
// call option avoids suppressing unrelated calls that reuse its context.
type observedUnaryCall struct {
	grpc.EmptyCallOption
	conn *grpc.ClientConn
}

// logClientInterceptor validates and observes calls made through the underlying
// connection. It avoids recording a second result for GrpcClient.Invoke.
func logClientInterceptor(ctx context.Context, method string, request, response any, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, options ...grpc.CallOption) error {
	invoke := func() error {
		return invoker(ctx, method, request, response, conn, options...)
	}

	for _, option := range options {
		if observed, ok := option.(observedUnaryCall); ok && observed.conn == conn {
			if err := validateUnaryCall(response, options); err != nil {
				return err
			}

			return invoke()
		}
	}

	return observeUnaryCall(ctx, method, response, options, invoke)
}

// observeUnaryCall measures a validated call before sending results to telemetry sinks.
func observeUnaryCall(ctx context.Context, fullMethod string, response any, options []grpc.CallOption, invoke func() error) error {
	startTime := time.Now()

	err := validateUnaryCall(response, options)
	if err == nil {
		err = invoke()
	}

	duration := time.Since(startTime)
	service, method, _ := splitMethodName(fullMethod)
	code := rpcStatusCode(err)

	observeGrpcClientLatency(ctx, duration, "unary", service, method, code.String())
	logFormatter(ctx, service, method, duration, code)

	return err
}

// validateUnaryCall rejects nil containers and options before the codec or
// gRPC applies them. Non-nil containers may be used with custom codecs.
func validateUnaryCall(response any, options []grpc.CallOption) error {
	if isNilValue(response) {
		return errors.New("tgrpc: gRPC response must not be nil")
	}

	return validateCallOptions(options)
}

// validateCallOptions rejects nil interface values before gRPC applies them.
func validateCallOptions(options []grpc.CallOption) error {
	for _, option := range options {
		if isNilValue(option) {
			return errors.New("tgrpc: gRPC call option must not be nil")
		}
	}

	return nil
}

// splitMethodName maps malformed RPC paths to stable telemetry labels.
func splitMethodName(fullMethod string) (string, string, bool) {
	path, prefixed := strings.CutPrefix(fullMethod, "/")
	service, method, separated := strings.Cut(path, "/")
	if !prefixed || !separated || !protoreflect.FullName(service).IsValid() || !protoreflect.Name(method).IsValid() {
		return "unknown", "unknown", false
	}

	return service, method, true
}

// rpcStatusCode preserves gRPC statuses and classifies plain context errors.
func rpcStatusCode(err error) codes.Code {
	if rpcStatus, ok := status.FromError(err); ok {
		return rpcStatus.Code()
	}

	return status.FromContextError(err).Code()
}

type grpcClientStats struct {
	stats.Handler
}

// TagRPC normalizes the method used by OpenTelemetry without changing the
// original method sent over the connection or passed to other stats handlers.
func (handler grpcClientStats) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if _, _, ok := splitMethodName(info.FullMethodName); !ok {
		sanitizedInfo := *info
		sanitizedInfo.FullMethodName = "/unknown/unknown"

		info = &sanitizedInfo
	}

	return handler.Handler.TagRPC(ctx, info)
}

// HandleRPC forwards the canonical status code to OpenTelemetry. The server's
// free-form status description remains available on the original RPC result.
func (handler grpcClientStats) HandleRPC(ctx context.Context, event stats.RPCStats) {
	if end, ok := event.(*stats.End); ok && end.Error != nil {
		sanitizedEnd := *end
		sanitizedEnd.Error = status.Error(rpcStatusCode(end.Error), "")

		event = &sanitizedEnd
	}

	handler.Handler.HandleRPC(ctx, event)
}
