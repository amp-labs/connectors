package subscriber

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
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

	subscriptionsToRemove := datautils.FromMap(subscriptionResult.Subscriptions).Keys()
	if len(subscriptionsToRemove) == 0 {
		return nil
	}

	result, err := s.removeSubscriptionsByIDs(ctx, subscriptionsToRemove)
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

// removeSubscriptionsByIDs executes batch DELETE for given subscription IDs.
// Generic helper; returns raw batch result for custom error handling.
func (s Strategy) removeSubscriptionsByIDs(
	ctx context.Context, identifiers []string,
) (*batch.Result[any], error) {
	batchParams, err := s.paramsForBatchRemoveSubscriptionsByIDs(identifiers)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[any](ctx, s.batchStrategy, batchParams)

	return bundledResponse, nil
}

// paramsForBatchRemoveSubscriptionsByIDs creates batch parameters for DELETE `/subscriptions/{id}` requests.
func (s Strategy) paramsForBatchRemoveSubscriptionsByIDs(identifiers []string) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, identifier := range identifiers {
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
