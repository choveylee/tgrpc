package tgrpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GrpcOption collects extra [grpc.DialOption] values for [NewGrpcClient].
type GrpcOption struct {
	options []grpc.DialOption
}

// NewGrpcOption returns an empty option set.
func NewGrpcOption() *GrpcOption {
	return &GrpcOption{
		options: make([]grpc.DialOption, 0),
	}
}

// WithDialOption appends a raw [grpc.DialOption].
// Prefer [GrpcOption.WithTransportCredentials] or
// [GrpcOption.WithCredentialsBundle] for transport security configuration.
func (p *GrpcOption) WithDialOption(option grpc.DialOption) {
	p.options = append(p.options, option)
}

// WithTransportCredentials appends transport credentials to the option set.
func (p *GrpcOption) WithTransportCredentials(creds credentials.TransportCredentials) {
	p.options = append(p.options, grpc.WithTransportCredentials(creds))
}

// WithCredentialsBundle appends a credentials bundle to the option set.
func (p *GrpcOption) WithCredentialsBundle(bundle credentials.Bundle) {
	p.options = append(p.options, grpc.WithCredentialsBundle(bundle))
}
