package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

// rollbackSubscribe reverses completed operations in reverse order based on progress.
// It removes successfully rolled-back members from the shared createdMembers map
// and returns a SubscriptionResult reflecting the post-rollback state.
func (c *Connector) rollbackSubscribe(
	ctx context.Context,
	sfRes *SubscribeResult,
	progress *subscribeProgress,
) (*common.SubscriptionResult, error) {
	var rollbackErr error

	// Reverse deployed apex triggers (last completed, first to rollback).
	for objName, trigger := range progress.deployedTriggers {
		if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf("failed to rollback apex trigger for object %s: %w", objName, err),
			)
		} else {
			delete(progress.deployedTriggers, objName)
		}
	}

	// Reverse created event channel members.
	for objName, member := range progress.createdMembers {
		if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf("failed to delete event channel member for object %s: %w", objName, err),
			)
		} else {
			delete(progress.createdMembers, objName)
		}
	}

	// Reverse quota optimization fields.
	if progress.quotaFieldsUpserted {
		if err := c.rollbackQuotaOptimizationFields(ctx, progress.req); err != nil {
			rollbackErr = errors.Join(rollbackErr, err)
		}
	}

	res := &common.SubscriptionResult{
		Result: sfRes,
	}

	if rollbackErr != nil {
		res.Status = common.SubscriptionStatusFailedToRollback
		res.Events = []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		}

		for objName := range sfRes.EventChannelMembers {
			res.Objects = append(res.Objects, objName)
		}
	} else {
		res.Status = common.SubscriptionStatusFailed
	}

	return res, rollbackErr
}

// rollbackUpdateSubscription reverses completed operations based on progress.
//
// Filter expressions on the recreated members are cleared FIRST so the new
// members no longer hold a metadata reference to the new quota fields. The
// quota field delete that follows is best-effort: if Salesforce still blocks
// it (for example because a redeployed apex trigger references the field),
// we log and continue rather than fail the rollback — leaving the orphan
// custom field is acceptable.
func (c *Connector) rollbackUpdateSubscription(
	ctx context.Context,
	progress *updateSubscriptionProgress,
) error {
	var rollbackErr error

	if err := c.clearRecreatedChannelMemberFilters(ctx, progress); err != nil {
		rollbackErr = errors.Join(rollbackErr, err)
	}

	if progress.quotaFieldsUpserted && len(progress.newQuotaFields) > 0 {
		req := &SubscriptionRequest{QuotaOptimizationObjectFields: progress.newQuotaFields}
		if err := c.rollbackQuotaOptimizationFields(ctx, req); err != nil {
			// Intentionally not joined into rollbackErr: leaving an orphan
			// quota field is preferable to surfacing a rollback failure that
			// the caller can't easily act on.
			logging.Logger(ctx).Warn("rollback: leaving quota optimization fields undeleted",
				"error", err)
		}
	}

	return rollbackErr
}

// clearRecreatedChannelMemberFilters PATCHes each recreated channel member to
// drop its FilterExpression and EnrichedFields. This releases the metadata
// reference on the new quota fields so the field delete in the next rollback
// step can succeed. Members are left in place with a no-op filter; the user
// can re-run UpdateSubscription to restore quota optimization.
func (c *Connector) clearRecreatedChannelMemberFilters(
	ctx context.Context,
	progress *updateSubscriptionProgress,
) error {
	if len(progress.recreatedMembers) == 0 {
		return nil
	}

	var clearErr error

	for objName, member := range progress.recreatedMembers {
		member.Metadata.FilterExpression = ""
		member.Metadata.EnrichedFields = nil

		if _, err := c.UpdateEventChannelMember(ctx, member); err != nil {
			clearErr = errors.Join(clearErr,
				fmt.Errorf("failed to clear filter expression on channel member for object %s: %w", objName, err))

			continue
		}

		delete(progress.recreatedMembers, objName)
	}

	return clearErr
}

func (c *Connector) rollbackQuotaOptimizationFields(ctx context.Context, req *SubscriptionRequest) error {
	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return nil
	}

	deleteFields := make(map[common.ObjectName][]string)

	for objectName, fieldName := range req.QuotaOptimizationObjectFields {
		deleteFields[objectName] = append(
			deleteFields[objectName], customFieldAPIName(fieldName),
		)
	}

	res, err := c.DeleteMetadata(ctx, &common.DeleteMetadataParams{
		Fields: deleteFields,
	})

	if err != nil || res != nil && !res.Success {
		return fmt.Errorf("failed to rollback quota optimization fields: %w", err)
	}

	return nil
}
