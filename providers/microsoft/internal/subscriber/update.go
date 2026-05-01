package subscriber

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// UpdateSubscription  TODO
// By default, Microsoft Graph API does not allow updating the many important field.
// Only the bookkeeping type of fields can be changed.
//
//	> Updates a subscription expiration time for renewal and/or updates the notificationUrl for delivery.
//
// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#methods
func (s Strategy) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Make a READ call to get the remote state.
	// TODO we could rely on previousResult but not sure how to handle the chore of cleaning up excess subs. See below.
	_ = previousResult

	remoteSubscriptions, err := s.fetchSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	plan := remoteSubscriptions.ReconcileTo(params.SubscriptionEvents)

	var (
		createResult, refreshResult, deleteResult *common.SubscriptionResult
		createErr, refreshErr, deleteErr          error
	)

	if len(plan.Create) != 0 {
		createResult, createErr = s.createSubscription(ctx, common.SubscribeParams{
			Request:            params.Request,
			RegistrationResult: params.RegistrationResult, // TODO this might be needed.
			SubscriptionEvents: plan.Create,
		})
	}

	if len(plan.Refresh) != 0 {
		refreshResult, refreshErr = s.refreshSubscription(ctx, plan.Refresh)
	}

	if len(plan.Delete) != 0 {
		deleteResult, deleteErr = s.deleteSubscription(ctx, plan.Delete)
	}

	// TODO !!!!!!!! This must happen only at the end. We must ensure that the creation has succeeded.
	// We would rather have duplicates than to loose any events.
	// Cleanup of extras. This is a chore and does not affect the state as far as the user is concerned.
	// However, since any excess subscriptions will create extra noise, they should be removed.
	// This is due to the following assumption: Multiple webhooks for the same object are not supported.
	// TODO should log error, rather than that the output is irrelevant for our cause.
	// TODO Should this be done as part of the Subscribe action?
	// The bottom line is the subscriptions expire on their own.
	if len(plan.Extra) != 0 {
		_, _ = s.removeSubscriptionsByIDs(ctx, plan.Extra)
	}

	result := combineResults(createResult, refreshResult, deleteResult)
	err = errors.Join(createErr, refreshErr, deleteErr)

	return result, err
}

// TODO now we have the refresh and the replace.
// combineResults merges the results of create, update, and delete operations into a single SubscriptionResult.
// Each input represents a disjoint set of objects -- no object may appear in more
// than one of create, update, or remove. Nil inputs are ignored.
func combineResults(create, update, remove *common.SubscriptionResult) *common.SubscriptionResult {
	var (
		state     = make(State)
		checklist = make(datautils.Set[common.SubscriptionStatus])
	)

	if create != nil {
		state.Add(create.ObjectEvents)
		checklist.AddOne(create.Status)
	}

	if update != nil {
		state.Add(update.ObjectEvents)
		checklist.AddOne(update.Status)
	}

	if remove != nil {
		state.Add(remove.ObjectEvents)
		checklist.AddOne(remove.Status)
	}

	// Derives a final Status using pessimistic precedence, where the worst outcome dominates.
	var status common.SubscriptionStatus
	if _, ok := checklist[common.SubscriptionStatusFailedToRollback]; ok {
		status = common.SubscriptionStatusFailedToRollback
	} else if _, ok = checklist[common.SubscriptionStatusFailed]; ok {
		status = common.SubscriptionStatusFailed
	} else if _, ok = checklist[common.SubscriptionStatusSuccess]; ok {
		status = common.SubscriptionStatusSuccess
	}

	return &common.SubscriptionResult{
		Result:       Output{},
		ObjectEvents: state,
		Status:       status,
	}
}

func (s Strategy) refreshSubscription(
	ctx context.Context,
	refreshPlan map[common.ObjectName]SubscriptionID,
) (*common.SubscriptionResult, error) {
	batchParams, err := s.paramsForBatchRefreshSubscriptions(refreshPlan)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)
	state := getStateFromCreateResponse(bundledResponse)

	status := common.SubscriptionStatusSuccess
	if len(bundledResponse.Errors) != 0 {
		// Some requests have failed. No rollback for the refresh.
		// The state itself must still be the same.
		status = common.SubscriptionStatusFailed
	}

	return &common.SubscriptionResult{
		Result:       Output{},
		ObjectEvents: state,
		Status:       status,
	}, nil
}

func (s Strategy) paramsForBatchRefreshSubscriptions(
	refreshPlan map[common.ObjectName]SubscriptionID,
) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for objectName, subscriptionID := range refreshPlan {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(string(subscriptionID))

		requestID := batch.RequestID(objectName)
		body := newPayloadRefreshSubscription(s.clock)
		batchParams.WithRequest(requestID, http.MethodPatch, url, body, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
