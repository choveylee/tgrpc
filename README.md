# tgrpc

`tgrpc` is a small Go library that constructs [gRPC](https://grpc.io/) client channels with sensible defaults for observability and operations: OpenTelemetry client instrumentation, structured access logging, client-side latency metrics, and unary interceptors. It wraps [`grpc.NewClient`](https://pkg.go.dev/google.golang.org/grpc#NewClient) and exposes unary RPCs through [`ClientConn.Invoke`](https://pkg.go.dev/google.golang.org/grpc#ClientConn.Invoke).

## Features

- **Modern client API** — Uses `grpc.NewClient` and `GrpcClient.Invoke` for unary calls; streaming APIs remain available via `GrpcClient.Conn`.
- **OpenTelemetry** — Registers the gRPC client stats handler from [`otelgrpc`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc).
- **Access logging** — Unary interceptor emits request metadata, latency, status, and JSON-marshaled protobuf payloads through [`tlog`](https://github.com/choveylee/tlog).
- **Metrics** — Histogram `grpc_client_latency` (milliseconds) with labels `type`, `service`, `method`, and `code`, via [`tmetric`](https://github.com/choveylee/tmetric) / Prometheus.
- **Lifecycle** — Idempotent `Close`; optional automatic shutdown when the constructor `context` is canceled (close errors are logged at warn level).

## Requirements

- Go **1.25** or later (see `go.mod`).

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

	opts := tgrpc.NewGrpcOption()
	client, err := tgrpc.NewGrpcClient(ctx, *opts, "127.0.0.1:50051")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var req MyRequest
	var resp MyResponse
	if err := client.Invoke(ctx, "/package.Service/Method", &req, &resp); err != nil {
		log.Fatal(err)
	}
}
```

Substitute `MyRequest`, `MyResponse`, and `/package.Service/Method` with types and the method string from your `.proto` generated code.

**Transport security:** `NewGrpcClient` uses **insecure** credentials by default. For TLS or other transports, add the corresponding `grpc.DialOption` values through `GrpcOption.WithDialOption`, consistent with [gRPC Go](https://github.com/grpc/grpc-go) documentation on credentials and channel configuration.

## Client options

`GrpcOption` collects additional [`grpc.DialOption`](https://pkg.go.dev/google.golang.org/grpc#DialOption) values passed to `grpc.NewClient` after the library defaults. Use it for transport credentials, custom authority, keepalive, and other channel settings.

## Observability

| Mechanism        | Description |
|------------------|-------------|
| Tracing / stats  | OpenTelemetry gRPC client handler (`otelgrpc`). |
| Logs             | Per-unary-call access log via `tlog` (includes serialized request/response when inputs are `proto.Message`). |
| Metrics          | `grpc_client_latency` histogram (ms); labels: `type`, `service`, `method`, `code`. |

Ensure your process configures `tlog`, Prometheus registration for `tmetric`, and OpenTelemetry exporters according to your environment.

## Lifecycle and concurrency

- Call `GrpcClient.Close` when the client is no longer needed, or cancel the `context` passed to `NewGrpcClient` to close the connection in a background goroutine.
- `Close` is safe to call multiple times; only the first call performs the shutdown.
- Do not copy a non-zero `GrpcClient`; treat it as owning the underlying `*grpc.ClientConn`.

## API overview

| Symbol | Role |
|--------|------|
| `NewGrpcClient` | Builds a `GrpcClient` with default interceptors, OTel stats, and insecure credentials unless overridden. |
| `GrpcClient.Invoke` | Sends a unary RPC (`ClientConn.Invoke`). |
| `GrpcClient.Conn` | Returns `*grpc.ClientConn` for streaming or advanced use. |
| `GrpcClient.Close` | Closes the channel (`error`, idempotent). |
| `NewGrpcOption` / `GrpcOption.WithDialOption` | Extra dial options. |

## Documentation

Package documentation is available with:

```bash
go doc github.com/choveylee/tgrpc
```

## Related modules

This library depends on internal Chovey Lee tooling (`tlog`, `tmetric`) and OpenTelemetry gRPC instrumentation. Version pins are listed in `go.mod`.
