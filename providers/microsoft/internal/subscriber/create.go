package subscriber

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
	"github.com/amp-labs/connectors/providers/microsoft/internal/webhook"
)

const subscriptionExpirationWindow = 5 * time.Hour

// Subscribe creates a Microsoft Graph subscription for the specified objects and events.
//
// nolint:lll
// See the [request body]( https://learn.microsoft.com/en-us/graph/api/subscription-post-subscriptions?view=graph-rest-1.0&tabs=http#request-body) for details.
// Supported resources are listed [here](https://learn.microsoft.com/en-us/graph/api/resources/change-notifications-api-overview?view=graph-rest-1.0).
func (s Strategy) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return s.createSubscription(ctx, params)
}

// createSubscription creates subscriptions using batch requests for efficiency.
// Handles rollback on partial failures to maintain consistency.
// Pre-existing subscriptions for the same resource may result in duplicates (handled only by UpdateSubscription).
func (s Strategy) createSubscription(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	batchParams, err := s.paramsForBatchCreateSubscriptions(params)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)
	state := getStateFromCreateResponse(bundledResponse)

	if len(bundledResponse.Errors) != 0 {
		// Some requests failed; initiate rollback.
		res, err := s.rollbackSubscriptionCreation(ctx, params, state, bundledResponse)
		if err != nil {
			return nil, err
		}

		err = nil
		for _, wrapper := range bundledResponse.Errors {
			err = errors.Join(err, wrapper.Data)
		}

		return res, err
	}

	subscriptions := make(map[string]SubscriptionResource)
	for _, subscription := range bundledResponse.Responses {
		subscriptions[subscription.Data.ID] = subscription.Data
	}

	return &common.SubscriptionResult{
		Result:       &Result{Subscriptions: subscriptions},
		ObjectEvents: state,
		Status:       common.SubscriptionStatusSuccess,
	}, nil
}

