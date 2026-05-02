package tgrpc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/choveylee/tlog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient wraps a gRPC client connection created by [NewGrpcClient].
// Do not copy a non-zero GrpcClient: [Close] must run at most once per connection.
type GrpcClient struct {
	conn *grpc.ClientConn
	once sync.Once

	stopContextWaiter context.CancelFunc
	contextWaiterDone chan struct{}
}

// Conn returns the underlying [grpc.ClientConn].
func (p *GrpcClient) Conn() *grpc.ClientConn {
	if p == nil {
		return nil
	}
	return p.conn
}

// Invoke sends a unary RPC. It is a thin wrapper around
// [*grpc.ClientConn.Invoke] so callers can use the recommended API without
// reaching through [GrpcClient.Conn].
func (p *GrpcClient) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	if p == nil {
		return errors.New("tgrpc: cannot invoke unary RPC: client receiver is nil")
	}
	if p.conn == nil {
		return errors.New("tgrpc: cannot invoke unary RPC: underlying connection is nil")
	}
	if ctx == nil {
		return errors.New("tgrpc: cannot invoke unary RPC: context is nil")
	}
	return p.conn.Invoke(ctx, method, args, reply, opts...)
}

// Close closes the client connection. It is safe to call more than once; only
// the first call closes the connection. A nil receiver is a no-op.
func (p *GrpcClient) Close() error {
	if p == nil {
		return nil
	}
	var err error
	p.once.Do(func() {
		if p.stopContextWaiter != nil {
			p.stopContextWaiter()
		}
		if p.conn != nil {
			err = p.conn.Close()
		}
	})
	return err
}

// NewGrpcClient creates a [GrpcClient] with [grpc.NewClient], applying the
// OpenTelemetry client stats handler and the unary access-log interceptor.
// When the caller does not configure transport security, the constructor uses
// default insecure transport credentials. Use [GrpcClient.Invoke] for unary
// RPCs. Call [GrpcClient.Close] when the client is no longer needed, or cancel
// ctx to close the connection in the background. Background close failures are
// logged at warn level.
func NewGrpcClient(ctx context.Context, grpcOption GrpcOption, address string) (*GrpcClient, error) {
	if ctx == nil {
		return nil, errors.New("tgrpc: cannot create gRPC client: context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("tgrpc: cannot create gRPC client: %w", err)
	}

	options := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(logClientInterceptor),
	}

	options = append(options, grpcOption.options...)

	conn, err := newClientConn(address, options)
	if err != nil {
		return nil, err
	}

	grpcClient := &GrpcClient{
		conn: conn,
	}

	if done := ctx.Done(); done != nil {
		waiterCtx, stopWaiter := context.WithCancel(context.Background())
		grpcClient.stopContextWaiter = stopWaiter
		grpcClient.contextWaiterDone = make(chan struct{})

		go func() {
			defer close(grpcClient.contextWaiterDone)

			select {
			case <-done:
				if err := grpcClient.Close(); err != nil {
					tlog.W(context.Background()).Err(err).Msgf(
						"Failed to close the gRPC client connection for target %s",
						address,
					)
				}
			case <-waiterCtx.Done():
			}
		}()
	}

	return grpcClient, nil
}

func newClientConn(address string, options []grpc.DialOption) (*grpc.ClientConn, error) {
	optionsWithDefaultInsecure := append(
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		options...,
	)

	conn, err := grpc.NewClient(address, optionsWithDefaultInsecure...)
	if err == nil {
		return conn, nil
	}

	conn, retryErr := grpc.NewClient(address, options...)
	if retryErr == nil {
		return conn, nil
	}

	return nil, fmt.Errorf(
		"tgrpc: failed to create gRPC client: attempt with default insecure transport failed: %v; attempt without default insecure transport failed: %w",
		err,
		retryErr,
	)
}
