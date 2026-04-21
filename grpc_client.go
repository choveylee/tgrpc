package tgrpc

import (
	"context"
	"errors"
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
		return errors.New("tgrpc: cannot invoke unary RPC: client is nil")
	}
	if p.conn == nil {
		return errors.New("tgrpc: cannot invoke unary RPC: connection is nil")
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
		if p.conn != nil {
			err = p.conn.Close()
		}
	})
	return err
}

// NewGrpcClient creates a channel with [grpc.NewClient], applying OTel stats,
// logging and metadata interceptors, and insecure credentials. Use
// [GrpcClient.Invoke] for unary RPCs. Call [GrpcClient.Close] when done, or
// cancel ctx to close the connection in the background (errors are logged at
// warn level).
func NewGrpcClient(ctx context.Context, grpcOption GrpcOption, address string) (*GrpcClient, error) {
	options := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			logClientInterceptor,
			metaDataClientInterceptor,
		),
		grpc.WithAuthority(address),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	options = append(options, grpcOption.options...)

	conn, err := grpc.NewClient(address, options...)
	if err != nil {
		return nil, err
	}

	grpcClient := &GrpcClient{
		conn: conn,
	}

	go func() {
		<-ctx.Done()
		if err := grpcClient.Close(); err != nil {
			tlog.W(context.Background()).Err(err).Msgf("close grpc conn (%s) err (%v).",
				address, err)
		}
	}()

	return grpcClient, nil
}
