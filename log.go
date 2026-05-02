package tgrpc

import (
	"context"
	"time"

	"github.com/choveylee/tlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func logFormatter(ctx context.Context, service, method string, duration time.Duration, req interface{}, resp interface{}, err error) {
	code := status.Code(err)

	latency := duration

	var event *tlog.Tevent

	if code == codes.OK {
		event = tlog.D(ctx)
	} else {
		event = tlog.E(ctx).Err(err)
	}

	event = event.Detailf("service:%s", service).
		Detailf("method:%s", method).
		Detailf("latency:%v", latency).
		Detailf("code:%s", code.String())

	event = event.Detailf("request:%s", newProtoJSONLogValue(ctx, "request", req)).
		Detailf("response:%s", newProtoJSONLogValue(ctx, "response", resp))

	event.Msg("gRPC client access log")
}

type protoJSONLogValue struct {
	ctx        context.Context
	fieldName  string
	protoValue proto.Message
}

func newProtoJSONLogValue(ctx context.Context, fieldName string, value interface{}) protoJSONLogValue {
	message, _ := value.(proto.Message)

	return protoJSONLogValue{
		ctx:        ctx,
		fieldName:  fieldName,
		protoValue: message,
	}
}

func (p protoJSONLogValue) String() string {
	if p.protoValue == nil {
		return ""
	}

	data, err := protojson.Marshal(p.protoValue)
	if err != nil {
		tlog.W(p.ctx).Err(err).Msgf("Failed to marshal the gRPC %s payload", p.fieldName)
		return ""
	}

	return string(data)
}
