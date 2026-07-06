package subscriber

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/subscriptionhelper"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// UpdateSubscription reconciles existing remote subscriptions with the desired state from params.
// Microsoft Graph does not support full subscription updates, so UpdateSubscription performs
// targeted create/refresh/delete operations:
//   - CREATE subscriptions for missing objects
//   - REFRESH (PATCH expirationDateTime/notificationUrl) subscriptions that are expiring
//   - DELETE undesired objects
//   - CREATE/DELETE to imitate update.
//
// It returns a merged SubscriptionResult representing the final state after all operations.
//
// Microsoft Graph limitation: Only expirationDateTime and notificationUrl are PATCHable.
// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#methods
func (s Strategy) UpdateSubscription( // nolint:cyclop,funlen
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Partition subscription events into ToCreate, ToKeep, ToUpdate, and ToRemove categories.
	segmentEvents := subscriptionhelper.SegmentSubscriptionEvents(
		previousResult.ObjectEvents, params.SubscriptionEvents,
	)

	// Microsoft Graph API does not support updating subscriptions.
	// We simulate an update by creating a new subscription and removing the outdated one.
	for name, events := range segmentEvents.ToUpdate {
		segmentEvents.ToRemove[name] = common.ObjectEvents{}
		segmentEvents.ToCreate[name] = events
	}

	subsReq, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	prevRes, err := s.TypedSubscriptionResult(*previousResult)
	if err != nil {
		return nil, err
	}

	var (
		createResult, refreshResult, deleteResult *common.SubscriptionResult
		createErr, refreshErr, deleteErr          error
	)

	if len(segmentEvents.ToCreate) != 0 {
		createResult, createErr = s.createSubscription(ctx, common.SubscribeParams{
			Request:            params.Request,
			RegistrationResult: params.RegistrationResult,
			SubscriptionEvents: segmentEvents.ToCreate,
		})
	}

	if len(segmentEvents.ToKeep) != 0 {
		subsToRefresh := make([]string, 0)

		for _, sub := range prevRes.Subscriptions {
			if segmentEvents.ToKeep.Has(sub.ObjectName) {
				subsToRefresh = append(subsToRefresh, sub.ID)
			}
		}

		if len(subsToRefresh) != 0 {
			refreshResult, refreshErr = s.refreshSubscription(ctx, subsToRefresh, prevRes, subsReq.WebhookURL)
		} else {
			refreshErr = fmt.Errorf("%w: no subscriptions to refresh", common.ErrPrevSubscriptionResultInvalid)
		}
	}

	if len(segmentEvents.ToRemove) != 0 {
		subsToRemove := make([]string, 0)

		for _, sub := range prevRes.Subscriptions {
			if segmentEvents.ToRemove.Has(sub.ObjectName) {
				subsToRemove = append(subsToRemove, sub.ID)
			}
		}

		if len(subsToRemove) != 0 {
			deleteResult, deleteErr = s.deleteSubscriptionsWithResult(ctx, subsToRemove, prevRes)
		} else {
			deleteErr = fmt.Errorf("%w: no subscriptions to remove", common.ErrPrevSubscriptionResultInvalid)
		}
	}

	result, fatalErr := combineResults(createResult, refreshResult, deleteResult)
	if fatalErr != nil {
		return nil, fatalErr
	}

	err = errors.Join(createErr, refreshErr, deleteErr)

	return result, err
}

