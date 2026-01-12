package stripe

import (
	"context"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
)

// UpdateSubscription updates an existing subscription by comparing the previous
// subscription state with the new desired state.
// It merges events: keeps events for objects not in params, adds/updates events for objects in params.
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Validate inputs
	prevState, err := validatePreviousResult(previousResult)
	if err != nil {
		return nil, err
	}

	existingEndpoint, err := getExistingEndpoint(prevState.Subscriptions)
	if err != nil {
		return nil, err
	}

	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Build merged event set
	mergedEventNames, err := buildMergedEventNames(prevState, params)
	if err != nil {
		return nil, err
	}

	// Update webhook endpoint
	payload := &WebhookPayload{
		URL:           req.WebhookEndPoint,
		EnabledEvents: mergedEventNames,
	}

	response, err := c.updateWebhookEndpoint(ctx, existingEndpoint.ID, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	// Build result with merged subscription events
	mergedSubscriptionEvents := buildMergedSubscriptionEvents(previousResult, params)

	result, err := buildSubscriptionResult(response, mergedSubscriptionEvents)
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

// getExistingEndpoint extracts the real endpoint ID from subscriptions.
// Since IDs are stored as "endpointID:objectName", we extract the base endpoint ID.
func getExistingEndpoint(subscriptions map[common.ObjectName]WebhookResponse) (WebhookResponse, error) {
	if len(subscriptions) == 0 {
		return WebhookResponse{}, fmt.Errorf("%w: no existing subscriptions", errMissingParams)
	}

	for _, endpoint := range subscriptions {
		realEndpointID := extractBaseEndpointID(endpoint.ID)

		result := endpoint
		result.ID = realEndpointID

		return result, nil
	}

	return WebhookResponse{}, fmt.Errorf("%w: unable to extract existing endpoint", errMissingParams)
}

// buildMergedEventNames builds a merged event list by keeping events from objects not being updated
// and adding requested events for objects being updated.
func buildMergedEventNames(
	prevState *SubscriptionResult,
	params common.SubscribeParams,
) ([]string, error) {
	mergedEventsSet := make(map[string]bool)

	// Add all existing events from objects not being updated
	for obj, endpoint := range prevState.Subscriptions {
		if _, isBeingUpdated := params.SubscriptionEvents[obj]; !isBeingUpdated {
			for _, event := range endpoint.EnabledEvents {
				mergedEventsSet[event] = true
			}
		}
	}

	// Add requested events for objects being updated
	requestedEventsSet, err := buildRequestedEventSet(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	for event := range requestedEventsSet {
		mergedEventsSet[event] = true
	}

	if len(mergedEventsSet) == 0 {
		return nil, fmt.Errorf("%w: no events to subscribe to after merge", errMissingParams)
	}

	mergedEventNames := make([]string, 0, len(mergedEventsSet))
	for event := range mergedEventsSet {
		mergedEventNames = append(mergedEventNames, event)
	}

	return mergedEventNames, nil
}

// buildMergedSubscriptionEvents builds a merged subscription events map by keeping previous objects
// not being updated and adding new/updated objects from params.
func buildMergedSubscriptionEvents(
	previousResult *common.SubscriptionResult,
	params common.SubscribeParams,
) map[common.ObjectName]common.ObjectEvents {
	mergedSubscriptionEvents := make(map[common.ObjectName]common.ObjectEvents)

	// Keep previous objects not being updated - use ObjectEvents from previousResult if available
	if previousResult != nil && previousResult.ObjectEvents != nil {
		for obj, events := range previousResult.ObjectEvents {
			if _, isBeingUpdated := params.SubscriptionEvents[obj]; !isBeingUpdated {
				mergedSubscriptionEvents[obj] = events
			}
		}
	}

	// Add new/updated objects
	maps.Copy(mergedSubscriptionEvents, params.SubscriptionEvents)

	return mergedSubscriptionEvents
}
