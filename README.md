# tgrpc

`tgrpc` is a small Go library for constructing [gRPC](https://grpc.io/) client channels with production-oriented defaults for observability and operational safety: OpenTelemetry client instrumentation, structured access logging, client-side latency metrics, and a unary access-log interceptor. It wraps [`grpc.NewClient`](https://pkg.go.dev/google.golang.org/grpc#NewClient) and exposes unary RPCs through [`ClientConn.Invoke`](https://pkg.go.dev/google.golang.org/grpc#ClientConn.Invoke).

## Features

- **Modern client API** — `GrpcClient` implements `grpc.ClientConnInterface` for generated unary and streaming clients.
- **OpenTelemetry** — Registers the gRPC client stats handler from [`otelgrpc`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc).
- **Access logging** — The unary interceptor records the service, method, elapsed time, and gRPC status code through [`tlog`](https://github.com/choveylee/tlog).
- **Metrics** — Histogram `grpc_client_latency` (milliseconds) with labels `type`, `service`, `method`, and `code`, via [`tmetric`](https://github.com/choveylee/tmetric) / Prometheus.
- **Lifecycle** — Provides idempotent `Close`; optionally shuts down the connection automatically when the constructor `context` is canceled (close failures are logged at warn level).

## Requirements

- Go **1.27.4** or later (see `go.mod`).

## Installation

```bash
go get github.com/choveylee/tgrpc@latest
```

## Usage

The example below shows a unary RPC. Replace the method name, request, and response types with those generated for your service.

```go
package main

import (
	"context"
	"log"

	"github.com/choveylee/tgrpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	options := tgrpc.NewGrpcOption()
	client, err := tgrpc.NewGrpcClient(ctx, *options, "127.0.0.1:50051")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var request MyRequest
	var response MyResponse
	if err := client.Invoke(ctx, "/package.Service/Method", &request, &response); err != nil {
		log.Fatal(err)
	}
}
```

Substitute `MyRequest`, `MyResponse`, and `/package.Service/Method` with types and the method string from your `.proto` generated code.

An empty `GrpcOption` defaults to plaintext. When adding dial options, select transport explicitly. Use `WithTransportCredentials` or `WithCredentialsBundle` for TLS, or call `WithInsecure()` for an intentional plaintext connection. For example, a plaintext client that adds a keepalive option also calls `WithInsecure()`.

## Client options

`GrpcOption` collects additional [`grpc.DialOption`](https://pkg.go.dev/google.golang.org/grpc#DialOption) values passed to `grpc.NewClient`. Use `WithDialOption` for advanced channel settings such as keepalive, authority overrides, or resolver configuration. Copies of an option set can be extended independently.

## Observability

| Mechanism        | Description |
|------------------|-------------|
| Tracing / stats  | OpenTelemetry gRPC client handler (`otelgrpc`) with canonical response status. |
| Logs             | Unary service, method, duration, and status code via `tlog`. |
| Metrics          | `grpc_client_latency` histogram (ms); labels: `type`, `service`, `method`, `code`. |

Ensure your process configures `tlog`, Prometheus registration for `tmetric`, and OpenTelemetry exporters according to your environment.
The complete gRPC error remains available to the caller.
Malformed RPC method paths use `unknown` labels in logs, metrics, and built-in traces.
Pass `client` to generated constructors such as `servicepb.NewServiceClient(client)` so calls completed by custom unary interceptors are also measured. When using the raw connection from `client.Conn()`, add custom unary interceptors with `grpc.WithChainUnaryInterceptor` to keep them inside the access-log and metrics interceptor.

## Lifecycle and concurrency

- `NewGrpcClient` creates a channel that connects on demand. Use a client lifetime context for construction and a per-call context for each RPC.
- Call `GrpcClient.Close` or cancel the lifetime context to close the channel. Concurrent `Close` callers receive the same shutdown result.
- Do not copy a non-zero `GrpcClient`; treat it as owning the underlying `*grpc.ClientConn`.

## API overview

| Symbol | Role |
|--------|------|
| `NewGrpcClient` | Builds a client with a unary access-log interceptor and OpenTelemetry stats. |
| `GrpcClient.Invoke` | Validates, sends, and observes a unary RPC. |
| `GrpcClient.NewStream` | Opens a streaming RPC for a generated client. |
| `GrpcClient.Conn` | Returns the raw `*grpc.ClientConn` for advanced use. |
| `GrpcClient.Close` | Closes the channel (`error`, idempotent). |
| `GrpcOption` | Additional channel options, including explicit TLS credentials, a credentials bundle, or `WithInsecure()`. |

## Documentation

Package documentation is available with:

```bash
go doc github.com/choveylee/tgrpc
```

## Related modules

This library depends on internal Chovey Lee tooling (`tlog`, `tmetric`) and OpenTelemetry gRPC instrumentation. Version pins are listed in `go.mod`.
