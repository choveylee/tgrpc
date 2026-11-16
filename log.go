package tgrpc

import (
	"context"
	"time"

	"github.com/choveylee/tlog"
	"google.golang.org/grpc/codes"
)

// logFormatter records unary RPC metadata and the canonical status code.
func logFormatter(ctx context.Context, service, method string, duration time.Duration, code codes.Code) {
	var event *tlog.Tevent
	if code == codes.OK {
		event = tlog.D(ctx)
	} else {
		event = tlog.E(ctx)
	}

	event.Detailf("service:%s", service).
		Detailf("method:%s", method).
		Detailf("latency:%v", duration).
		Detailf("code:%s", code.String()).
		Msg("Complete unary gRPC call")
}
