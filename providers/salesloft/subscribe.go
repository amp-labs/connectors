package salesloft

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
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
		Result: &subscriptionResult{},
	}
}

// nolint: funlen,cyclop,gocognit,godoclint
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

	// Store successful subscriptions with their full response data
	subscriptionsMap := make(map[common.ObjectName]map[moduleEvent]subscriptionResponse)
	successfulSubscriptions := make([]successfulSubscription, 0)

	var firstError error

	var errorOnce sync.Once

	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0)

	// Process all object+event combinations
	for obj, events := range params.SubscriptionEvents {
		for _, event := range events.Events {
			// This converts common event type to Salesloft event type format and also
			// expands events if needed (e.g., "tasks" update -> "task_completed" and "task_updated")
			providerEvents, err := toModuleEventName(obj, event)
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
							subscriptionsMap[currObj] = make(map[moduleEvent]subscriptionResponse)
						}

						subscriptionsMap[currObj][currProviderEvent] = *response

						// Keep track of successful subscriptions for rollback
						successfulSubscriptions = append(successfulSubscriptions, successfulSubscription{
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
	res.Result = &subscriptionResult{
		Subscriptions: subscriptionsMap,
	}

	return res, nil
}

// nolint: funlen, cyclop,gocognit,gocyclo,godoclint
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Validate previous result
	prevState, err := validatePreviousResult(previousResult)
	if err != nil {
		return nil, err
	}

	// Build subscription maps
	existingSubscriptions, requestedSubscriptions, err := buildSubscriptionMaps(prevState, params)
	if err != nil {
		return nil, err
	}

	// Categorize existing subscriptions into delete/keep buckets.
	subscriptionsToDelete, subscriptionsToKeep := categorizeSubscriptions(prevState, requestedSubscriptions)

	// Delete subscriptions that are no longer needed
	if len(subscriptionsToDelete.Subscriptions) > 0 {
		deleteResult := common.SubscriptionResult{Result: subscriptionsToDelete}
		if err := c.DeleteSubscription(ctx, deleteResult); err != nil {
			return nil, fmt.Errorf("failed to delete previous subscriptions: %w", err)
		}
	}

	// Determine what to create (in requested but not in existing).
	// We check against existingSubscriptions to avoid recreating webhooks that already exist.
	newSubscriptionEvents := make(map[common.ObjectName]common.ObjectEvents)

	for objName, events := range params.SubscriptionEvents {
		var eventsToCreate []common.SubscriptionEventType

		for _, event := range events.Events {
			providerEvents, err := toModuleEventName(objName, event)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event type %s: %w", event, err)
			}

			for _, providerEvt := range providerEvents {
				// Only create if it doesn't already exist (wasn't in the "keep" list)
				if !existingSubscriptions[providerEvt] {
					// Avoid duplicates
					// Events like "update" can map to multiple provider events,
					// so we need to ensure we don't add the same common event multiple times.
					if !slices.Contains(eventsToCreate, event) {
						eventsToCreate = append(eventsToCreate, event)
					}
				}
			}
		}

		if len(eventsToCreate) > 0 {
			newSubscriptionEvents[objName] = common.ObjectEvents{Events: eventsToCreate}
		}
	}

	// Create new subscriptions
	var createResult *common.SubscriptionResult

	if len(newSubscriptionEvents) > 0 {
		newParams := params
		newParams.SubscriptionEvents = newSubscriptionEvents

		var err error

		createResult, err = c.Subscribe(ctx, newParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create new subscriptions: %w", err)
		}
	}

	finalResult := mergeSubscriptionResults(subscriptionsToKeep, createResult)
	finalObjectEvents := buildFinalObjectEvents(finalResult)

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		Result:       finalResult,
		ObjectEvents: finalObjectEvents,
	}, nil
}

