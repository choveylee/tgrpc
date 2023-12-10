/**
 * @Author: lidonglin
 * @Description:
 * @File:  grpc_client.go
 * @Version: 1.0.0
 * @Date: 2023/12/10 15:03
 */

package tgrpc

import (
	"context"

	"github.com/choveylee/tlog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	conn *grpc.ClientConn
}

func (p *GrpcClient) Conn() *grpc.ClientConn {
	return p.conn
}

func NewGrpcClient(ctx context.Context, address string, opts ...grpc.DialOption) (*GrpcClient, error) {
	options := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			logClientInterceptor,
			metaDataClientInterceptor,
		),
		grpc.WithAuthority(address),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithBlock(),
	}

	if opts != nil {
		options = append(options, opts...)
	}

	conn, err := grpc.Dial(address, options...)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		err := conn.Close()
		if err != nil {
			tlog.W(ctx).Err(err).Msgf("close grpc conn (%s) err (%v).",
				address, err)
		}
	}()

	grpcClient := &GrpcClient{
		conn: conn,
	}

	return grpcClient, nil
}
