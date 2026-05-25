package salesforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// rollbackSubscribe is the failure-path counterpart to executeSubscribe. It runs
// only when executeSubscribe returned an error, and its job is to undo whatever
// completed steps left state in Salesforce so the caller is not left with a
// half-built subscription.
//
// What it undoes (and the order)
//
// The forward path creates artifacts in this order:
//
//	(A) quota optimization custom fields  →  (C) apex triggers  →  (B) channel members
//
// Both apex triggers and channel members reference the quota custom fields,
// while neither references the other. Cleanup follows the dependency-removal
// rule — dependents are removed before the dependency:
//
//	(B) channel members  →  (C) apex triggers  →  (A) quota custom fields
//
// Tearing channel members down first also stops CDC traffic before the trigger
// that maintains the indicator field is removed, avoiding a window where the
// pipeline emits events with a stale indicator value.
//
// # Progress tracking
//
// rollbackSubscribe consults subscribeProgress (populated by executeSubscribe)
// to know which steps actually completed:
//   - progress.createdMembers   — channel members successfully created
//     (this is the same map as sfRes.EventChannelMembers — they share a
//     reference, so deleting from one updates the other)
//   - progress.deployedTriggers — apex triggers whose Metadata API deploy
//     reported success (filtered from the bulk deploy result)
//   - progress.quotaFieldsUpserted — bool flag set after a successful
//     UpsertMetadata; rollback then iterates progress.req's
//     QuotaOptimizationObjectFields (already filtered by upsert to only
//     subscribed objects)
//
// # Best-effort across the loops
//
// Each per-object delete is independent. Failures don't abort — the function
// joins errors via errors.Join and continues, so the maximum number of
// artifacts get cleaned up even when one fails. Successful deletes are removed
// from progress (and from sfRes for ECMs/triggers) so the returned result
// reflects what is actually still in Salesforce. Failed deletes are retained
// in both progress and sfRes for operator visibility, with a per-object
// slog.Warn so the operator sees which objects need manual cleanup.
//
// Quota-field cleanup is bulk (one DeleteMetadata call). On error we still
// clear sfRes.QuotaOptimizationObjectFields (per project policy: residual
// inert custom fields in Salesforce are tolerable, and sfRes must not
// advertise fields this rollback attempted to remove). The error is recorded
// in rollbackErr and a warn log captures the snapshot for forensic cleanup.
//
// # Returned result
//
// The returned *common.SubscriptionResult wraps sfRes (which has been mutated
// throughout this function to mirror Salesforce). The Status field encodes
// whether the rollback itself succeeded:
//   - SubscriptionStatusFailed            — original Subscribe failed AND
//     rollback successfully undid every completed step.
//   - SubscriptionStatusFailedToRollback  — original Subscribe failed AND
//     rollback could not fully undo. The returned Result lists what's still
//     stranded in Salesforce so the caller can drive a follow-up cleanup.
//
// Subscribe (the public entry point) uses errors.Join to combine execErr and
// rollbackErr before returning to the caller.
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
