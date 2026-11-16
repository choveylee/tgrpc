package tgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/choveylee/tlog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient wraps a gRPC client connection and implements [grpc.ClientConnInterface].
// Do not copy a non-zero GrpcClient: [Close] must run at most once per connection.
type GrpcClient struct {
	conn *grpc.ClientConn

	stopAutoClose func() bool
	closeOnce     sync.Once
	closeErr      error
}

var _ grpc.ClientConnInterface = (*GrpcClient)(nil)

// Conn returns the underlying [grpc.ClientConn].
func (client *GrpcClient) Conn() *grpc.ClientConn {
	if client == nil {
		return nil
	}

	return client.conn
}

// Invoke validates and observes a unary RPC, including calls completed by an
// interceptor before transport.
func (client *GrpcClient) Invoke(ctx context.Context, method string, request, response any, options ...grpc.CallOption) error {
	if client == nil || client.conn == nil {
		return errors.New("tgrpc: gRPC client must be initialized")
	}

	if ctx == nil {
		return errors.New("tgrpc: context must not be nil")
	}

	return observeUnaryCall(ctx, method, response, options, func() error {
		observed := observedUnaryCall{conn: client.conn}

		callOptions := append([]grpc.CallOption{observed}, options...)

		return client.conn.Invoke(ctx, method, request, response, callOptions...)
	})
}

// NewStream opens an RPC stream with the underlying gRPC connection.
func (client *GrpcClient) NewStream(ctx context.Context, description *grpc.StreamDesc, method string, options ...grpc.CallOption) (grpc.ClientStream, error) {
	if client == nil || client.conn == nil {
		return nil, errors.New("tgrpc: gRPC client must be initialized")
	}

	if ctx == nil {
		return nil, errors.New("tgrpc: context must not be nil")
	}

	if description == nil {
		return nil, errors.New("tgrpc: gRPC stream description must not be nil")
	}

	if err := validateCallOptions(options); err != nil {
		return nil, err
	}

	return client.conn.NewStream(ctx, description, method, options...)
}

// Close stops automatic shutdown and waits for the shared connection close.
// Concurrent callers receive the result of the same shutdown. A nil receiver is a no-op.
func (client *GrpcClient) Close() error {
	if client == nil {
		return nil
	}

	if client.stopAutoClose != nil {
		client.stopAutoClose()
	}

	return client.closeConnection()
}

// NewGrpcClient creates a [GrpcClient] with [grpc.NewClient], applying the
// OpenTelemetry client stats handler and the unary access-log interceptor.
// An empty option set defaults to plaintext. An option set with dial options
// must select its transport explicitly. The channel connects on demand.
// Call [GrpcClient.Close] or cancel its lifetime context to close the channel.
func NewGrpcClient(ctx context.Context, optionSet GrpcOption, address string) (*GrpcClient, error) {
	if ctx == nil {
		return nil, errors.New("tgrpc: context must not be nil")
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("tgrpc: create gRPC client: %w", err)
	}

	if strings.TrimSpace(address) == "" {
		return nil, errors.New("tgrpc: gRPC target must not be empty")
	}

	if optionSet.validationErr != nil {
		return nil, optionSet.validationErr
	}

	statsHandler := grpcClientStats{Handler: otelgrpc.NewClientHandler()}

	options := []grpc.DialOption{
		grpc.WithStatsHandler(statsHandler),
		grpc.WithChainUnaryInterceptor(logClientInterceptor),
	}

	if len(optionSet.options) == 0 {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		options = append(options, optionSet.options...)
	}

	conn, err := newClientConn(address, options)
	if err != nil {
		return nil, fmt.Errorf("tgrpc: create gRPC client: %w", err)
	}

	if err := ctx.Err(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("tgrpc: create gRPC client: %w", err)
	}

	grpcClient := &GrpcClient{
		conn: conn,
	}

	if ctx.Done() != nil {
		grpcClient.stopAutoClose = context.AfterFunc(ctx, func() {
			if err := grpcClient.closeConnection(); err != nil {
				tlog.W(ctx).Err(err).Msg("Close gRPC client connection error")
			}
		})
	}

	return grpcClient, nil
}

// closeConnection runs shutdown once. The cancellation callback does not need
// the stop function, which may still be getting assigned when it fires.
func (client *GrpcClient) closeConnection() error {
	client.closeOnce.Do(func() {
		if client.conn != nil {
			client.closeErr = client.conn.Close()
		}
	})

	return client.closeErr
}

// newClientConn converts panics raised by opaque dial options into constructor
// errors, without exposing values carried by credentials or options.
func newClientConn(address string, options []grpc.DialOption) (conn *grpc.ClientConn, err error) {
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("gRPC client configuration panicked (%T)", value)
		}
	}()

	return grpc.NewClient(address, options...)
}
