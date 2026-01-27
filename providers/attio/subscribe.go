package attio

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/go-playground/validator"
)

var _ connectors.SubscribeConnector = &Connector{}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &subscriptionResult{},
	}
}

// nolint: funlen, cyclop, godoclint
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Validate that requested events are supported
	err = validateSubscriptionEvents(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	payload, err := buildPayload(params.SubscriptionEvents, req.WebhookEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to build subscription payload: %w", err)
	}

	res := &common.SubscriptionResult{
		ObjectEvents: params.SubscriptionEvents,
	}

	response, err := c.createSubscriptions(ctx, payload, c.Client.Post)
	if err != nil {
		res.Status = common.SubscriptionStatusFailed
		res.Events = nil

		return res, fmt.Errorf("failed to create subscriptions: %w", err)
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = &subscriptionResult{
		Data: response.Data,
	}

	return res, nil
}

// UpdateSubscription implements [connectors.SubscribeConnector].
// nolint: nilnil
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// TODO: Implement update logic
	return nil, nil
}

func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be nil", errMissingParams)
	}

	subscriptionData, ok := result.Result.(*subscriptionResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType, subscriptionData, result.Result)
	}

	if len(subscriptionData.Data.Subscriptions) == 0 {
		return fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	err := c.deleteSubscription(ctx, subscriptionData.Data.ID.WebhookID)
	if err != nil {
		return fmt.Errorf(
			"failed to delete subscription (ID: %s): %w",
			subscriptionData.Data.ID.WebhookID,
			err,
		)
	}

	return nil
}

func validateRequest(params common.SubscribeParams) (*subscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*subscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T' got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: request is invalid: %w", errInvalidRequestType, err)
	}

	return req, nil
}

func (c *Connector) getSubscribeURL() (*urlbuilder.URL, error) {
	return c.getApiURL("webhooks")
}

func (c *Connector) createSubscriptions(ctx context.Context,
	payload *subscriptionPayload,
	updater common.WriteMethod,
) (*createSubscriptionsResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	resp, err := updater(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[createSubscriptionsResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return result, nil
}

func (c *Connector) deleteSubscription(ctx context.Context, subscriptionID string) error {
	url, err := c.getSubscribeURL()
	if err != nil {
		return err
	}

	url.AddPath(subscriptionID)

	_, err = c.Client.Delete(ctx, url.String())

	return err
}

func buildPayload(
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
	webhookURL string,
) (*subscriptionPayload, error) {
	subscriptions := make([]subscription, 0)

	for objectName, events := range subscriptionEvents {
		for _, event := range events.Events {
			EventsMap, err := getObjectEvents(objectName)
			// This should never happen due to prior validation
			if err != nil {
				return nil, err
			}

			providerEvents := EventsMap.toProviderEvents(event)

			if len(providerEvents) == 0 {
				return nil, fmt.Errorf("%w: no provider events for object '%s' and event '%s'",
					errUnsupportedSubscriptionEvent, objectName, event)
			}

			for _, e := range providerEvents {
				subscriptions = append(subscriptions, subscription{
					EventType: e,
					// Filter is an object used to limit which webhook events are delivered.
					// Filters can target specific records (by list_id, entry_id) and specific
					// It cannot be used to do field level filtering.
					// Use null to receive all events without filtering.
					// Ref: https://docs.attio.com/rest-api/guides/webhooks#filtering
					Filter: nil,
				})
			}
		}
	}

	payload := &subscriptionPayload{
		Data: subscriptionData{
			TargetURL:     webhookURL,
			Subscriptions: subscriptions,
		},
	}

	return payload, nil
}

func getObjectEvents(objectName common.ObjectName) (objectEvents, error) {
	events, exists := attioObjectEvents[objectName]
	if !exists {
		return objectEvents{}, fmt.Errorf("%w: %s", errUnsupportedObject, objectName)
	}

	return events, nil
}

func validateSubscriptionEvents(subscriptionEvents map[common.ObjectName]common.ObjectEvents) error {
	var validationErrors error

	for objectName, objectEvents := range subscriptionEvents {
		attioEvents, exist := attioObjectEvents[objectName]
		if !exist {
			validationErrors = errors.Join(validationErrors,
				fmt.Errorf("%s %w", objectName, errUnsupportedObject))

			continue
		}

		// Get all supported events for this object
		supportedEvents := attioEvents.getAllSupportedEvents()

		supportedSet := make(map[providerEvent]bool)
		for _, evt := range supportedEvents {
			supportedSet[evt] = true
		}

		for _, event := range objectEvents.Events {
			providerEvents := attioEvents.toProviderEvents(event)

			if providerEvents == nil {
				validationErrors = errors.Join(validationErrors,
					fmt.Errorf("%w for object '%s'", errUnsupportedSubscriptionEvent, objectName))

				continue
			}

			if len(providerEvents) == 0 {
				validationErrors = errors.Join(validationErrors,
					fmt.Errorf("%w: event '%s' for object '%s'", errUnsupportedSubscriptionEvent, event, objectName))

				continue
			}

			// Validate that all provider events are supported
			for _, providerEvent := range providerEvents {
				if !supportedSet[providerEvent] {
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("%w: provider event '%s' for common event '%s' and object '%s'",
							errUnsupportedSubscriptionEvent, providerEvent, event, objectName))
				}
			}
		}
	}

	return validationErrors
}

// getAllSupportedEvents returns all provider events that this object supports.
func (e objectEvents) getAllSupportedEvents() []providerEvent {
	var events []providerEvent

	events = append(events, e.createEvents...)
	events = append(events, e.updateEvents...)
	events = append(events, e.deleteEvents...)

	return events
}

// getProviderEvents converts a common event type to corresponding Attio provider events.
func (e objectEvents) toProviderEvents(commonEvent common.SubscriptionEventType) []providerEvent {
	switch commonEvent { // nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return e.createEvents
	case common.SubscriptionEventTypeUpdate:
		return e.updateEvents
	case common.SubscriptionEventTypeDelete:
		return e.deleteEvents
	default:
		return nil
	}
}
