package salesloft

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

// nolint: funlen, cyclop,gocognit
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	// Validate that requested events are supported
	err = validateSubscriptionRequest(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	// Store successful subscriptions with their full response data
	subscriptionsMap := make(map[common.ObjectName]map[ModuleEvent]SubscriptionResponse)
	successfulSubscriptions := make([]SuccessfulSubscription, 0)

	var firstError error

	var errorOnce sync.Once

	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0)

	// Process all object+event combinations
	for obj, events := range params.SubscriptionEvents {
		for _, event := range events.Events {
			// This converts common event type to Salesloft event type format and also
			// expands events if needed (e.g., "tasks" update -> "task_completed" and "task_updated")
			providerEvents, err := expandEvent(obj, event)
			if err != nil {
				return nil, err
			}

			for _, providerEvent := range providerEvents {
				currObj := obj
				currProviderEvent := providerEvent

				callbacks = append(callbacks, func(ctx context.Context) error {
					response, failErr := c.createSingleSubscription(ctx, currProviderEvent, currObj, req)

					mutex.Lock()
					defer mutex.Unlock()

					if failErr != nil {
						errorOnce.Do(func() {
							firstError = failErr
						})
					} else {
						// Initialize nested map if needed
						if subscriptionsMap[currObj] == nil {
							subscriptionsMap[currObj] = make(map[ModuleEvent]SubscriptionResponse)
						}

						subscriptionsMap[currObj][currProviderEvent] = *response

						// Keep track of successful subscriptions for rollback
						successfulSubscriptions = append(successfulSubscriptions, SuccessfulSubscription{
							ID:         strconv.Itoa(response.ID),
							ObjectName: string(currObj),
							EventName:  string(currProviderEvent),
						})
					}

					return nil
				})
			}
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
				if _, ok := objectEvents[common.ObjectName(failedSub.ObjectName)]; !ok {
					objectEvents[common.ObjectName(failedSub.ObjectName)] = common.ObjectEvents{
						Events: []common.SubscriptionEventType{},
					}
				}

				currentEvent := objectEvents[common.ObjectName(failedSub.ObjectName)]

				currentEvent.Events = append(currentEvent.Events, common.SubscriptionEventType(failedSub.EventName))

				objectEvents[common.ObjectName(failedSub.ObjectName)] = currentEvent
			}

			res.ObjectEvents = objectEvents

			return res, errors.Join(firstError, rollbackErr)
		}

		res.Status = common.SubscriptionStatusFailed
		res.ObjectEvents = nil
		// rolledBack and failedToRollBack are available for caller to use if needed

		return res, firstError
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = &SubscriptionResult{
		Subscriptions: subscriptionsMap,
	}

	return res, nil
}

func (c *Connector) UpdateSubscription(ctx context.Context,
	params common.SubscribeParams, previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	panic("unimplemented")
}

// DeleteSubscription deletes webhook subscriptions.
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

// createSingleSubscription attempts to create a single subscription and returns the full response.
func (c *Connector) createSingleSubscription(
	ctx context.Context,
	event ModuleEvent,
	obj common.ObjectName,
	req *SubscriptionRequest,
) (*SubscriptionResponse, error) {
	payload := &SubscriptionPayload{
		CallbackURL:   req.WebhookEndPoint,
		EventType:     string(event),
		CallbackToken: req.Secret,
	}

	result, err := c.createSubscription(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %w", obj, event, err)
	}

	return result, nil
}

// createSubscription makes the API call to create a webhook subscription.
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

// rollbackSubscriptions attempts to delete all successful subscriptions in case of partial failure.
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

func (c *Connector) getSubscribeURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, ApiVersion, "webhook_subscriptions")
}

// deleteSubscription deletes a single subscription by ID.
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

// expandEvent converts a common event type into one or more Salesloft module events using the mapping.
func expandEvent(objectName common.ObjectName, eventType common.SubscriptionEventType) ([]ModuleEvent, error) {
	mapping, exists := salesloftEventMappings[objectName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errUnsupportedObject, objectName)
	}

	events := mapping.Events.ToProviderEvents(eventType)
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: %s for object %s", errUnsupportedEventType, eventType, objectName)
	}

	return events, nil
}

func validateSubscriptionRequest(subscriptionEvents map[common.ObjectName]common.ObjectEvents) error {
	var validationErrors error

	for objectName, events := range subscriptionEvents {
		mapping, exist := salesloftEventMappings[objectName]
		if !exist {
			validationErrors = errors.Join(validationErrors,
				fmt.Errorf("%s %w", objectName, errUnsupportedObject))

			continue
		}

		// Get all supported events for this object
		supportedEvents := mapping.Events.GetAllSupportedEvents()

		supportedSet := make(map[ModuleEvent]bool)
		for _, evt := range supportedEvents {
			supportedSet[evt] = true
		}

		for _, event := range events.Events {
			providerEvents := mapping.Events.ToProviderEvents(event)
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

// ToProviderEvents converts a common event type to one or more Salesloft provider events.
func (m EventMapping) ToProviderEvents(commonEvent common.SubscriptionEventType) []ModuleEvent {
	switch commonEvent { // nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return m.CreateEvents
	case common.SubscriptionEventTypeUpdate:
		return m.UpdateEvents
	case common.SubscriptionEventTypeDelete:
		return m.DeleteEvents
	default:
		return nil
	}
}

// GetAllSupportedEvents returns all provider events that this mapping supports.
func (m EventMapping) GetAllSupportedEvents() []ModuleEvent {
	var events []ModuleEvent
	events = append(events, m.CreateEvents...)
	events = append(events, m.UpdateEvents...)
	events = append(events, m.DeleteEvents...)

	return events
}
