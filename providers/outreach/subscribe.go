package outreach

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	var failedSubscriptions []FailedSubscription

	// Process all object+event combinations
	for obj, events := range params.SubscriptionEvents {
		providerEvents := events.Events
		for _, event := range providerEvents {
			successful, failed := c.createSingleSubscription(ctx, event, obj, req)
			if successful != nil {
				successfulSubscriptions = append(successfulSubscriptions, *successful)
			} else {
				failedSubscriptions = append(failedSubscriptions, *failed)
			}
		}
	}

	subscriptionResult := &common.SubscriptionResult{
		Result: &SubscriptionResultData{
			SuccessfulSubscriptions: successfulSubscriptions,
			FailedSubscriptions:     failedSubscriptions,
		},
		ObjectEvents: params.SubscriptionEvents,
	}

	if len(failedSubscriptions) > 0 {
		if len(successfulSubscriptions) > 0 {
			// Partial success - some worked, some failed
			subscriptionResult.Status = common.SubscriptionStatusSuccess

			return subscriptionResult, nil
		} else {
			// Complete failure - nothing worked
			subscriptionResult.Status = common.SubscriptionStatusFailed

			return subscriptionResult, nil
		}
	}

	// Complete success - everything worked
	subscriptionResult.Status = common.SubscriptionStatusSuccess

	return subscriptionResult, nil
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

//nolint:cyclop,funlen
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	if previousResult.Result == nil {
		return nil, fmt.Errorf("%w: Result cannot be null", errMissingParams) //nolint:err113,lll
	}

	subscriptionData, ok := previousResult.Result.(*SubscriptionResultData)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected SubscriptionResult to be type %T but got %T",
			errInvalidRequestType,
			subscriptionData,
			previousResult.Result)
	}

	if len(subscriptionData.SuccessfulSubscriptions) == 0 {
		return nil, fmt.Errorf("%w: subscription is empty", errMissingParams)
	}

	// currentSubs are the currently active subscriptions
	// desiredSubs are the subscriptions we want to have after the update
	currentSubs, desiredSubs := buildSubscriptionMaps(
		subscriptionData.SuccessfulSubscriptions,
		params.SubscriptionEvents,
	)

	var newsuccessfulSubscriptions []SuccessfulSubscription //nolint:prealloc

	var newfailedSubscriptions []FailedSubscription

	// Remove subscriptions that are no longer desired
	for key, subID := range currentSubs {
		if !desiredSubs[key] {
			err := c.deleteSubscription(ctx, subID)
			if err != nil {
				newfailedSubscriptions = append(newfailedSubscriptions, FailedSubscription{
					ObjectName: strings.Split(key, ":")[0],
					EventName:  strings.Split(key, ":")[1],
					Error:      fmt.Sprintf("failed to delete subscription for %s: %v", key, err),
				})

				continue
			}

			delete(currentSubs, key)
		}
	}

	// Add new subscriptions that are desired but not currently present
	for key := range desiredSubs {
		_, exist := currentSubs[key]

		// If the subscription already exists, we don't need to create it
		if exist {
			continue
		}

		parts := strings.Split(key, ":")
		objectName, event := parts[0], parts[1]

		successful, failed := c.createSingleSubscription(ctx, common.SubscriptionEventType(event), common.ObjectName(objectName), req) //nolint:lll
		if successful != nil {
			newsuccessfulSubscriptions = append(newsuccessfulSubscriptions, *successful)
		} else {
			newfailedSubscriptions = append(newfailedSubscriptions, *failed)
		}
	}

	var updatedSuccessfulSubscriptions []SuccessfulSubscription

	// Delete subscriptions that are no longer desired
	for _, sub := range subscriptionData.SuccessfulSubscriptions {
		key := fmt.Sprintf("%s:%s", sub.ObjectName, sub.EventName)
		if _, exist := desiredSubs[key]; exist {
			updatedSuccessfulSubscriptions = append(updatedSuccessfulSubscriptions, sub)
		}
	}

	updatedSuccessfulSubscriptions = append(updatedSuccessfulSubscriptions, newsuccessfulSubscriptions...)

	var updatedFailedSubscriptions []FailedSubscription
	updatedFailedSubscriptions = append(updatedFailedSubscriptions, subscriptionData.FailedSubscriptions...)
	updatedFailedSubscriptions = append(updatedFailedSubscriptions, newfailedSubscriptions...)

	updatedResult := &common.SubscriptionResult{
		Result: &SubscriptionResultData{
			SuccessfulSubscriptions: updatedSuccessfulSubscriptions,
			FailedSubscriptions:     updatedFailedSubscriptions,
		},
		ObjectEvents: params.SubscriptionEvents,
	}

	if len(updatedFailedSubscriptions) > 0 {
		if len(updatedSuccessfulSubscriptions) > 0 {
			// Partial success - some worked, some failed
			updatedResult.Status = common.SubscriptionStatusSuccess

			return updatedResult, nil
		} else {
			// Complete failure - nothing worked
			updatedResult.Status = common.SubscriptionStatusFailed

			return updatedResult, nil
		}
	}

	// Complete success - everything worked
	updatedResult.Status = common.SubscriptionStatusSuccess

	return updatedResult, nil
}

// createSingleSubscription attempts to create a single subscription and returns either a successful or failed result.
func (c *Connector) createSingleSubscription(
	ctx context.Context,
	event common.SubscriptionEventType,
	obj common.ObjectName,
	req *SubscriptionRequest,
) (*SuccessfulSubscription, *FailedSubscription) {
	payload, err := buildPayload(event, obj, req.WebhookEndPoint, req.Secret)
	if err != nil {
		return nil, &FailedSubscription{
			ObjectName: string(obj),
			EventName:  string(event),
			Error:      fmt.Sprintf("failed to create subscription for object %s, event %s: %v", obj, event, err),
		}
	}

	result, err := c.createSubscriptions(ctx, payload, c.Client.Post)
	if err != nil {
		return nil, &FailedSubscription{
			ObjectName: string(obj),
			EventName:  string(event),
			Error:      fmt.Sprintf("failed to create subscription for object %s, event %s: %v", obj, event, err),
		}
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

// buildSubscriptionMaps creates maps of current and desired subscriptions.
func buildSubscriptionMaps(
	successfulSubs []SuccessfulSubscription,
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
) (map[string]string, map[string]bool) {
	// currentSubs maps "objectName:eventName" to subscription ID
	currentSubs := make(map[string]string)

	for _, sub := range successfulSubs {
		key := fmt.Sprintf("%s:%s", sub.ObjectName, sub.EventName)
		currentSubs[key] = sub.ID
	}

	// desiredSubs maps "objectName:eventName" to true
	desiredSubs := make(map[string]bool)

	for obj, events := range subscriptionEvents {
		for _, evt := range events.Events {
			key := fmt.Sprintf("%s:%s", string(obj), string(evt))
			desiredSubs[key] = true
		}
	}

	return currentSubs, desiredSubs
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
