package outreach

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/go-playground/validator"
)

// nolint: funlen
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	var successfulSubscriptions []SuccessfulSubscription
	var firstError error
	var errorOnce sync.Once
	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0)

	// Process all object+event combinations
	for obj, events := range params.SubscriptionEvents {
		for _, event := range events.Events {

			currObj := obj
			currentEvent := event

			callbacks = append(callbacks, func(ctx context.Context) error {
				successful, failed := c.createSingleSubscription(ctx, currentEvent, currObj, req)
				mutex.Lock()
				defer mutex.Unlock()

				if failed != nil {
					errorOnce.Do(func() {
						firstError = failed
					})
				} else {
					successfulSubscriptions = append(successfulSubscriptions, *successful)
				}

				return nil
			})

		}
	}

	err = simultaneously.DoCtx(ctx, -1, callbacks...)
	if err != nil {
		return nil, fmt.Errorf("failed to process subscriptions: %w", err)
	}

	res := &common.SubscriptionResult{
		ObjectEvents: params.SubscriptionEvents,
	}

	if firstError != nil {
		rollbackErr := c.rollbackSubscriptions(ctx, successfulSubscriptions)
		if rollbackErr != nil {

			res.Status = common.SubscriptionStatusFailedToRollback
			res.Result = rollbackErr

			return res, errors.Join(firstError, rollbackErr)
		}
		res.Status = common.SubscriptionStatusFailed
		res.ObjectEvents = nil

		return res, firstError
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = &SubscriptionResultData{
		SuccessfulSubscriptions: successfulSubscriptions,
	}

	return res, nil
}

func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be null", errMissingParams) //nolint:err113,lll
	}

	subscriptionData, ok := result.Result.(*SubscriptionResultData)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T", errInvalidRequestType, subscriptionData, result.Result) //nolint:err113,lll
	}

	if len(subscriptionData.SuccessfulSubscriptions) == 0 {
		return fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	for _, subscription := range subscriptionData.SuccessfulSubscriptions {
		err := c.deleteSubscription(ctx, subscription.ID)
		if err != nil {
			return fmt.Errorf("failed to delete subscription with ID %s: %w", subscription.ID, err)
		}
	}

	return nil
}

// createSingleSubscription attempts to create a single subscription and returns either a successful or failed result.
func (c *Connector) createSingleSubscription(
	ctx context.Context,
	event common.SubscriptionEventType,
	obj common.ObjectName,
	req *SubscriptionRequest,
) (*SuccessfulSubscription, error) {
	payload, err := buildPayload(event, obj, req.WebhookEndPoint, req.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %v", obj, event, err)
	}

	result, err := c.createSubscriptions(ctx, payload, c.Client.Post)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %v", obj, event, err)
	}

	return &SuccessfulSubscription{
		ID:         result.Data.ID,
		ObjectName: string(obj),
		EventName:  string(event),
	}, nil
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

	if validate.Struct(req) != nil {
		return nil, fmt.Errorf("%w: request is invalid", errInvalidRequestType)
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
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) rollbackSubscriptions(
	ctx context.Context,
	subscriptions []SuccessfulSubscription,
) error {
	var rollbackErrors error
	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0, len(subscriptions))
	for _, sub := range subscriptions {
		sub := sub
		callbacks = append(callbacks, func(ctx context.Context) error {
			err := c.deleteSubscription(ctx, sub.ID)
			if err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				rollbackErrors = errors.Join(rollbackErrors, fmt.Errorf("failed to rollback subscription %s (%s:%s): %w",
					sub.ID, sub.ObjectName, sub.EventName, err))
			}
			return nil
		})
	}

	err := simultaneously.DoCtx(ctx, -1, callbacks...)
	if err != nil {
		rollbackErrors = errors.Join(rollbackErrors, fmt.Errorf("failed to rollback subscriptions: %w", err))
	}

	return rollbackErrors
}

func getProviderEventName(subscriptionEvent common.SubscriptionEventType) (ModuleEvent, error) {
	switch subscriptionEvent { //nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return Created, nil
	case common.SubscriptionEventTypeUpdate:
		return Updated, nil
	case common.SubscriptionEventTypeDelete:
		return Destroyed, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedEventType, subscriptionEvent)
	}
}

func buildPayload(
	event common.SubscriptionEventType,
	objectName common.ObjectName,
	webhookURL string,
	secret string,
) (*SubscriptionPayload, error) {
	Event, err := getProviderEventName(event)
	if err != nil {
		return nil, err
	}

	payload := &SubscriptionPayload{
		Data: SubscriptionData{
			Type: "webhook",
			Attributes: AttributesPayload{
				Action:   string(Event),
				Resource: string(objectName),
				URL:      webhookURL,
				Secret:   secret,
			},
		},
	}

	return payload, nil
}
