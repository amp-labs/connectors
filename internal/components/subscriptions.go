package components

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrInvalidSubscriptionRequestType = errors.New("request type of common.SubscribeParams is invalid")
	ErrInvalidSubscriptionResultType  = errors.New("result type of common.SubscriptionResult is invalid")
	ErrMissingSubscriptionRequest     = errors.New("common.SubscribeParams.Request is missing")
	ErrMissingSubscriptionResult      = errors.New("common.SubscriptionResult.Result is missing")
)

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
func (s SubscriptionInputOutput[I, O]) TypedSubscriptionRequest(params common.SubscribeParams) (*I, error) {
	if params.Request == nil {
		return nil, ErrMissingSubscriptionRequest
	}

	input, ok := params.Request.(*I)
	if !ok {
		return nil, fmt.Errorf("%w: expected %T, got %T", ErrInvalidSubscriptionRequestType, input, params.Request)
	}

	return input, nil
}

// TypedSubscriptionResult extracts and casts subscription.Result into the concrete output type O.
//
// It provides a safe way to recover the typed result from the
// non-generic SubscriptionResult, returning an error if the underlying
// type does not match.
func (s SubscriptionInputOutput[I, O]) TypedSubscriptionResult(subscription common.SubscriptionResult) (*O, error) {
	if subscription.Result == nil {
		return nil, ErrMissingSubscriptionResult
	}

	output, ok := subscription.Result.(*O)
	if !ok {
		return nil, fmt.Errorf("%w: expected %T, got %T", ErrInvalidSubscriptionResultType, output, subscription.Result)
	}

	return output, nil
}
