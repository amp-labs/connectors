package components

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var ErrInvalidSubscriptionRequestType = errors.New("request type of common.SubscribeParams is invalid")

// SubscriptionInputOutput is a generic helper that provides out-of-the-box,
// type-safe implementations of the EmptySubscriptionParams and
// EmptySubscriptionResult methods required by connectors.SubscribeConnector.
//
// It acts as a bridge around the non-generic interface by encapsulating the
// untyped fields (any) and restoring type safety via generics. By embedding
// this struct, connectors can work with concrete input/output types without
// manual casting boilerplate.
type SubscriptionInputOutput[I, O any] struct{}

func (SubscriptionInputOutput[I, O]) EmptySubscriptionParams() *common.SubscribeParams {
	var request I

	return &common.SubscribeParams{
		Request: &request,
	}
}

func (SubscriptionInputOutput[I, O]) EmptySubscriptionResult() *common.SubscriptionResult {
	var result O

	return &common.SubscriptionResult{
		Result: &result,
	}
}

// TypedSubscriptionRequest extracts and casts params.Request into the concrete input type I.
//
// It provides a safe way to recover the typed request from the
// non-generic SubscribeParams, returning an error if the underlying
// type does not match.
func (s SubscriptionInputOutput[I, O]) TypedSubscriptionRequest(params common.SubscribeParams) (I, error) {
	var input I

	if params.Request != nil {
		var ok bool

		input, ok = params.Request.(I)
		if !ok {
			return input, fmt.Errorf("%w: expected %T, got %T", ErrInvalidSubscriptionRequestType, input, params.Request)
		}
	}

	return input, nil
}
