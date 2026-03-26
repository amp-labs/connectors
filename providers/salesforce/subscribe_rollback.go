package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
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
func (c *Connector) rollbackUpdateSubscription(
	ctx context.Context,
	progress *updateSubscriptionProgress,
) error {
	if !progress.quotaFieldsUpserted || len(progress.newQuotaFields) == 0 {
		return nil
	}

	req := &SubscriptionRequest{QuotaOptimizationObjectFields: progress.newQuotaFields}

	return c.rollbackQuotaOptimizationFields(ctx, req)
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
