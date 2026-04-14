package components

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// SubscriptionInputOutput is a generic helper that provides type-safe implementations
// of the EmptySubscriptionParams and EmptySubscriptionResult methods
// required by connectors.SubscribeConnector.
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

// SubscriptionUpdator is the minimal interface for a connector that can update subscriptions.
type SubscriptionUpdator interface {
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
	SubscriptionUpdator
	SubscriptionRemover

	// Supporting helpers (secondary).
	SubscriptionInputOutput[any, any]
}
