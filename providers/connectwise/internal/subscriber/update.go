package subscriber

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/subscriptionhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// UpdateSubscription updates object subscriptions by comparing previous and desired states,
// then creating new subscriptions and removing obsolete ones as needed.
//
// The update process segments subscription events into four categories:
//   - ToUpdate: Objects where event types differ (e.g., SubscriptionEventType changes).
//     Note: ConnectWise webhook definitions don't have mutable fields to change, so these
//     are effectively treated like ToKeep (no action taken). The final output always includes
//     Create/Update/Delete events regardless of what the user requested.
//   - ToKeep: Objects with identical events in both states. ConnectWise doesn't need
//     webhook refreshing, so no action is taken.
//   - ToCreate: Objects in desired state but not in previous state. New subscriptions
//     are created for these objects.
//   - ToRemove: Objects in previous state but not in desired state. Subscriptions
//     are removed for these objects.
//
// Only ToCreate and ToRemove require actual action. ToUpdate and ToKeep are passive.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - params: Subscription parameters containing the desired subscription events
//   - previousResult: The result from the previous subscription operation (existing state)
//
// Returns:
//   - *common.SubscriptionResult: The result of the update operation containing:
//   - Output.ObjectWebhooks: Combined webhooks after create and remove operations
//   - ObjectEvents: Combined events after create and remove operations
//   - Status: Merged status from create and remove operations (failed_to_rollback or failed if any operation failed)
//   - error: Any error encountered during the update process
func (s Strategy) UpdateSubscription( // nolint:cyclop,funlen
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Segment subscription events into ToCreate, ToKeep, ToUpdate, and ToRemove categories.
	segmentEvents := subscriptionhelper.SegmentSubscriptionEvents(
		previousResult.ObjectEvents, params.SubscriptionEvents,
	)

	// Cast subscription input/output into local typed representations.
	subsInput, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	prevOutput, err := s.TypedSubscriptionResult(*previousResult)
	if err != nil {
		return nil, err
	}

	// Create Event subscriptions for objects that need new subscriptions.
	createSubsResult, err := s.createSubscription(ctx, subsInput, segmentEvents.ToCreate)
	if err != nil {
		return nil, err
	}

	result, ok := createSubsResult.Result.(Result)
	if !ok {
		return nil, fmt.Errorf("%w: common.SubscriptionResult.Result cannot be cast to connectwise.Result",
			common.ErrInvalidImplementation)
	}

	// Merge newly created webhooks and events with existing ones.
	combinedWebhooks := datautils.MergeMaps(prevOutput.ObjectWebhooks, result.ObjectWebhooks)
	combinedEvents := datautils.MergeMaps(previousResult.ObjectEvents, createSubsResult.ObjectEvents)

	// Delete Event subscriptions for objects that need to be removed.
	webhooksToRemove := make([]int, 0)
	for _, webhook := range datautils.FromMap(prevOutput.ObjectWebhooks).ShallowSubset(segmentEvents.ToRemove.Keys()) {
		webhooksToRemove = append(webhooksToRemove, webhook.ID)
	}

	removeResult := s.removeSubscriptionsByIDs(ctx, webhooksToRemove)

	// Track which objects were successfully deleted by matching removed subscription IDs.
	deletedObjects := make(datautils.Set[common.ObjectName])

	for deletedSubID := range removeResult.Records {
		for name, webhook := range combinedWebhooks {
			if webhook.ID == deletedSubID {
				deletedObjects.AddOne(name)
			}
		}
	}

	// Remove deleted objects from combined webhooks and events maps.
	for name := range deletedObjects {
		delete(combinedWebhooks, name)
		delete(combinedEvents, name)
	}

	// Determine remove operation status based on errors.
	removeStatus := common.SubscriptionStatusSuccess
	if len(removeResult.Errors) != 0 {
		removeStatus = common.SubscriptionStatusFailed
	}

	// Merge create and remove statuses to get the final combined status.
	// If either operation failed, the combined status reflects the failure.
	combinedStatus := createSubsResult.Status.Resolve(removeStatus)

	return &common.SubscriptionResult{
		Result:       Result{ObjectWebhooks: combinedWebhooks},
		ObjectEvents: combinedEvents,
		Status:       combinedStatus,
	}, nil
}
