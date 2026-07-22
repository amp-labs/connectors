package attio

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

var (
	_ connectors.SubscribeConnector                   = &Connector{}
	_ connectors.SubscriptionEventObjectNameConnector = &Connector{}
)

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscriptionResult{},
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
//   - Object metadata is fetched dynamically from Attio API
//   - Objects can be activated/deactivated in the workspace settings
//   - Ref: https://docs.attio.com/rest-api/endpoint-reference/webhooks/create-a-webhook
//
// nolint: funlen, cyclop, godoclint
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	payload, err := c.buildSubscribePayload(ctx, params)
	if err != nil {
		return nil, err
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
	res.Result = &SubscriptionResult{
		Data: response.Data,
	}

	return res, nil
}

// buildSubscribePayload fetches the workspace's objects, validates the requested events, and builds
// the Attio webhook payload. Shared by Subscribe and UpdateSubscription.
func (c *Connector) buildSubscribePayload(
	ctx context.Context,
	params common.SubscribeParams,
) (*subscriptionPayload, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Fetch the current list of objects from Attio API, then build a map: object name -> object ID
	// Example: "people" -> "0e80364d-70b1-44d3-b7ba-0a6a564a7152"
	objectList, err := c.readStandardOrCustomObjectsList(ctx)
	if err != nil {
		return nil, err
	}

	standardObjects := make(map[common.ObjectName]string)
	for _, obj := range objectList {
		standardObjects[common.ObjectName(obj.ApiSlug)] = obj.Id.ObjectId
	}

	// Validate that requested events are supported
	if err = validateSubscriptionEvents(params.SubscriptionEvents, standardObjects); err != nil {
		return nil, err
	}

	payload, err := buildPayload(params.SubscriptionEvents, standardObjects, req.WebhookEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to build subscription payload: %w", err)
	}

	return payload, nil
}

// nolint: nilnil, godoclint
// UpdateSubscription updates an existing Attio webhook in place: it rebuilds the subscription
// payload from params and PATCHes the previously-created webhook, preserving its id (and therefore
// its signing secret).
// Reference: https://docs.attio.com/rest-api/endpoint-reference/webhooks/update-a-webhook
// nolint: godoclint
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	if previousResult == nil || previousResult.Result == nil {
		return nil, fmt.Errorf("%w: previous subscription result cannot be nil", errMissingParams)
	}

	prev, ok := previousResult.Result.(*SubscriptionResult)
	if !ok {
		return nil, fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType, prev, previousResult.Result)
	}

	webhookID := prev.Data.Id.WebhookId
	if webhookID == "" {
		return nil, fmt.Errorf("%w: previous subscription is missing a webhook id", errMissingParams)
	}

	payload, err := c.buildSubscribePayload(ctx, params)
	if err != nil {
		return nil, err
	}

	res := &common.SubscriptionResult{
		ObjectEvents: params.SubscriptionEvents,
	}

	response, err := c.updateSubscriptions(ctx, webhookID, payload)
	if err != nil {
		res.Status = common.SubscriptionStatusFailed

		return res, fmt.Errorf("failed to update subscriptions: %w", err)
	}

	// The webhook id is unchanged, so its signing secret is unchanged. Attio's update response may
	// omit the secret, so preserve the previously stored one when it is absent.
	if response.Data.Secret == "" {
		response.Data.Secret = prev.Data.Secret
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = &SubscriptionResult{
		Data: response.Data,
	}

	return res, nil
}

// DeleteSubscription removes an existing webhook subscription in Attio.
// Reference: https://docs.attio.com/rest-api/endpoint-reference/webhooks/delete-a-webhook
// nolint: godoclint
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be nil", errMissingParams)
	}

	subscriptionData, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType, subscriptionData, result.Result)
	}

	if len(subscriptionData.Data.Subscriptions) == 0 {
		return fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	err := c.deleteSubscription(ctx, subscriptionData.Data.Id.WebhookId)
	if err != nil {
		return fmt.Errorf(
			"failed to delete subscription (id: %s): %w",
			subscriptionData.Data.Id.WebhookId,
			err,
		)
	}

	return nil
}

func (c *Connector) createSubscriptions(ctx context.Context,
	payload *subscriptionPayload,
	updater common.WriteMethod,
) (*CreateSubscriptionsResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	resp, err := updater(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[CreateSubscriptionsResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return result, nil
}

// updateSubscriptions PATCHes an existing webhook (identified by webhookID) with a new set of
// subscriptions, replacing the webhook's subscriptions while preserving its id.
// Reference: https://docs.attio.com/rest-api/endpoint-reference/webhooks/update-a-webhook
func (c *Connector) updateSubscriptions(
	ctx context.Context,
	webhookID string,
	payload *subscriptionPayload,
) (*CreateSubscriptionsResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	url.AddPath(webhookID)

	resp, err := c.Client.Patch(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[CreateSubscriptionsResponse](resp)
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
	standardObjects map[common.ObjectName]string,
	webhookURL string,
) (*subscriptionPayload, error) {
	subscriptions := make([]Subscription, 0)

	for objectName, events := range subscriptionEvents {
		for _, event := range events.Events {
			objectEvents, isCoreObject := getObjectEvents(objectName)

			if isCoreObject {
				// Handle building subscriptions for core objects
				subs, err := buildSubscriptionPayloadForCoreObj(objectEvents, objectName, event)
				if err != nil {
					return nil, err
				}

				subscriptions = append(subscriptions, subs...)
			} else {
				// Handle building subscriptions for standard/custom objects
				subs, err := buildSubscriptionPayloadForStandardObj(standardObjects, objectName, event)
				if err != nil {
					return nil, err
				}

				subscriptions = append(subscriptions, subs...)
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
