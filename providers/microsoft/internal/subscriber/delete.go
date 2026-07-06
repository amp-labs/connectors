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

func (s Strategy) deleteSubscriptionsWithResult(ctx context.Context,
	identifiers []string,
	prevResult *Result,
) (*common.SubscriptionResult, error) {
	subscriptions, objectEvents := prevResult.extractSubscriptionsByIds(identifiers)

	// Attempt removal of the requested subscriptions.
	// If removal fails (non-nil err), the previous state is unchanged,
	// and we return a failed result containing the initial state.
	result, err := s.removeSubscriptionsByIds(ctx, identifiers)
	if err != nil {
		return &common.SubscriptionResult{
			Result:       &Result{Subscriptions: subscriptions},
			ObjectEvents: objectEvents,
			Status:       common.SubscriptionStatusFailed,
		}, err
	}

	// If there were per-item errors, aggregate them into a single error.
	status := common.SubscriptionStatusSuccess

	if len(result.Errors) == 0 {
		// Prune the initial state for every record.
		for _, id := range identifiers {
			objectEvents[subscriptions[id].ObjectName] = common.ObjectEvents{}
			delete(subscriptions, id)
		}
	} else {
		err = nil // must be empty anyway.
		for _, errWrapper := range result.Errors {
			err = errors.Join(err, errWrapper.Data)
		}

		// If aggregation produced an error, mark overall status as failed.
		if err != nil {
			status = common.SubscriptionStatusFailed
		}
	}

	return &common.SubscriptionResult{
		Result:       &Result{Subscriptions: subscriptions},
		ObjectEvents: objectEvents,
		Status:       status,
	}, err
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
