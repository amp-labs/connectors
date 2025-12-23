package outreach

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"sync"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
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

// Subscribe creates subscriptions for the specified objects and events.
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
	subscriptionsMap := make(map[common.ObjectName]map[ModuleEvent]createSubscriptionsResponse)
	successfulSubscriptions := make([]SuccessfulSubscription, 0)

	var firstError error

	var errorOnce sync.Once

	var mutex sync.Mutex

	callbacks := make([]simultaneously.Job, 0)

	// Process all object+event combinations
	for obj, events := range params.SubscriptionEvents {
		for _, event := range events.Events {
			currObj := common.ObjectName(naming.NewSingularString(string(obj)).String())
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
					// Convert common event type to provider event type (ModuleEvent) as string
					providerEvent, err := getProviderEventName(currentEvent)
					if err != nil {
						errorOnce.Do(func() {
							firstError = fmt.Errorf("failed to convert event type %s: %w", currentEvent, err)
						})

						return nil
					}

					// Initialize nested map if needed
					if subscriptionsMap[currObj] == nil {
						subscriptionsMap[currObj] = make(map[ModuleEvent]createSubscriptionsResponse)
					}

					subscriptionsMap[currObj][providerEvent] = *response

					// Keep track of successful subscriptions for rollback
					successfulSubscriptions = append(successfulSubscriptions, SuccessfulSubscription{
						ID:         strconv.Itoa(response.Data.ID),
						ObjectName: string(currObj),
						EventName:  string(providerEvent),
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

// UpdateSubscription updates an existing subscription by comparing the previous
// subscription state with the new desired state.
// It deletes subscriptions that are no longer needed and creates new ones.
//
//nolint:funlen,cyclop,gocognit,gocyclo
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
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
		)
	}

	// Build a map of existing subscriptions for quick lookup.
	// Key format: "objectName:eventName" (e.g., "account:created")
	// This composite key allows O(1) lookup when comparing existing vs requested subscriptions.
	existingSubscriptions := make(map[string]bool)

	for objName, eventsMap := range prevState.Subscriptions {
		for eventName := range eventsMap {
			key := string(objName) + ":" + string(eventName)
			existingSubscriptions[key] = true
		}
	}

	// Build a map of requested subscriptions
	requestedSubscriptions := make(map[string]bool)

	for objName, events := range params.SubscriptionEvents {
		for _, event := range events.Events {
			providerEvent, err := getProviderEventName(event)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event type %s: %w", event, err)
			}

			key := string(objName) + ":" + string(providerEvent)
			requestedSubscriptions[key] = true
		}
	}

	// Categorize existing subscriptions into delete/keep buckets.
	//
	// Algorithm:
	// - If subscription exists but is NOT requested → delete it (webhook no longer needed)
	// - If subscription exists AND is requested → keep it (reuse existing webhook)
	// - If subscription is requested but NOT existing → will be created in next step
	subscriptionsToDelete := &SubscriptionResult{
		Subscriptions: make(map[common.ObjectName]map[ModuleEvent]createSubscriptionsResponse),
	}
	subscriptionsToKeep := &SubscriptionResult{
		Subscriptions: make(map[common.ObjectName]map[ModuleEvent]createSubscriptionsResponse),
	}

	for objName, eventsMap := range prevState.Subscriptions {
		for eventName, response := range eventsMap {
			key := string(objName) + ":" + string(eventName)
			if !requestedSubscriptions[key] {
				// Need to delete this subscription
				if subscriptionsToDelete.Subscriptions[objName] == nil {
					subscriptionsToDelete.Subscriptions[objName] = make(map[ModuleEvent]createSubscriptionsResponse)
				}

				subscriptionsToDelete.Subscriptions[objName][eventName] = response
			} else {
				// Keep this subscription
				if subscriptionsToKeep.Subscriptions[objName] == nil {
					subscriptionsToKeep.Subscriptions[objName] = make(map[ModuleEvent]createSubscriptionsResponse)
				}

				subscriptionsToKeep.Subscriptions[objName][eventName] = response
			}
		}
	}

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
			providerEvent, err := getProviderEventName(event)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event type %s: %w", event, err)
			}

			key := string(objName) + ":" + string(providerEvent)
			// Only create if it doesn't already exist (wasn't in the "keep" list)
			if !existingSubscriptions[key] {
				eventsToCreate = append(eventsToCreate, event)
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

	// Merge the results: kept subscriptions + newly created subscriptions.
	// Start with subscriptions we kept from the previous state (these already existed and are still wanted).
	finalResult := &SubscriptionResult{
		Subscriptions: subscriptionsToKeep.Subscriptions,
	}

	// Add newly created subscriptions to the final result.
	if createResult != nil && createResult.Result != nil {
		newSubs, ok := createResult.Result.(*SubscriptionResult)
		if ok {
			for objName, eventsMap := range newSubs.Subscriptions {
				if finalResult.Subscriptions[objName] == nil {
					finalResult.Subscriptions[objName] = make(map[ModuleEvent]createSubscriptionsResponse)
				}

				maps.Copy(finalResult.Subscriptions[objName], eventsMap)
			}
		}
	}

	// Build the final ObjectEvents map
	finalObjectEvents := make(map[common.ObjectName]common.ObjectEvents)

	for objName, eventsMap := range finalResult.Subscriptions {
		var events []common.SubscriptionEventType

		for eventName := range eventsMap {
			// Convert provider event name back to common event type
			switch eventName {
			case Created:
				events = append(events, common.SubscriptionEventTypeCreate)
			case Updated:
				events = append(events, common.SubscriptionEventTypeUpdate)
			case Destroyed:
				events = append(events, common.SubscriptionEventTypeDelete)
			}
		}

		if len(events) > 0 {
			finalObjectEvents[objName] = common.ObjectEvents{Events: events}
		}
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		Result:       finalResult,
		ObjectEvents: finalObjectEvents,
	}, nil
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

	if len(subscriptionData.Subscriptions) == 0 {
		return fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	// Extract subscription IDs from the nested map and delete them
	for objName, eventsMap := range subscriptionData.Subscriptions {
		for eventType, response := range eventsMap {
			err := c.deleteSubscription(ctx, strconv.Itoa(response.Data.ID))
			if err != nil {
				return fmt.Errorf(
					"failed to delete subscription for object %s, event %s (ID: %d): %w",
					objName,
					eventType,
					response.Data.ID,
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
	event common.SubscriptionEventType,
	obj common.ObjectName,
	req *SubscriptionRequest,
) (*createSubscriptionsResponse, error) {
	payload, err := buildPayload(event, obj, req.WebhookEndPoint, req.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %w", obj, event, err)
	}

	result, err := c.createSubscriptions(ctx, payload, c.Client.Post)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription for object %s, event %s: %w", obj, event, err)
	}

	return result, nil
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
	if err != nil {
		return err
	}

	return nil
}

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
