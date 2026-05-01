package subscriber

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// DeleteSubscription removes the subscription for all object mentioned in the SubscriptionResult.
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#delete-a-subscription
func (s Strategy) DeleteSubscription(
	ctx context.Context,
	previousResult common.SubscriptionResult,
) error {
	objectNames := previousResult.ObjectNames()
	// The state is discarded.
	_, err := s.deleteSubscription(ctx, objectNames)

	return err
}

func (s Strategy) deleteSubscription(
	ctx context.Context,
	objectNames []common.ObjectName,
) (*common.SubscriptionResult, error) {
	remoteSubscriptions, err := s.fetchSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	subscriptionsToRemove := make([]SubscriptionID, 0)

	objects := datautils.NewSetFromList(objectNames)
	for name, subscriptions := range remoteSubscriptions {
		if objects.Has(name) {
			for _, subscription := range subscriptions {
				subscriptionsToRemove = append(subscriptionsToRemove, subscription.ID)
			}
		}
	}

	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIDs(subscriptionsToRemove)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)

	// Any non 2xx responses are errors. Among them 404 is acceptable for the delete operation.
	// Others would be considered a genuine errors.
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
