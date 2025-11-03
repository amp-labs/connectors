package outreach

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
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
	for obj, event := range params.SubscriptionEvents {
		providerEvents := event.Events
		for _, providerEvt := range providerEvents {
			Event, err := getProviderEventName(providerEvt)
			if err != nil {
				failedSubscriptions = append(failedSubscriptions, FailedSubscription{
					ObjectName: string(obj),
					EventName:  string(providerEvt),
					Error:      fmt.Sprintf("failed to map event type for object %s, event %s: %v", obj, providerEvt, err),
				})

				continue
			}

			payload := &SubscriptionPayload{
				Data: SubscriptionData{
					Type: "webhook",
					Attributes: AttributesPayload{
						Action:   string(Event),
						Resource: string(obj),
						URL:      req.WebhookEndPoint,
					},
				},
			}

			result, err := c.createSubscriptions(ctx, payload, c.Client.Post)
			if err != nil {
				failedSubscriptions = append(failedSubscriptions, FailedSubscription{
					ObjectName: string(obj),
					EventName:  string(providerEvt),
					Error:      fmt.Sprintf("failed to create subscription for object %s, event %s: %v", obj, providerEvt, err),
				})

				continue
			}

			successfulSubscriptions = append(successfulSubscriptions, SuccessfulSubscription{
				ID:         result.Data.ID,
				ObjectName: string(obj),
				EventName:  string(providerEvt),
			})
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

func (c *Connector) getSubscribeURL() (*string, error) {
	url, err := c.getApiURL("webhooks")
	if err != nil {
		return nil, err
	}

	urlStr := url.String()

	return &urlStr, nil
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

func (c *Connector) createSubscriptions(ctx context.Context,
	payload *SubscriptionPayload,
	updater common.WriteMethod,
) (*createSubscriptionsResponse, error) {
	url, err := c.getSubscribeURL()
	if err != nil {
		return nil, err
	}

	resp, err := updater(ctx, *url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	result, err := common.UnmarshalJSON[createSubscriptionsResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription response: %w", err)
	}

	return result, nil
}
