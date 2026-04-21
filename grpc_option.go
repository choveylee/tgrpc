package tgrpc

import (
	"google.golang.org/grpc"
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

// WithDialOption appends a dial option.
func (p *GrpcOption) WithDialOption(option grpc.DialOption) {
	p.options = append(p.options, option)
}
