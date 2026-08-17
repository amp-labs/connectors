package stripe

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// UpdateSubscription reconciles the existing subscription (previousResult) with the new
// desired state (params). params.SubscriptionEvents is the full desired state: objects
// present in the previous state but absent from params are unsubscribed, since the shared
// endpoint's enabled_events is replaced with exactly the desired event set.
// Doc URL: https://docs.stripe.com/api/webhook_endpoints/update
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	prevState, err := validatePreviousResult(previousResult)
	if err != nil {
		return nil, err
	}

	payload, err := buildWebhookPayloadFromParams(params)
	if err != nil {
		return nil, err
	}

	response, err := c.updateWebhookEndpoint(ctx, prevState.WebhookId, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	// Stripe returns the signing secret only when the endpoint is created; carry it over
	// from the previous state so the stored result keeps verifying webhook signatures.
	response.Secret = prevState.Secret

	result, err := buildSubscriptionResult(response, params.SubscriptionEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to build subscription result: %w", err)
	}

	return result, nil
}

// validatePreviousResult validates and extracts the previous subscription result.
func validatePreviousResult(previousResult *common.SubscriptionResult) (*SubscriptionResult, error) {
	if previousResult == nil || previousResult.Result == nil {
		return nil, fmt.Errorf("%w: missing previousResult or previousResult.Result", errMissingParams)
	}

	prevState, ok := previousResult.Result.(*SubscriptionResult)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected previousResult.Result to be type %T, but got %T",
			errInvalidRequestType,
			prevState,
			previousResult.Result,
		)
	}

	return prevState, nil
}
