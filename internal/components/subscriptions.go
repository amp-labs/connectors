package components

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
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

// WebhookMessageVerifier is the minimal interface for a connector that can verify webhook messages.
type WebhookMessageVerifier interface {
	VerifyWebhookMessage(
		ctx context.Context,
		request *common.WebhookRequest,
		params *common.VerificationParams,
	) (bool, error)
}

// SubscriptionCreator is the minimal interface for a connector that can create subscriptions.
type SubscriptionCreator interface {
	Subscribe(
		ctx context.Context,
		params common.SubscribeParams,
	) (*common.SubscriptionResult, error)
}

// SubscriptionUpdater is the minimal interface for a connector that can update subscriptions.
type SubscriptionUpdater interface {
	UpdateSubscription(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)
}

// SubscriptionRemover is the minimal interface for a connector that can delete subscriptions.
type SubscriptionRemover interface {
	DeleteSubscription(
		ctx context.Context,
		previousResult common.SubscriptionResult,
	) error
}

// Compile-time assertion that the minimal subscription interfaces
// satisfy connectors.SubscribeConnector.
//
// Each connector asserts only the interfaces it implements.
// If connectors.SubscribeConnector changes, the compiler forces updates,
// causing dependent connectors to fail in tests.
//
// Enables incremental, method-by-method implementation with full compatibility.
var (
	_ connectors.SubscribeConnector = (*dummySubscribeConnector)(nil)
)

// dummySubscribeConnector composes the minimal subscription interfaces and
// required base behavior. It has no implementations and exists purely for
// compile-time interface verification.
type dummySubscribeConnector struct {
	// Base.
	connectors.BatchRecordReaderConnector

	// Decomposed interfaces (primary).
	WebhookMessageVerifier
	SubscriptionCreator
	SubscriptionUpdater
	SubscriptionRemover

	// Supporting helpers (secondary).
	SubscriptionInputOutput[any, any]
}
