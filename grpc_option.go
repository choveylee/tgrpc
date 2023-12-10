/**
 * @Author: lidonglin
 * @Description:
 * @File:  grpc_interceptor.go
 * @Version: 1.0.0
 * @Date: 2023/12/10 16:29
 */

package tgrpc

import (
	"google.golang.org/grpc"
)

type GrpcOption struct {
	options []grpc.DialOption
}

func NewGrpcOption() *GrpcOption {
	return &GrpcOption{
		options: make([]grpc.DialOption, 0),
	}
}

func (p *GrpcOption) WithDialOption(option grpc.DialOption) {
	p.options = append(p.options, option)
}
