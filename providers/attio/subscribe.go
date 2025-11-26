package attio

import (
	"context"
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
		Result: &SubscriptionResult{},
	}
}

// nolint: funlen, cyclop
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	payload, err := buildPayload(params.SubscriptionEvents, req.WebhookEndPoint)
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
	res.Result = &SubscriptionResult{
		Data: response.Data,
	}

	return res, nil
}

// UpdateSubscription implements connectors.SubscribeConnector.
func (c *Connector) UpdateSubscription(ctx context.Context,
	params common.SubscribeParams, previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Validate the previous result
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
		) //nolint:lll
	}

	// Delete the existing subscription
	err := c.deleteSubscription(ctx, prevState.Data.ID.WebhookID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to delete previous subscription (ID: %s): %w",
			prevState.Data.ID.WebhookID,
			err,
		)
	}

	// Create a new subscription
	newResult, err := c.Subscribe(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create new subscription: %w", err)
	}

	return newResult, nil
}

func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be null", errMissingParams) //nolint:err113,lll
	}

	subscriptionData, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T", errInvalidRequestType, subscriptionData, result.Result) //nolint:err113,lll
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

func validateRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
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
	url, err := c.getApiURL("webhooks")
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (c *Connector) createSubscriptions(ctx context.Context,
	payload *SubscriptionPayload,
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

func getProviderEventName(subscriptionEvent common.SubscriptionEventType) (ModuleEvent, error) {
	switch subscriptionEvent { //nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return Created, nil
	case common.SubscriptionEventTypeUpdate:
		return Updated, nil
	case common.SubscriptionEventTypeDelete:
		return Deleted, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedEventType, subscriptionEvent)
	}
}

func buildPayload(
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
	webhookURL string,
) (*SubscriptionPayload, error) {
	subscriptions := make([]Subscription, 0)

	for objectName, events := range subscriptionEvents {
		for _, event := range events.Events {
			Event, err := getProviderEventName(event)
			if err != nil {
				return nil, err
			}

			subscriptionObjectName := readObjectNameToSubscriptionName.Get(string(objectName))

			providerventType := subscriptionObjectName + "." + string(Event)

			subscriptions = append(subscriptions, Subscription{
				EventType: providerventType,
				// Filter is an object used to limit which webhook events are delivered.
				// Filters can target specific records (by list_id, entry_id) and specific
				// It cannot be used to do field level filtering.
				// Use null to receive all events without filtering.
				// Ref: https://docs.attio.com/rest-api/guides/webhooks#filtering
				Filter: nil,
			})
		}
	}

	payload := &SubscriptionPayload{
		Data: SubscriptionData{
			TargetURL:     webhookURL,
			Subscriptions: subscriptions,
		},
	}

	return payload, nil
}
