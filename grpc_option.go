package tgrpc

import (
	"errors"
	"reflect"
	"slices"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcOption collects extra [grpc.DialOption] values for [NewGrpcClient].
// A copy can be extended independently of the original.
type GrpcOption struct {
	options       []grpc.DialOption
	validationErr error
}

// NewGrpcOption returns an empty option set.
func NewGrpcOption() *GrpcOption {
	return new(GrpcOption)
}

// WithDialOption appends a raw [grpc.DialOption].
// Explicitly select transport credentials or call [GrpcOption.WithInsecure]
// when adding raw options.
func (optionSet *GrpcOption) WithDialOption(option grpc.DialOption) {
	if isNilValue(option) {
		optionSet.validationErr = errors.New("tgrpc: gRPC dial option must not be nil")

		return
	}

	optionSet.appendDialOption(option)
}

// WithTransportCredentials appends transport credentials to the option set.
func (optionSet *GrpcOption) WithTransportCredentials(transportCredentials credentials.TransportCredentials) {
	if isNilValue(transportCredentials) {
		optionSet.validationErr = errors.New("tgrpc: transport credentials must not be nil")

		return
	}

	optionSet.appendDialOption(grpc.WithTransportCredentials(transportCredentials))
}

// WithCredentialsBundle appends a credentials bundle to the option set.
func (optionSet *GrpcOption) WithCredentialsBundle(bundle credentials.Bundle) {
	if isNilValue(bundle) {
		optionSet.validationErr = errors.New("tgrpc: credentials bundle must not be nil")

		return
	}

	optionSet.appendDialOption(grpc.WithCredentialsBundle(bundle))
}

// WithInsecure explicitly selects plaintext transport.
func (optionSet *GrpcOption) WithInsecure() {
	optionSet.appendDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// appendDialOption makes append allocate instead of overwriting a slot shared
// with another copy of the option set.
func (optionSet *GrpcOption) appendDialOption(option grpc.DialOption) {
	optionSet.options = append(slices.Clip(optionSet.options), option)
}

// isNilValue detects ordinary and typed nils carried by an interface.
func isNilValue(value any) bool {
	reflectedValue := reflect.ValueOf(value)
	switch reflectedValue.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflectedValue.IsNil()
	}

	return false
}
