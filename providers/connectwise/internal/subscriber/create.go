package subscriber

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/webhook"
)

// maxConcurrency is the maximum number of goroutines that will run
// concurrently when creating subscriptions. The value is chosen arbitrarily.
const maxConcurrency = 3

func (s Strategy) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	request, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	return s.createSubscription(ctx, request, params.SubscriptionEvents)
}

func (s Strategy) createSubscription(ctx context.Context,
	request Request,
	subscriptionEvents datautils.Map[common.ObjectName, common.ObjectEvents],
) (*common.SubscriptionResult, error) {
	url, err := s.getSubscriptionURL()
	if err != nil {
		return nil, err
	}

	subscriptionURL := url.String()

	tasks := make([]parallelfetch.Task[common.ObjectName, SubscriptionResource], len(subscriptionEvents))
	index := 0

	for objectName := range subscriptionEvents {
		tasks[index], err = s.newTaskCreateSubscription(objectName, subscriptionURL, request)
		if err != nil {
			return nil, err
		}

		index += 1
	}

	createResult := parallelfetch.Execute(ctx, tasks, maxConcurrency)
	state := getStateFromCreateResponse(createResult)

	if len(createResult.Errors) != 0 {
		// Some requests failed; initiate rollback.
		return s.rollbackSubscriptionCreation(ctx, subscriptionEvents, state, createResult)
	}

	return &common.SubscriptionResult{
		Result: Result{
			ObjectWebhooks: createResult.Records,
		},
		ObjectEvents: state,
		Status:       common.SubscriptionStatusSuccess,
	}, nil
}

// getStateFromCreateResponse extracts the subscription state from batch responses.
// Successful responses set the events state; errors result in empty ObjectEvents state.
func getStateFromCreateResponse(
	subscriptions parallelfetch.Result[common.ObjectName, SubscriptionResource],
) map[common.ObjectName]common.ObjectEvents {
	result := make(map[common.ObjectName]common.ObjectEvents)

	for _, subscription := range subscriptions.Records {
		// Map successful subscription to its events.
		objectName := webhook.ObjectTypeToObjectName[subscription.ObjectType]
		result[objectName] = common.ObjectEvents{
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
				common.SubscriptionEventTypeDelete,
			},
			WatchFields:       nil,
			WatchFieldsAll:    false,
			PassThroughEvents: nil,
		}
	}

	for objectName := range subscriptions.Errors {
		// Failed requests yield no subscription.
		result[objectName] = common.ObjectEvents{}
	}

	return result
}

// rollbackSubscriptionCreation deletes successfully created subscriptions on partial failure.
// It updates state to reflect remaining subscriptions after attempted rollback.
// Returns appropriate status based on rollback success.
func (s Strategy) rollbackSubscriptionCreation(
	ctx context.Context,
	subscriptionEvents datautils.Map[common.ObjectName, common.ObjectEvents],
	state map[common.ObjectName]common.ObjectEvents,
	partialCreation parallelfetch.Result[common.ObjectName, SubscriptionResource],
) (*common.SubscriptionResult, error) {
	requestsRegistry := make(datautils.Map[int, common.ObjectName])
	for objectName, subscription := range partialCreation.Records {
		requestsRegistry[subscription.ID] = objectName
	}

	removeResult := s.removeSubscriptionsByIDs(ctx, requestsRegistry.Keys())
	if len(removeResult.Errors) == 0 {
		// Full rollback succeeded.
		objectNames := datautils.FromMap(subscriptionEvents).Keys()

		events := make(map[common.ObjectName]common.ObjectEvents)
		for _, objectName := range objectNames {
			events[objectName] = common.ObjectEvents{}
		}

		return &common.SubscriptionResult{
			Result: Result{
				ObjectWebhooks: make(map[common.ObjectName]SubscriptionResource),
			},
			ObjectEvents: events,
			Status:       common.SubscriptionStatusFailed,
		}, nil
	}

	// Partial rollback; track remaining subscriptions.
	existingObjects := datautils.NewSet[common.ObjectName]()

	for requestID := range removeResult.Errors {
		objectName := requestsRegistry[requestID]
		existingObjects.AddOne(objectName)
	}

	allObjects := datautils.FromMap(subscriptionEvents).KeySet()
	removedObjects := allObjects.Subtract(existingObjects)

	// Clear state for successfully removed objects.
	for _, objectName := range removedObjects {
		state[objectName] = common.ObjectEvents{}
	}

	remainingWebhooks := make(map[common.ObjectName]SubscriptionResource)
	for name := range existingObjects {
		remainingWebhooks[name] = partialCreation.Records[name]
	}

	return &common.SubscriptionResult{
		Result: Result{
			ObjectWebhooks: remainingWebhooks,
		},
		ObjectEvents: state,
		Status:       common.SubscriptionStatusFailedToRollback,
	}, nil
}

func (s Strategy) newTaskCreateSubscription(objectName common.ObjectName,
	url string,
	request Request,
) (parallelfetch.Task[common.ObjectName, SubscriptionResource], error) {
	objectType, found := webhook.ObjectNameToObjectType[objectName]
	if !found {
		return nil, fmt.Errorf("%w: cannot subscribe to '%v' object", common.ErrObjectNotSupported, objectName)
	}

	return func(ctx context.Context) (taskID common.ObjectName, data *SubscriptionResource, err error) {
		body := SubscriptionResource{
			// Webhook must end with recordId without the value.
			// The ConnectWise will append the record identifier as a raw string when sending the event.
			WebhookURL:     request.WebhookURL + "?recordId=",
			ObjectType:     objectType,
			ObjectLevel:    "owner",
			PayloadVersion: messageVersion,
		}

		resp, err := s.client.Post(ctx, url, body, s.clientIdHeader())
		if err != nil {
			return objectName, nil, err
		}

		subscription, err := common.UnmarshalJSON[SubscriptionResource](resp)
		if err != nil {
			return objectName, nil, err
		}

		return objectName, subscription, nil
	}, nil
}

// SubscriptionResource models a ConnectWise callback (subscription) returned by
// the /system/callbacks endpoint. The ConnectWise API refers to these as
// "CallbackEntries" and the same JSON shape is returned for GET/POST/PATCH/PUT.
//
// Reference: https://developer.connectwise.com/Products/ConnectWise_PSA/REST#/CallbackEntries
type SubscriptionResource struct {
	// Callback properties.
	ID           int    `json:"id"`
	InactiveFlag bool   `json:"inactiveFlag"`
	WebhookURL   string `json:"url"`

	// Object configurations.
	// ObjectID is required even if we send 0 for the owner.
	// When ObjectLevel is not "owner", you typically need to specify
	// a parent ID or context ID to subscribe the object to.
	ObjectID   int    `json:"objectId"`
	ObjectType string `json:"type"`
	// ObjectLevel specifies how granular the subscription should be.
	// It can be used alongside ObjectID, but the connector doesn't support that.
	ObjectLevel    string `json:"level"`
	PayloadVersion string `json:"payloadVersion"`

	// Miscellaneous properties.
	// MemberID is the person who created the callback.
	MemberID int `json:"memberId"`
	// IsSelfSuppressedFlag indicates whether the callback creator receives messages.
	IsSelfSuppressedFlag bool   `json:"isSelfSuppressedFlag"`
	ConnectWiseID        string `json:"connectWiseID"`
}
