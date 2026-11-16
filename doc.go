// Package tgrpc constructs gRPC clients with [grpc.NewClient], applying the
// OpenTelemetry client stats handler, a unary access-log interceptor, and a
// Prometheus latency histogram. Default telemetry records the gRPC status code
// and leaves the complete error in the RPC result. An empty option set uses
// plaintext. Additional dial options require an explicit transport choice.
// Pass [GrpcClient] to generated unary or streaming clients. Release resources with
// [GrpcClient.Close], or cancel the context passed to [NewGrpcClient] to close
// the connection in the background.
package tgrpc
