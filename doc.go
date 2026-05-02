// Package tgrpc constructs gRPC clients with [grpc.NewClient], applying the
// OpenTelemetry client stats handler, a unary access-log interceptor, and a
// Prometheus latency histogram. The package uses default insecure transport
// credentials only when the caller does not configure transport security.
// Send unary RPCs with [GrpcClient.Invoke]. Release resources with
// [GrpcClient.Close], or cancel the context passed to [NewGrpcClient] to close
// the connection in the background.
package tgrpc
