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
	subscriptionData, err := validateSubscriptionResult(result)
	if err != nil {
		return err
	}

	endpointID, err := extractEndpointID(subscriptionData)
	if err != nil {
		return err
	}

	return c.deleteWebhookEndpoint(ctx, endpointID)
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

// extractEndpointID extracts the real endpoint ID shared by all composite IDs.
func extractEndpointID(subscriptionData *SubscriptionResult) (string, error) {
	endpointIDs := make(map[string]bool)

	var realEndpointID string

	for _, response := range subscriptionData.Subscriptions {
		compositeID := response.ID
		baseID := extractBaseEndpointID(compositeID)
		endpointIDs[baseID] = true

		if realEndpointID == "" {
			realEndpointID = baseID
		}
	}

	if len(endpointIDs) != 1 {
		return "", fmt.Errorf(
			"%w: expected all subscriptions to share the same endpoint ID, but found %d different IDs: %v",
			errInvalidRequestType,
			len(endpointIDs),
			endpointIDs,
		)
	}

	if realEndpointID == "" {
		return "", fmt.Errorf("%w: endpoint ID is empty", errMissingParams)
	}

	return realEndpointID, nil
}
