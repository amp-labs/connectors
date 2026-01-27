package stripe

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// DeleteSubscription deletes webhook subscriptions for the specified objects.
// Extracts the real endpoint ID from composite IDs (format: "endpointID:objectName").
// If only some objects are deleted, the endpoint is updated to remove their events.
// If all objects are deleted, the entire endpoint is deleted.
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	subscriptionData, err := validateSubscriptionResult(result)
	if err != nil {
		return err
	}

	endpointInfo, err := extractEndpointInfo(subscriptionData)
	if err != nil {
		return err
	}

	return c.deleteWebhookEndpoint(ctx, endpointInfo.ID)
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

// endpointInfo holds information about the endpoint to delete from.
type endpointInfo struct {
	ID              string
	ObjectsToDelete map[common.ObjectName]bool
}

// extractEndpointInfo extracts the real endpoint ID and objects to delete from composite IDs.
func extractEndpointInfo(subscriptionData *SubscriptionResult) (*endpointInfo, error) {
	endpointIDs := make(map[string]bool)
	objectsToDelete := make(map[common.ObjectName]bool)

	var realEndpointID string

	for obj, response := range subscriptionData.Subscriptions {
		compositeID := response.ID
		baseID := extractBaseEndpointID(compositeID)
		endpointIDs[baseID] = true

		if realEndpointID == "" {
			realEndpointID = baseID
		}

		objectsToDelete[obj] = true
	}

	if len(endpointIDs) != 1 {
		return nil, fmt.Errorf(
			"%w: expected all subscriptions to share the same endpoint ID, but found %d different IDs: %v",
			errInvalidRequestType,
			len(endpointIDs),
			endpointIDs,
		)
	}

	if realEndpointID == "" {
		return nil, fmt.Errorf("%w: endpoint ID is empty", errMissingParams)
	}

	return &endpointInfo{
		ID:              realEndpointID,
		ObjectsToDelete: objectsToDelete,
	}, nil
}