// combineResults merges create, refresh, and remove subscription results into a single SubscriptionResult.
//
// Semantics:
//   - refresh is disjoint from both create and remove.
//   - create and remove are not necessarily disjoint.
//   - an object appearing in both create and remove represents an update
//     (remove old subscription + create new subscription).
//
// Merge order is significant:
//   - create and refresh are applied first and are authoritative.
//   - remove is applied only for objects not already present in the merged state.
//
// This ensures updated objects are preserved and not overwritten by stale remove entries.
// Objects present only in remove represent truly deleted subscriptions and will have an empty event list.
//
// Nil inputs are ignored.
func combineResults( // nolint:cyclop,funlen
	create, refresh, remove *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	var (
		status                = common.SubscriptionStatusSuccess
		combinedEvents        = make(map[common.ObjectName]common.ObjectEvents)
		combinedSubscriptions = make(map[string]SubscriptionResource)
	)

	if create != nil {
		status = status.Resolve(create.Status)
		maps.Copy(combinedEvents, create.ObjectEvents)

		result, ok := create.Result.(*Result)
		if !ok {
			return nil, fmt.Errorf("%w: create.Result cannot be cast to microsoft.Result",
				common.ErrInvalidImplementation)
		}

		maps.Copy(combinedSubscriptions, result.Subscriptions)
	}

	if refresh != nil {
		status = status.Resolve(refresh.Status)
		maps.Copy(combinedEvents, refresh.ObjectEvents)

		result, ok := refresh.Result.(*Result)
		if !ok {
			return nil, fmt.Errorf("%w: refresh.Result cannot be cast to microsoft.Result",
				common.ErrInvalidImplementation)
		}

		maps.Copy(combinedSubscriptions, result.Subscriptions)
	}

	if remove != nil {
		status = status.Resolve(remove.Status)

		// Apply remove entries only for objects that were truly deleted.
		// Objects already present in the state were recreated as part of an update
		// flow (create + remove) and must be preserved.
		for objectName, events := range remove.ObjectEvents {
			if _, exists := combinedEvents[objectName]; !exists {
				combinedEvents[objectName] = events
			}
		}

		result, ok := remove.Result.(*Result)
		if !ok {
			return nil, fmt.Errorf("%w: remove.Result cannot be cast to microsoft.Result",
				common.ErrInvalidImplementation)
		}

		for subId, sub := range result.Subscriptions {
			if _, exists := combinedSubscriptions[subId]; !exists {
				combinedSubscriptions[subId] = sub
			}
		}
	}

	return &common.SubscriptionResult{
		Result:       &Result{Subscriptions: combinedSubscriptions},
		ObjectEvents: combinedEvents,
		Status:       status,
	}, nil
}

// refreshSubscription batch-PATCHes expirationDateTime and notificationUrl for subscriptions nearing expiry.
func (s Strategy) refreshSubscription(ctx context.Context,
	identifiers []string,
	prevResult *Result,
	webhookURL string,
) (*common.SubscriptionResult, error) {
	batchParams, err := s.paramsForBatchRefreshSubscriptions(identifiers, webhookURL)
	if err != nil {
		return nil, err
	}

	bundledResponse := batch.Execute[SubscriptionResource](ctx, s.batchStrategy, batchParams)

	status := common.SubscriptionStatusSuccess
	if len(bundledResponse.Errors) != 0 {
		// Some requests failed. There is no rollback for refresh operations.
		// The state must remain unchanged.
		status = common.SubscriptionStatusFailed

		err = nil
		for _, wrapper := range bundledResponse.Errors {
			err = errors.Join(err, wrapper.Data)
		}
	}

	subscriptions, objectEvents := prevResult.extractSubscriptionsByIds(identifiers)

	return &common.SubscriptionResult{
		Result:       &Result{Subscriptions: subscriptions},
		ObjectEvents: objectEvents,
		Status:       status,
	}, err
}

// paramsForBatchRefreshSubscriptions builds PATCH /subscriptions/{id} requests for subscription renewal.
func (s Strategy) paramsForBatchRefreshSubscriptions(
	identifiers []string,
	webhookURL string,
) (*batch.Params, error) {
	batchParams := &batch.Params{}

	for _, subscriptionId := range identifiers {
		url, err := s.getSubscriptionURL()
		if err != nil {
			return nil, err
		}

		url.AddPath(subscriptionId)

		requestId := batch.RequestID(subscriptionId)
		body := newPayloadRefreshSubscription(webhookURL)
		batchParams.WithRequest(requestId, http.MethodPatch, url, body, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}
