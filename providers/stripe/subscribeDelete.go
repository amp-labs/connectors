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

	currentEndpoint, err := c.GetWebhookEndpoint(ctx, endpointInfo.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch webhook endpoint (ID: %s): %w", endpointInfo.ID, err)
	}

	eventsToRemove := collectEventsToRemove(subscriptionData, endpointInfo.ObjectsToDelete)
	eventsToKeep := filterEventsToKeep(currentEndpoint.EnabledEvents, eventsToRemove)

	// If no events remain, delete the entire endpoint
	if len(eventsToKeep) == 0 {
		return c.deleteWebhookEndpoint(ctx, endpointInfo.ID)
	}

	// Otherwise, update the endpoint to remove events for deleted objects
	return c.updateEndpointAfterPartialDelete(ctx, endpointInfo.ID, eventsToKeep, currentEndpoint, subscriptionData)
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

// collectEventsToRemove collects all events that should be removed for the objects being deleted.
func collectEventsToRemove(
	subscriptionData *SubscriptionResult,
	objectsToDelete map[common.ObjectName]bool,
) map[string]bool {
	eventsToRemove := make(map[string]bool)

	for obj := range objectsToDelete {
		if endpoint, ok := subscriptionData.Subscriptions[obj]; ok {
			for _, event := range endpoint.EnabledEvents {
				eventsToRemove[event] = true
			}
		}
	}

	return eventsToRemove
}

// filterEventsToKeep filters out events that should be removed, keeping only the remaining events.
func filterEventsToKeep(currentEvents []string, eventsToRemove map[string]bool) []string {
	eventsToKeep := make([]string, 0)

	for _, event := range currentEvents {
		if !eventsToRemove[event] {
			eventsToKeep = append(eventsToKeep, event)
		}
	}

	return eventsToKeep
}

// updateEndpointAfterPartialDelete updates the endpoint after a partial delete operation.
func (c *Connector) updateEndpointAfterPartialDelete(
	ctx context.Context,
	endpointID string,
	eventsToKeep []string,
	currentEndpoint *WebhookEndpointResponse,
	subscriptionData *SubscriptionResult,
) error {
	webhookURL := getWebhookURL(currentEndpoint, subscriptionData)
	if webhookURL == "" {
		return fmt.Errorf("%w: webhook URL is required for partial delete", errMissingParams)
	}

	payload := &WebhookEndpointPayload{
		URL:     webhookURL,
		Enabled: eventsToKeep,
	}

	_, err := c.updateWebhookEndpoint(ctx, endpointID, payload)
	if err != nil {
		return fmt.Errorf("failed to update webhook endpoint after partial delete (ID: %s): %w", endpointID, err)
	}

	return nil
}

// getWebhookURL extracts the webhook URL from the current endpoint.
func getWebhookURL(
	currentEndpoint *WebhookEndpointResponse,
	subscriptionData *SubscriptionResult,
) string {
	if currentEndpoint.URL != "" {
		return currentEndpoint.URL
	}

	for _, endpoint := range subscriptionData.Subscriptions {
		if endpoint.URL != "" {
			return endpoint.URL
		}
	}

	return ""
}