// paramsForBatchCreateSubscriptions prepares batch parameters for creating multiple subscriptions.
// Constructs payloads for each object-event combination using the webhook URL.
func (s Strategy) paramsForBatchCreateSubscriptions(params common.SubscribeParams) (*batch.Params, error) {
	input, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	webhookURL := input.WebhookURL

	url, err := s.getSubscriptionURL()
	if err != nil {
		return nil, err
	}

	batchParams := &batch.Params{}

	for objectName, events := range params.SubscriptionEvents {
		requestID := batch.RequestID(objectName)
		body := newPayloadCreateSubscription(objectName, events, webhookURL)
		batchParams.WithRequest(requestID, http.MethodPost, url, body, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}

// getStateFromCreateResponse extracts the subscription state from batch responses.
// Successful responses set the events state; errors result in empty ObjectEvents state.
func getStateFromCreateResponse(
	response *batch.Result[SubscriptionResource],
) map[common.ObjectName]common.ObjectEvents {
	result := make(map[common.ObjectName]common.ObjectEvents)

	for objectName, envelope := range response.Responses {
		// Map successful subscription to its events.
		subscription := envelope.Data
		result[common.ObjectName(objectName)] = common.ObjectEvents{
			Events:            subscription.ChangeType.EventTypes(),
			WatchFields:       nil,
			WatchFieldsAll:    false,
			PassThroughEvents: nil,
		}
	}

	for objectName := range response.Errors {
		// Failed requests yield no subscription.
		result[common.ObjectName(objectName)] = common.ObjectEvents{}
	}

	return result
}

// rollbackSubscriptionCreation deletes successfully created subscriptions on partial failure.
// It updates state to reflect remaining subscriptions after attempted rollback.
// Returns appropriate status based on rollback success.
func (s Strategy) rollbackSubscriptionCreation(
	ctx context.Context,
	params common.SubscribeParams,
	state map[common.ObjectName]common.ObjectEvents,
	partialCreation *batch.Result[SubscriptionResource],
) (*common.SubscriptionResult, error) {
	subscriptions := make(map[string]SubscriptionResource)
	for _, envelope := range partialCreation.Responses {
		subscriptions[envelope.Data.ID] = envelope.Data
	}

	bundledResponse, err := s.removeSubscriptionsByIds(ctx, datautils.FromMap(subscriptions).Keys())
	if err != nil {
		return nil, err
	}

	if len(bundledResponse.Errors) == 0 {
		// Full rollback succeeded.
		objectNames := datautils.FromMap(params.SubscriptionEvents).Keys()

		return &common.SubscriptionResult{
			Result:       &Result{Subscriptions: map[string]SubscriptionResource{}},
			ObjectEvents: common.NewEmptyObjectEvents(objectNames),
			Status:       common.SubscriptionStatusFailed,
		}, nil
	}

	// Partial rollback; track remaining subscriptions.
	existingObjects := datautils.NewSet[common.ObjectName]()

	for requestID := range bundledResponse.Errors {
		// Convert request ID back to object.
		id := string(requestID)
		subscription := subscriptions[id]
		existingObjects.AddOne(subscription.ObjectName)
	}

	allObjects := datautils.FromMap(params.SubscriptionEvents).KeySet()
	removedObjects := allObjects.Subtract(existingObjects)

	// Clear state for successfully removed objects.
	for _, objectName := range removedObjects {
		state[objectName] = common.ObjectEvents{}
		for _, subscription := range subscriptions {
			if subscription.ObjectName == objectName {
				delete(subscriptions, subscription.ID)
			}
		}
	}

	return &common.SubscriptionResult{
		Result:       &Result{Subscriptions: subscriptions},
		ObjectEvents: state,
		Status:       common.SubscriptionStatusFailedToRollback,
	}, nil
}

// SubscriptionResource represents a Microsoft Graph subscription.
// See [properties](https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#properties).
//
// Custom usage: clientState field is repurposed to store ObjectName for identification.
// Ignored fields:
//
//	encryptionCertificateId, encryptionCertificate, lifecycleNotificationUrl,
//	notificationQueryOptions, notificationUrlAppId.
type SubscriptionResource struct {
	// ID is the unique subscription identifier returned by POST/GET/PATCH requests.
	ID string `json:"id,omitempty"`
	// ChangeType specifies the event types (created, updated, deleted) to subscribe to.
	ChangeType webhook.ChangeType `json:"changeType,omitempty"`
	// ObjectName uses the clientState field to store the connector's object name for identification.
	ObjectName common.ObjectName `json:"clientState,omitempty"`
	// WebhookURL is the notification URL where Microsoft Graph sends change notifications.
	WebhookURL string `json:"notificationUrl,omitempty"`
	// Resource identifies the Microsoft Graph resource being monitored (e.g., "me/messages").
	Resource string `json:"resource,omitempty"`
	// ExpirationDateTime is the UTC datetime when the subscription expires and auto-deletes.
	// Must respect per-resource maximum lifetimes (ranges from 5 hours to 30 days).
	// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#subscription-lifetime
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	// IncludeResourceData is set to false. This is to avoid encryption requirements.
	// Resource data is fetched separately via ReadByIds, therefore it is not needed.
	IncludeResourceData bool `json:"includeResourceData,omitempty"`
}

// newPayloadCreateSubscription constructs a subscription payload for creation.
// Uses clientState to store objectName for identification.
// Expiration is set to 5 hours to safely fit common maximums (e.g., presence: 1h excluded; others 3-30 days).
//
// nolint:lll
// See [lifetime limits](https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#subscription-lifetime)
func newPayloadCreateSubscription(
	objectName common.ObjectName,
	events common.ObjectEvents,
	webhookURL string,
) SubscriptionResource {
	resource := objectName.String()

	fiveHoursFromNow := time.Now().Add(subscriptionExpirationWindow)
	body := SubscriptionResource{
		ChangeType:          webhook.NewChangeType(events.Events),
		ObjectName:          objectName,
		WebhookURL:          webhookURL,
		Resource:            resource,
		ExpirationDateTime:  fiveHoursFromNow,
		IncludeResourceData: false,
	}

	return body
}
