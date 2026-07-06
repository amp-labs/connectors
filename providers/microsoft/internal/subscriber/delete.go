package subscriber

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// DeleteSubscription removes all remote subscriptions for objects specified in previousResult.
// Executes batch DELETE requests to Microsoft Graph, tolerating 404s (already deleted) as success.
// Discards final state as delete operations are fire-and-forget; returns first error if any.
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#delete-a-subscription
func (s Strategy) DeleteSubscription(
	ctx context.Context,
	previousResult common.SubscriptionResult,
) error {
	subscriptionResult, err := s.TypedSubscriptionResult(previousResult)
	if err != nil {
		return err
	}

	subscriptionsToRemove := make([]string, 0, len(subscriptionResult.Subscriptions))
	for subId, sub := range subscriptionResult.Subscriptions {
		if events, ok := previousResult.ObjectEvents[sub.ObjectName]; ok {
			if len(events.Events) == 0 {
				// Subscription for this object should be removed.
				subscriptionsToRemove = append(subscriptionsToRemove, subId)
			}
		} else {
			// The object not found in ObjectEvents should be removed.
			subscriptionsToRemove = append(subscriptionsToRemove, subId)
		}
	}

	if len(subscriptionsToRemove) == 0 {
		return nil
	}

	result, err := s.removeSubscriptionsByIds(ctx, subscriptionsToRemove)
	if err != nil {
		return err
	}

	if len(result.Errors) != 0 {
		err = nil
		for _, errWrapper := range result.Errors {
			err = errors.Join(err, errWrapper.Data)
		}

		return err
	}

	return nil
}

// removeSubscriptionsByIds executes batch DELETE for given subscription IDs.
// Generic helper; returns raw batch result for custom error handling.
func (s Strategy) removeSubscriptionsByIds(
	ctx context.Context, ids []string,
) (*batch.Result[any], error) {
	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIds(ids)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[any](ctx, s.batchStrategy, batchParams)

	return bundledResponse, nil
}

// paramsForBatchRemoveSubscriptionsByIds creates batch parameters for DELETE `/subscriptions/{id}` requests.
func (s Strategy) paramsForBatchRemoveSubscriptionsByIds(ids []string) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, identifier := range ids {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(identifier)

		// RequestID is Subscription identifier.
		batchParams.WithRequest(batch.RequestID(identifier), http.MethodDelete, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
