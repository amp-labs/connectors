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
	remoteSubscriptions, err := s.fetchSubscriptions(ctx)
	if err != nil {
		return err
	}

	subscriptionsForRemoval := remoteSubscriptions.subset(previousResult.ObjectNames())

	// The state is discarded.
	_, err = s.deleteSubscriptions(ctx, subscriptionsForRemoval)

	return err
}

// deleteSubscriptions batch-deletes all subscriptions in the given RemoteSubscriptions set.
// Handles 404 responses as success (already deleted); aggregates other errors.
//
// Returns updated SubscriptionResult reflecting remaining subscriptions and operation status.
func (s Strategy) deleteSubscriptions(
	ctx context.Context,
	remoteSubscriptions RemoteSubscriptions,
) (*common.SubscriptionResult, error) {
	subscriptionsToRemove := remoteSubscriptions.getIDs()

	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIDs(subscriptionsToRemove)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)

	// Aggregate non-404 errors (404 indicates record was already deleted).
	var outErr error

	for _, e := range bundledResponse.Errors {
		if e.Status != http.StatusNotFound {
			// Resource no longer exists.
		} else {
			outErr = errors.Join(outErr, e.Data)
		}
	}

	// Filter out the subscription that were removed if any are left .
	for requestID, resp := range bundledResponse.Responses {
		name := ObjectName(resp.Data.Resource)
		RemoteSubsType(remoteSubscriptions).Remove(name, func(subscription SubscriptionResource) bool {
			return subscription.ID == SubscriptionID(requestID)
		})
	}

	status := common.SubscriptionStatusSuccess
	if outErr != nil {
		// The rollback does not happen for delete. So it is either Success or Failure.
		status = common.SubscriptionStatusFailed
	}

	return &common.SubscriptionResult{
		Result:       Output{},
		ObjectEvents: remoteSubscriptions.toState(),
		Status:       status,
	}, outErr
}

// removeSubscriptionsByIDs executes batch DELETE for given subscription IDs.
// Generic helper; returns raw batch result for custom error handling.
func (s Strategy) removeSubscriptionsByIDs(
	ctx context.Context, identifiers []SubscriptionID,
) (*batch.Result[any], error) {
	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIDs(identifiers)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[any](ctx, s.batchStrategy, batchParams)

	return bundledResponse, nil
}

// paramsForBatchRemoveSubscriptionsByIDs creates batch parameters for DELETE `/subscriptions/{id}` requests.
func (s Strategy) paramsForBatchRemoveSubscriptionsByIDs(identifiers []SubscriptionID) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, identifier := range identifiers {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(string(identifier))

		// RequestID is Subscription identifier.
		batchParams.WithRequest(batch.RequestID(identifier), http.MethodDelete, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
