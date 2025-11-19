package salesloft

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/simultaneously"
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

	// Store successful subscriptions with their full response data
	subscriptionsMap := make(map[common.ObjectName]map[SalesloftEventType]SubscriptionResponse)
	successfulSubscriptions := make([]SuccessfulSubscription, 0)

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
				response, failErr := c.createSingleSubscription(ctx, currentEvent, currObj, req)

				mutex.Lock()
				defer mutex.Unlock()

				if failErr != nil {
					errorOnce.Do(func() {
						firstError = failErr
					})
				} else {
					// Convert common event type to Salesloft event type format
					salesloftEventType, err := buildSalesloftEventType(currObj, currentEvent)
					if err != nil {
						errorOnce.Do(func() {
							firstError = fmt.Errorf("failed to convert event type %s for object %s: %w", currentEvent, currObj, err)
						})
						return nil
					}

					// Initialize nested map if needed
					if subscriptionsMap[currObj] == nil {
						subscriptionsMap[currObj] = make(map[SalesloftEventType]SubscriptionResponse)
					}

					subscriptionsMap[currObj][salesloftEventType] = *response

					// Keep track of successful subscriptions for rollback
					successfulSubscriptions = append(successfulSubscriptions, SuccessfulSubscription{
						ID:         strconv.Itoa(response.ID),
						ObjectName: string(currObj),
						EventName:  string(salesloftEventType),
					})
				}

				return nil
			})
		}
	}

	res := &common.SubscriptionResult{
		ObjectEvents: params.SubscriptionEvents,
	}

	err = simultaneously.DoCtx(ctx, -1, callbacks...)
	if err != nil {
		return nil, fmt.Errorf("failed to process subscriptions: %w", err)
	}

	objectEvents := make(map[common.ObjectName]common.ObjectEvents)

	if firstError != nil {
		_, failedToRollBack, rollbackErr := c.rollbackSubscriptions(ctx, successfulSubscriptions)
		if rollbackErr != nil {
			res.Status = common.SubscriptionStatusFailedToRollback

			for _, failedSub := range failedToRollBack {
				objectEvents[common.ObjectName(failedSub.ObjectName)] = common.ObjectEvents{
					Events: []common.SubscriptionEventType{common.SubscriptionEventType(failedSub.EventName)},
				}
			}

			res.ObjectEvents = objectEvents
			return res, errors.Join(firstError, rollbackErr)
		}

		res.Status = common.SubscriptionStatusFailed
		res.ObjectEvents = nil
		return res, firstError
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = &SubscriptionResult{
		Subscriptions: subscriptionsMap,
	}

	return res, nil
}

// parseSalesloftEventType splits a Salesloft event type back into object name and action
// Example: parseSalesloftEventType("person_created") -> ("person", "created", nil)
func parseSalesloftEventType(eventType SalesloftEventType) (string, EventAction, error) {
	parts := strings.Split(string(eventType), "_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: invalid format '%s', expected 'objectName_action'", errInvalidModuleEvent, eventType)
	}

	objectName := parts[0]
	action := EventAction(parts[1])

	// Validate the action part
	switch action {
	case ActionCreated, ActionUpdated, ActionDestroyed:
		return objectName, action, nil
	default:
		return "", "", fmt.Errorf("%w: unknown action '%s'", errUnsupportedEventType, action)
	}
}

func validateRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if validate.Struct(req) != nil {
		return nil, fmt.Errorf("%w: request is invalid", errInvalidRequestType)
	}

	return req, nil
}

// createSingleSubscription attempts to create a single subscription and returns the full response
func (c *Connector) createSingleSubscription(
	ctx context.Context,
	event common.SubscriptionEventType,
	obj common.ObjectName,
	req *SubscriptionRequest,
) (*SubscriptionResponse, error) {
	payload, err := buildSubscriptionPayload(event, obj, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription payload for object %s, event %s: %w", obj, event, err)
	}

	result, err := c.createSubscription(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %w", obj, event, err)
	}

	return result, nil
}

