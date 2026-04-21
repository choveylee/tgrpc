// Package tgrpc builds gRPC clients using [grpc.NewClient], with OpenTelemetry
// client stats, a unary access-logging interceptor, and Prometheus latency
// histograms. Send unary RPCs with [GrpcClient.Invoke]. Release
// resources with [GrpcClient.Close], or cancel the context passed to
// [NewGrpcClient] to close in the background.
package tgrpc
