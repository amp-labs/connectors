package salesforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// rollbackSubscribe reverses completed operations in reverse order based on progress.
// It removes successfully rolled-back members from the shared createdMembers map
// and returns a SubscriptionResult reflecting the post-rollback state.
//
//nolint:cyclop,funlen
func (c *Connector) rollbackSubscribe(
	ctx context.Context,
	sfRes *SubscribeResult,
	progress *subscribeProgress,
) (*common.SubscriptionResult, error) {
	var rollbackErr error

	// Reverse in the inverse of creation order (fields → triggers → members).
	// Members are torn down first to stop CDC traffic, then triggers, then the
	// custom fields they both reference.
	for objName, member := range progress.createdMembers {
		if member == nil {
			slog.Warn("event channel member entry is nil during rollback, skipping",
				"object", objName,
			)

			continue
		}

		// TODO: check existence before delete
		if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
			slog.Warn("event channel member rollback failed; entry retained in sfRes for operator visibility",
				"object", objName,
				"error", err,
			)

			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf("failed to delete event channel member for object %s: %w", objName, err),
			)
		} else {
			delete(progress.createdMembers, objName)
		}
	}

	for objName, trigger := range progress.deployedTriggers {
		if trigger == nil {
			slog.Warn("apex trigger entry is nil during rollback, skipping",
				"object", objName,
			)

			continue
		}

		// TODO: check existence before delete
		if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
			slog.Warn("apex trigger rollback failed; entry retained in sfRes for operator visibility",
				"object", objName,
				"error", err,
			)

			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf("failed to rollback apex trigger for object %s: %w", objName, err),
			)
		} else {
			delete(progress.deployedTriggers, objName)
			delete(sfRes.ApexTriggers, objName)
		}
	}

	if progress.quotaFieldsUpserted {
		// TODO: check existence before delete
		if err := c.rollbackQuotaOptimizationFields(ctx, progress.req); err != nil {
			// Residual fields left in Salesforce are tolerable: they are inert
			// custom checkbox fields with no behavioral side effects once the
			// referencing apex trigger and channel member are gone. We log the
			// snapshot so an admin can clean them up manually if desired, then
			// continue — sfRes must not keep advertising fields we believe
			// (best-effort) we removed.
			slog.Warn(
				"quota optimization field rollback failed; clearing tracked fields anyway, "+
					"residual fields may remain in Salesforce but are tolerable",
				"error", err,
				"fields", sfRes.QuotaOptimizationObjectFields,
			)

			rollbackErr = errors.Join(rollbackErr, err)
		}
		// Clear regardless of error: any residual fields on the Salesforce side
		// are tolerable, and sfRes should not advertise fields that this rollback
		// attempted to remove.
		clear(sfRes.QuotaOptimizationObjectFields)
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

	// TODO: check existence before delete
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

	// TODO: check existence before delete
	res, err := c.DeleteMetadata(ctx, &common.DeleteMetadataParams{
		Fields: deleteFields,
	})

	if err != nil || res != nil && !res.Success {
		return fmt.Errorf("failed to rollback quota optimization fields: %w", err)
	}

	return nil
}