// DeleteSubscription deletes webhook subscriptions.
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	if result.Result == nil {
		return fmt.Errorf("%w: Result cannot be null", errMissingParams)
	}

	subscriptionData, ok := result.Result.(*subscriptionResult)
	if !ok {
		return fmt.Errorf("%w: expected subscriptionResult to be type %T but got %T",
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
	event moduleEvent,
	obj common.ObjectName,
	req *subscriptionRequest,
) (*subscriptionResponse, error) {
	payload := &subscriptionPayload{
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
	payload *subscriptionPayload,
) (*subscriptionResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[subscriptionResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return result, nil
}

// rollbackSubscriptions attempts to delete all successful subscriptions in case of partial failure.
func (c *Connector) rollbackSubscriptions(
	ctx context.Context,
	subscriptions []successfulSubscription,
) (rolledBack []successfulSubscription, failedToRollBack []successfulSubscription, err error) {
	var rollbackErrors error

	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0, len(subscriptions))

	for _, subFromList := range subscriptions {
		callbacks = append(callbacks,
			func(sub successfulSubscription) func(ctx context.Context) error {
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

func validateRequest(params common.SubscribeParams) (*subscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*subscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T', got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if validate.Struct(req) != nil {
		return nil, fmt.Errorf("%w: request is invalid", errInvalidRequestType)
	}

	return req, nil
}

// toModuleEventName converts a common event type into one or more Salesloft module events using the mapping.
func toModuleEventName(objectName common.ObjectName, eventType common.SubscriptionEventType) ([]moduleEvent, error) {
	mapping, exists := salesloftEventMappings[objectName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errUnsupportedObject, objectName)
	}

	events := mapping.Events.toProviderEvents(eventType)
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: %s for object %s", errUnsupportedEventType, eventType, objectName)
	}

	return events, nil
}

func validateSubscriptionEvents(subscriptionEvents map[common.ObjectName]common.ObjectEvents) error {
	var validationErrors error

	for objectName, events := range subscriptionEvents {
		mapping, exist := salesloftEventMappings[objectName]
		if !exist {
			validationErrors = errors.Join(validationErrors,
				fmt.Errorf("%s %w", objectName, errUnsupportedObject))

			continue
		}

		// Get all supported events for this object
		supportedEvents := mapping.Events.getAllSupportedEvents()

		supportedSet := make(map[moduleEvent]bool)
		for _, evt := range supportedEvents {
			supportedSet[evt] = true
		}

		for _, event := range events.Events {
			providerEvents := mapping.Events.toProviderEvents(event)
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

// toProviderEvents converts a common event type to one or more Salesloft provider events.
func (m eventMapping) toProviderEvents(commonEvent common.SubscriptionEventType) []moduleEvent {
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

// getAllSupportedEvents returns all provider events that this mapping supports.
func (m eventMapping) getAllSupportedEvents() []moduleEvent {
	var events []moduleEvent

	events = append(events, m.CreateEvents...)
	events = append(events, m.UpdateEvents...)
	events = append(events, m.DeleteEvents...)

	return events
}

// moduleEventToCommon converts a provider event back to common event type.
func moduleEventToCommon(e moduleEvent, objectName common.ObjectName) (common.SubscriptionEventType, bool) {
	mapping, exists := salesloftEventMappings[objectName]
	if !exists {
		return "", false
	}

	commonEvent, found := mapping.Events.toCommonEvent(e)
	if found {
		return commonEvent, true
	}

	return "", false
}

// ToCommonEvent converts a provider event back to common event type.
func (m eventMapping) toCommonEvent(providerEvent moduleEvent) (common.SubscriptionEventType, bool) {
	if slices.Contains(m.CreateEvents, providerEvent) {
		return common.SubscriptionEventTypeCreate, true
	}

	if slices.Contains(m.UpdateEvents, providerEvent) {
		return common.SubscriptionEventTypeUpdate, true
	}

	if slices.Contains(m.DeleteEvents, providerEvent) {
		return common.SubscriptionEventTypeDelete, true
	}

	return "", false
}

func validatePreviousResult(previousResult *common.SubscriptionResult) (*subscriptionResult, error) {
	// Validate the previous result
	if previousResult == nil || previousResult.Result == nil {
		return nil, fmt.Errorf("%w: missing previousResult or previousResult.Result", errMissingParams)
	}

	prevState, ok := previousResult.Result.(*subscriptionResult)
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

func buildSubscriptionMaps(prevState *subscriptionResult, params common.SubscribeParams) (
	map[moduleEvent]bool,
	map[moduleEvent]bool,
	error,
) {
	// Build a map of existing subscriptions for quick lookup.
	// This composite key allows O(1) lookup when comparing existing vs requested subscriptions.
	existingSubscriptions := make(map[moduleEvent]bool)

	for _, eventsMap := range prevState.Subscriptions {
		for eventName := range eventsMap {
			existingSubscriptions[eventName] = true
		}
	}

	// Build a map of requested subscriptions
	requestedSubscriptions := make(map[moduleEvent]bool)

	for objName, events := range params.SubscriptionEvents {
		for _, event := range events.Events {
			providerEvents, expandErr := toModuleEventName(objName, event)
			if expandErr != nil {
				return nil, nil, fmt.Errorf("failed to expand event type %s: %w", event, expandErr)
			}

			for _, providerEvt := range providerEvents {
				requestedSubscriptions[providerEvt] = true
			}
		}
	}

	return existingSubscriptions, requestedSubscriptions, nil
}

func buildFinalObjectEvents(finalResult *subscriptionResult) map[common.ObjectName]common.ObjectEvents {
	finalObjectEvents := make(map[common.ObjectName]common.ObjectEvents)

	for objName, eventsMap := range finalResult.Subscriptions {
		var events []common.SubscriptionEventType

		for eventName := range eventsMap {
			// Convert provider event name back to common event type
			if commonEvent, found := moduleEventToCommon(eventName, objName); found {
				// Avoid duplicates
				// Events like "update" can map to multiple provider events,
				// so we need to ensure we don't add the same common event multiple times.
				if !slices.Contains(events, commonEvent) {
					events = append(events, commonEvent)
				}
			}
		}

		if len(events) > 0 {
			finalObjectEvents[objName] = common.ObjectEvents{Events: events}
		}
	}

	return finalObjectEvents
}

func mergeSubscriptionResults(kept *subscriptionResult, created *common.SubscriptionResult) *subscriptionResult {
	// Merge the results: kept subscriptions + newly created subscriptions.
	// Start with subscriptions we kept from the previous state (these already existed and are still wanted).
	finalResult := &subscriptionResult{
		Subscriptions: kept.Subscriptions,
	}

	// Add newly created subscriptions to the final result.
	if created != nil && created.Result != nil {
		if newSubs, ok := created.Result.(*subscriptionResult); ok {
			for objName, eventsMap := range newSubs.Subscriptions {
				if finalResult.Subscriptions[objName] == nil {
					finalResult.Subscriptions[objName] = make(map[moduleEvent]subscriptionResponse)
				}
				maps.Copy(finalResult.Subscriptions[objName], eventsMap)
			}
		}
	}

	return finalResult
}

// Categorize existing subscriptions into delete/keep buckets.
//
// Algorithm:
// - If subscription exists but is NOT requested → delete it (webhook no longer needed)
// - If subscription exists AND is requested → keep it (reuse existing webhook)
// - If subscription is requested but NOT existing → will be created in next step.
func categorizeSubscriptions(
	prevState *subscriptionResult,
	requestedSubscriptions map[moduleEvent]bool,
) (subscriptionsToDelete, subscriptionsToKeep *subscriptionResult) {
	subscriptionsToDelete = &subscriptionResult{
		Subscriptions: make(map[common.ObjectName]map[moduleEvent]subscriptionResponse),
	}
	subscriptionsToKeep = &subscriptionResult{
		Subscriptions: make(map[common.ObjectName]map[moduleEvent]subscriptionResponse),
	}

	for objName, eventsMap := range prevState.Subscriptions {
		for eventName, response := range eventsMap {
			if !requestedSubscriptions[eventName] {
				// Need to delete this subscription
				if subscriptionsToDelete.Subscriptions[objName] == nil {
					subscriptionsToDelete.Subscriptions[objName] = make(map[moduleEvent]subscriptionResponse)
				}

				subscriptionsToDelete.Subscriptions[objName][eventName] = response
			} else {
				// Keep this subscription
				if subscriptionsToKeep.Subscriptions[objName] == nil {
					subscriptionsToKeep.Subscriptions[objName] = make(map[moduleEvent]subscriptionResponse)
				}

				subscriptionsToKeep.Subscriptions[objName][eventName] = response
			}
		}
	}

	return subscriptionsToDelete, subscriptionsToKeep
}
