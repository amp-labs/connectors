package attio

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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

// Subscribe handles two different subscription patterns based on object type:
//
// PATTERN 1 - Core Objects (lists, tasks, notes, workspace_members):
//   - Subscribes using specific event types (e.g., "task.created", "note.updated")
//   - No filters required - events are already object-specific
//   - Event mappings are predefined in attioObjectEvents
//
// PATTERN 2 - Standard/Custom Objects (people, companies, deals, etc.):
//   - Subscribes using generic "record.*" events (record.created, record.updated, record.deleted)
//   - Requires object_id filter to target the specific object type
//   - Object metadata is fetched dynamically from Attio API (Objects can be activated/deactivated in the workspace settings)
//
// nolint: funlen, cyclop, godoclint
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Fetch the current list of objects from Attio API
	objectList, err := c.geStandardOrCustomObjectsList(ctx)
	if err != nil {
		return nil, err
	}

	// Build a map: object name -> object ID
	// Example: "people" -> "0e80364d-70b1-44d3-b7ba-0a6a564a7152"
	objectIDMap := make(map[common.ObjectName]string)
	for _, obj := range objectList {
		objectIDMap[common.ObjectName(obj.ApiSlug)] = obj.Id.ObjectId
	}

	// Validate that requested events are supported
	err = validateSubscriptionEvents(params.SubscriptionEvents, objectIDMap)
	if err != nil {
		return nil, err
	}

	payload, err := buildPayload(params.SubscriptionEvents, objectIDMap, req.WebhookEndpoint)
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
	objectIDMap map[common.ObjectName]string,
	webhookURL string,
) (*subscriptionPayload, error) {
	subscriptions := make([]subscription, 0)

	for objectName, events := range subscriptionEvents {
		for _, event := range events.Events {
			EventsMap, isCoreObject := getObjectEvents(objectName)
			if isCoreObject {
				providerEvents := EventsMap.toProviderEvents(event)

				if len(providerEvents) == 0 {
					return nil, fmt.Errorf("%w: no provider events for object '%s' and event '%s'",
						errUnsupportedSubscriptionEvent, objectName, event)
				}

				for _, e := range providerEvents {
					subscriptions = append(subscriptions, subscription{
						EventType: e,
						Filter:    nil,
					})
				}
			} else {
				objectId, exists := objectIDMap[objectName]
				if !exists {
					return nil, fmt.Errorf("object '%s' not supported or not activated in workspace", objectName)
				}

				providerEvent, err := toRecordEvents(event)
				if err != nil {
					return nil, err
				}

				filter := map[string]any{
					"$and": []map[string]any{
						{
							"field":    "id.object_id",
							"operator": "equals",
							"value":    objectId,
						},
					},
				}
				subscriptions = append(subscriptions,
					subscription{
						EventType: providerEvent,
						Filter:    filter,
					},
				)
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

func getObjectEvents(objectName common.ObjectName) (objectEvents, bool) {
	events, exists := attioObjectEvents[objectName]
	if !exists {
		return objectEvents{}, false
	}

	return events, true
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

// toRecordEvents converts a common event type to generic Attio record events.
// This is used for standard/custom objects (people, companies, deals, etc.) which use
// generic "record.*" event types instead of object-specific events.
func toRecordEvents(commonEvent common.SubscriptionEventType) (providerEvent, error) {
	switch commonEvent {
	case common.SubscriptionEventTypeCreate:
		return "record.created", nil

	case common.SubscriptionEventTypeUpdate:
		return "record.updated", nil

	case common.SubscriptionEventTypeDelete:
		return "record.updated", nil

	default:
		return "", fmt.Errorf("unsupported event type %s", commonEvent)
	}
}
