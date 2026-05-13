package subscriber

import (
	"context"
	"errors"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// UpdateSubscription reconciles current remote subscriptions with new desired state from params.
// Since Microsoft Graph lacks full subscription updates, it performs targeted create/refresh/delete operations:
// - CREATE missing subscriptions
// - REFRESH (PATCH expirationDateTime/notificationUrl) expiring ones
// - DELETE undesried objects
// - DELETE extra subscriptions ensuring only 1 subscription exists for 1 object.
//
// Returns merged SubscriptionResult reflecting achieved state across all operations.
//
// Microsoft Graph limitation: Only expirationDateTime and notificationUrl are PATCHable.
// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#methods
func (s Strategy) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Previous result is not used.
	// We must fetch the remote subscriptions anyway to know what Subscription IDs to remove.
	_ = previousResult

	input, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

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
			RegistrationResult: params.RegistrationResult,
			SubscriptionEvents: plan.Create,
		})
	}

	if len(plan.Refresh) != 0 {
		refreshResult, refreshErr = s.refreshSubscription(ctx, plan.Refresh, input.WebhookURL)
	}

	if len(plan.DeleteSubscriptions) != 0 {
		deleteResult, deleteErr = s.deleteSubscriptions(ctx, plan.DeleteSubscriptions)
	}

	// Cleanup extra subscriptions to ensure 1 subscription exists for 1 object.
	// This step is done at the very end of this function and if it fails it is not a big deal.
	// Duplicate events will be sent. This is better than to miss data.
	if len(plan.Extra) != 0 {
		_, _ = s.removeSubscriptionsByIDs(ctx, plan.Extra)
	}

	result := combineResults(createResult, refreshResult, deleteResult)
	err = errors.Join(createErr, refreshErr, deleteErr)

	return result, err
}

// combineResults merges create, refresh, and remove subscription results into a single SubscriptionResult.
//
// Semantics:
//   - refresh is fully disjoint from both create and remove.
//   - create and remove are not necessarily disjoint.
//   - an object appearing in both create and remove represents an update
//     (remove old object + create new object).
//
// Merge order is therefore significant:
//   - create and refresh are applied first and are considered authoritative.
//   - remove is applied only for objects not already present in the merged state.
//
// This ensures that updated objects are preserved and not overwritten by stale
// remove entries. Objects present only in remove represent truly deleted subscriptions
// and will have an empty event list.
//
// Nil inputs are ignored.
func combineResults(create, refresh, remove *common.SubscriptionResult) *common.SubscriptionResult {
	var (
		state     = make(State)
		checklist = make(datautils.Set[common.SubscriptionStatus])
	)

	if create != nil {
		maps.Copy(state, create.ObjectEvents)
		checklist.AddOne(create.Status)
	}

	if refresh != nil {
		maps.Copy(state, refresh.ObjectEvents)
		checklist.AddOne(refresh.Status)
	}

	if remove != nil {
		// Remove entries are only applied for objects that were truly deleted.
		// Objects already present in state were recreated as part of an update
		// flow (create + remove) and therefore must be preserved.
		for objectName, events := range remove.ObjectEvents {
			if _, exists := state[objectName]; !exists {
				state[objectName] = events
			}
		}
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

// refreshSubscription batch-PATCHes expirationDateTime/notificationUrl for subscriptions nearing expiry.
func (s Strategy) refreshSubscription(
	ctx context.Context,
	refreshPlan map[common.ObjectName]SubscriptionID,
	webhookURL string,
) (*common.SubscriptionResult, error) {
	batchParams, err := s.paramsForBatchRefreshSubscriptions(refreshPlan, webhookURL)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)
	state := getStateFromCreateResponse(bundledResponse)

	status := common.SubscriptionStatusSuccess
	if len(bundledResponse.Errors) != 0 {
		// Some requests have failed. No rollback for the refresh.
		// The state must still be the same.
		status = common.SubscriptionStatusFailed
	}

	return &common.SubscriptionResult{
		Result:       Output{},
		ObjectEvents: state,
		Status:       status,
	}, nil
}

// paramsForBatchRefreshSubscriptions creates PATCH /subscriptions/{id} requests for renewal.
func (s Strategy) paramsForBatchRefreshSubscriptions(
	refreshPlan map[common.ObjectName]SubscriptionID,
	webhookURL string,
) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for objectName, subscriptionID := range refreshPlan {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(string(subscriptionID))

		requestID := batch.RequestID(objectName)
		body := newPayloadRefreshSubscription(s.clock, webhookURL)
		batchParams.WithRequest(requestID, http.MethodPatch, url, body, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