func buildSubscriptionPayload(
	event common.SubscriptionEventType,
	objectName common.ObjectName,
	req *SubscriptionRequest,
) (*SubscriptionPayload, error) {
	salesloftEventType, err := buildSalesloftEventType(objectName, event)
	if err != nil {
		return nil, err
	}

	payload := &SubscriptionPayload{
		CallbackURL:   req.WebhookEndPoint,
		EventType:     string(salesloftEventType),
		CallbackToken: req.Secret,
	}

	return payload, nil
}

// createSubscription makes the API call to create a webhook subscription
func (c *Connector) createSubscription(
	ctx context.Context,
	payload *SubscriptionPayload,
) (*SubscriptionResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[SubscriptionResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return result, nil
}

// DeleteSubscription deletes webhook subscriptions
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be null", errMissingParams)
	}

	subscriptionData, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return fmt.Errorf("%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType, subscriptionData, result.Result)
	}

	if len(subscriptionData.Subscriptions) == 0 {
		return fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	// Extract subscription IDs from the nested map and delete them
	for objName, eventsMap := range subscriptionData.Subscriptions {
		for eventType, response := range eventsMap {
			err := c.deleteSubscription(ctx, strconv.Itoa(response.ID))
			if err != nil {
				return fmt.Errorf(
					"failed to delete subscription for object %s, event %s (ID: %d): %w",
					objName,
					eventType,
					response.ID,
					err,
				)
			}
		}
	}

	return nil
}

func (c *Connector) getSubscribeURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, ApiVersion, "webhook_subscriptions")
}

// deleteSubscription deletes a single subscription by ID
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

// rollbackSubscriptions attempts to delete all successful subscriptions in case of partial failure
func (c *Connector) rollbackSubscriptions(
	ctx context.Context,
	subscriptions []SuccessfulSubscription,
) (rolledBack []SuccessfulSubscription, failedToRollBack []SuccessfulSubscription, err error) {
	var rollbackErrors error

	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0, len(subscriptions))

	for _, subFromList := range subscriptions {
		callbacks = append(callbacks,
			func(sub SuccessfulSubscription) func(ctx context.Context) error {
				return func(ctx context.Context) error {
					deleteErr := c.deleteSubscription(ctx, sub.ID)

					mutex.Lock()
					defer mutex.Unlock()

					if deleteErr != nil {
						failedToRollBack = append(failedToRollBack, sub)
						rollbackErrors = errors.Join(rollbackErrors, fmt.Errorf("failed to rollback subscription %s (%s:%s): %w",
							sub.ID, sub.ObjectName, sub.EventName, deleteErr))
					} else {
						rolledBack = append(rolledBack, sub)
					}
					return nil
				}

			}(subFromList),
		)
	}

	err = simultaneously.DoCtx(ctx, -1, callbacks...)
	if err != nil {
		rollbackErrors = errors.Join(rollbackErrors, fmt.Errorf("failed to rollback subscriptions: %w", err))
	}

	return rolledBack, failedToRollBack, rollbackErrors
}

// buildSalesloftEventType combines object name and event action into Salesloft's expected format.
// Example: account + create -> account_created
func buildSalesloftEventType(objectName common.ObjectName, eventType common.SubscriptionEventType) (SalesloftEventType, error) {
	action, err := getEventAction(eventType)
	if err != nil {
		return "", err
	}

	// Make object singular by removing 's' suffix for plural forms.
	// Current Salesloft objects only use 's' suffix for plurals (e.g., "accounts" -> "account"),
	// so simple trimming is sufficient. For future complex pluralization, consider using
	// a library like "github.com/jinzhu/inflection" or similar.
	objNameStr := string(objectName)
	if strings.HasSuffix(objNameStr, "s") {
		objNameStr = strings.TrimSuffix(objNameStr, "s")
	}

	// Salesloft format: "{objectName}_{action}"
	combined := fmt.Sprintf("%s_%s", objNameStr, action)
	return SalesloftEventType(combined), nil
}

// getEventAction converts common event types to Salesloft event actions
func getEventAction(eventType common.SubscriptionEventType) (EventAction, error) {
	switch eventType { //nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return ActionCreated, nil
	case common.SubscriptionEventTypeUpdate:
		return ActionUpdated, nil
	case common.SubscriptionEventTypeDelete:
		return ActionDestroyed, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedEventType, eventType)
	}
}
