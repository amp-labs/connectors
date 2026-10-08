package stripe

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// DeleteSubscription removes the shared webhook endpoint carrying all of the
// installation's object subscriptions.
// Extracts the real endpoint ID from composite IDs (format: "endpointID:objectName").
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	subscriptionResult, err := validateSubscriptionResult(result)
	if err != nil {
		return err
	}

	return c.deleteWebhookEndpoint(ctx, subscriptionResult.WebhookId)
}

// validateSubscriptionResult validates the subscription result and extracts the subscription data.
func validateSubscriptionResult(result common.SubscriptionResult) (*SubscriptionResult, error) {
	if result.Result == nil {
		return nil, fmt.Errorf("%w: Result cannot be null", errMissingParams)
	}

	subscriptionData, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType,
			subscriptionData,
			result.Result,
		)
	}

	if len(subscriptionData.Subscriptions) == 0 {
		return nil, fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	return subscriptionData, nil
}
