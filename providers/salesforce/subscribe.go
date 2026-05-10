package salesforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/go-playground/validator"
)

type SubscribeResult struct {
	EventChannelMembers map[common.ObjectName]*EventChannelMember
	// QuotaOptimizationObjectFields maps object names to custom checkbox field names that we create
	// on the object in Salesforce. These fields are not native to Salesforce — they are custom fields
	// managed by Ampersand to flag whether a record should trigger webhook messages via CDC (Change Data Capture).
	// By checking this field, we can selectively control which records produce CDC events,
	// reducing unnecessary webhook traffic and optimizing API quota usage.
	// We need to return this field and will be read by the DeleteSubscription method to delete the custom fields.
	QuotaOptimizationObjectFields map[common.ObjectName]string
	ApexTriggers                  map[common.ObjectName]*ApexTrigger

	// ManualCheckboxCreation mirrors SubscriptionRequest.ManualCheckboxCreation.
	// When true, the connector skipped creating quota-optimization checkbox fields
	// at subscribe time (after verifying they already existed) and DeleteSubscription
	// must skip deleting them so the caller-managed artifacts are not touched.
	ManualCheckboxCreation bool

	// ManualApexTriggerCreation mirrors SubscriptionRequest.ManualApexTriggerCreation.
	// When true, the connector skipped deploying apex triggers at subscribe time
	// (after verifying they already existed) and DeleteSubscription must skip
	// destructive deletes so the caller-managed triggers are not touched.
	ManualApexTriggerCreation bool
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscribeResult{},
	}
}

type SubscriptionRequest struct {
	// QuotaOptimizationObjectFields maps object names to custom checkbox field names to create
	// on the object in Salesforce. See SubscribeResult.QuotaOptimizationObjectFields for details.
	// The checkbox field is also used to build a CDC filter expression that allows all non-UPDATE
	// events through and only passes UPDATE events where the checkbox is true.
	QuotaOptimizationObjectFields map[common.ObjectName]string

	// ManualCheckboxCreation indicates that the caller wants to create the checkbox field manually.
	ManualCheckboxCreation bool

	// ManualApexTriggerCreation indicates that the caller wants to create the apex trigger manually.
	ManualApexTriggerCreation bool
}

// subscribeProgress tracks which reversible operations completed during executeSubscribe,
// so that rollbackSubscribe knows what to undo.
type subscribeProgress struct {
	quotaFieldsUpserted bool
	req                 *SubscriptionRequest
	createdMembers      map[common.ObjectName]*EventChannelMember
	deployedTriggers    map[common.ObjectName]*ApexTriggerResult
}

// Subscribe creates a Salesforce CDC subscription for the given objects, performing
// up to three operations in order:
//
//  1. Upsert quota optimization custom fields — only when the request configures
//     them via SubscriptionRequest.QuotaOptimizationObjectFields. Skipped entirely
//     when no quota config is provided.
//  2. Deploy apex triggers that maintain those fields per object — only for
//     objects that have both a quota field configured (step 1) and WatchFields
//     set in SubscriptionEvents. Skipped when no objects qualify.
//  3. Create PlatformEventChannelMember records bound to the registered channel.
//     The filter expression and enriched fields are populated only for objects
//     with a quota field configured; otherwise the member is created without them.
//
// Any failure triggers a rollback that reverses completed steps in inverse order.
// On success, returns a SubscriptionResult with Status = Success. On failure with
// successful rollback, returns post-rollback state with Status = Failed. On failure
// with failed rollback, returns the partial state with Status = FailedToRollback;
// the caller should inspect Result to see what survived.
//
// Registration is required prior to subscribing.
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if params.RegistrationResult == nil {
		return nil, fmt.Errorf("%w: missing RegistrationResult", errMissingParams)
	}

	if params.RegistrationResult.Result == nil {
		return nil, fmt.Errorf("%w: missing RegistrationResult.Result", errMissingParams)
	}

	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		return nil, fmt.Errorf("invalid registration result: %w", err)
	}

	registrationParams, registrationOk := params.RegistrationResult.Result.(*ResultData)
	if !registrationOk {
		return nil, fmt.Errorf(
			"%w: expected SubscribeParams.RegistrationResult.Result to be type '%T', but got '%T'", errInvalidRequestType,
			registrationParams,
			params.RegistrationResult.Result,
		)
	}

	var req *SubscriptionRequest

	if params.Request != nil {
		var requestOk bool

		req, requestOk = params.Request.(*SubscriptionRequest)
		if !requestOk {
			return nil, fmt.Errorf(
				"%w: expected SubscribeParams.Request to be type '%T', but got '%T'", errInvalidRequestType,
				req, params.Request,
			)
		}
	}

	sfRes, progress, execErr := c.executeSubscribe(ctx, params, registrationParams, req)
	if execErr != nil {
		rollbackRes, rollbackErr := c.rollbackSubscribe(ctx, sfRes, progress)

		return rollbackRes, errors.Join(execErr, rollbackErr)
	}

	return &common.SubscriptionResult{
		Status: common.SubscriptionStatusSuccess,
		Result: sfRes,
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
		Objects: datautils.FromMap(sfRes.EventChannelMembers).Keys(),
	}, nil
}

// executeSubscribe performs the forward-path logic of Subscribe.
// It returns partial results and progress on error, without performing any rollback.
//
//nolint:cyclop
func (c *Connector) executeSubscribe(
	ctx context.Context,
	params common.SubscribeParams,
	registrationParams *ResultData,
	req *SubscriptionRequest,
) (*SubscribeResult, *subscribeProgress, error) {
	sfRes := &SubscribeResult{
		EventChannelMembers: make(map[common.ObjectName]*EventChannelMember),
	}

	if req != nil {
		sfRes.ManualCheckboxCreation = req.ManualCheckboxCreation
		sfRes.ManualApexTriggerCreation = req.ManualApexTriggerCreation
	}

	progress := &subscribeProgress{
		req:            req,
		createdMembers: sfRes.EventChannelMembers,
	}

	if err := c.upsertQuotaOptimizationFields(ctx, params, req); err != nil {
		return sfRes, progress, fmt.Errorf("failed to upsert quota optimization fields: %w", err)
	}

	// Skip the upsert flag in manual mode: no Metadata API write occurred, so
	// rollback must not attempt to delete the (caller-managed) fields.
	if req == nil || !req.ManualCheckboxCreation {
		progress.quotaFieldsUpserted = true
	}

	// Clone instead of aliasing so subsequent mutations of
	// sfRes.QuotaOptimizationObjectFields (e.g. clear() in DeleteSubscription
	// or rollback paths) do not silently mutate the caller's req map.
	if req != nil && req.QuotaOptimizationObjectFields != nil {
		sfRes.QuotaOptimizationObjectFields = maps.Clone(req.QuotaOptimizationObjectFields)
	}

	// Deploy apex triggers before creating event channel members. Both reference
	// the quota optimization field, but neither references the other, so the order
	// is free. Doing triggers first means: (a) if the failure-prone trigger deploy
	// fails, rollback only has to drop the custom field — no ECMs to tear down;
	// (b) once ECMs go live, the trigger is already maintaining the indicator,
	// so no UPDATE events get silently filtered out during the setup window.
	//
	// In ManualApexTriggerCreation mode the deploy is skipped: we only verify the
	// caller-managed triggers exist. progress.deployedTriggers stays nil so the
	// rollback path leaves the caller's triggers alone.
	if req != nil && req.ManualApexTriggerCreation {
		triggers, err := c.verifyApexTriggersForCDC(ctx, params, req)
		sfRes.ApexTriggers = triggers

		if err != nil {
			return sfRes, progress, err
		}
	} else {
		deployOut, err := c.deployApexTriggersForCDC(ctx, params, req)

		progress.deployedTriggers = filterSuccessfulTriggers(deployOut)
		sfRes.ApexTriggers = toApexTriggers(deployOut)

		if err != nil {
			return sfRes, progress, err
		}
	}

	if err := c.createEventChannelMembers(ctx, params, registrationParams, req, sfRes); err != nil {
		return sfRes, progress, err
	}

	return sfRes, progress, nil
}

// createEventChannelMembers creates CDC event channel members for each subscribed object,
// setting filter expressions and enriched fields from quota optimization configuration.
func (c *Connector) createEventChannelMembers(
	ctx context.Context,
	params common.SubscribeParams,
	registrationParams *ResultData,
	req *SubscriptionRequest,
	sfRes *SubscribeResult,
) error {
	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(registrationParams.EventChannel.FullName)

		channelMetadata := &EventChannelMemberMetadata{
			EventChannel:   GetChannelName(rawChannelName),
			SelectedEntity: eventName,
		}

		if req != nil {
			channelMetadata.FilterExpression = buildQuotaFilterExpression(req.QuotaOptimizationObjectFields, objName)
			channelMetadata.EnrichedFields = buildQuotaEnrichedFields(req.QuotaOptimizationObjectFields, objName)
		}

		channelMember := &EventChannelMember{
			FullName: GetChangeDataCaptureChannelMembershipName(rawChannelName, eventName),
			Metadata: channelMetadata,
		}

		newChannelMember, err := c.CreateEventChannelMember(ctx, channelMember)
		if err != nil {
			return fmt.Errorf("failed to create event channel member for object %s, %w", objName, err)
		}

		sfRes.EventChannelMembers[objName] = newChannelMember
	}

	return nil
}

// DeleteSubscription tears down a Salesforce CDC subscription by removing the
// artifacts created by Subscribe / UpdateSubscription, in the inverse of creation
// order:
//
//  1. Delete event channel members. The first failure aborts and returns an error;
//     remaining triggers and fields are not touched.
//  2. Delete apex triggers (and their companion Test_<TriggerName> classes via
//     destructive deploy). Same fail-fast behavior as members.
//  3. Delete quota optimization custom fields. Best-effort: a failure here is
//     logged as a warning and the function returns nil. Common cause is that the
//     field is still referenced by other metadata (e.g. PlatformEventChannelMembers
//     on channels we don't manage).
//
// As each artifact is successfully removed, the corresponding entry is deleted
// from params.Result so the surviving entries faithfully describe what is still
// in Salesforce. Callers can inspect params.Result after the call (success or
// failure) to see what remains.
//
// The dependency-removal rule is reflected in the order: filter expressions and
// triggers (which reference the custom fields) are removed before the fields
// they reference, so field deletion is not blocked by stale references.
//
//nolint:cyclop,funlen
func (c *Connector) DeleteSubscription(ctx context.Context, params common.SubscriptionResult) error {
	if params.Result == nil {
		return fmt.Errorf("%w: missing SubscriptionResult.Result", errMissingParams)
	}

	sfRes, ok := params.Result.(*SubscribeResult)
	if !ok {
		return fmt.Errorf(
			"%w: expected SubscriptionResult.Result to be type '%T', but got '%T'",
			errInvalidRequestType,
			sfRes,
			params.Result,
		)
	}

	// Migrate old CheckboxField to IndicatorField for backwards compatibility.
	migrateApexTriggers(sfRes.ApexTriggers)

	// Tear down in reverse of creation: event channel members first to stop CDC
	// traffic, then apex triggers. Both reference the quota optimization fields
	// and must be removed before the custom fields can be deleted.
	//
	// Each successful per-object delete is removed from sfRes so that, if a later
	// step fails and the caller retries or inspects sfRes, the surviving entries
	// faithfully describe what is still in Salesforce.
	for objectName, member := range sfRes.EventChannelMembers {
		if member == nil {
			slog.Warn("event channel member entry is nil, skipping delete",
				"object", objectName,
			)

			continue
		}

		// TODO: check existence before delete
		if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
			slog.Warn(
				"event channel member delete failed mid-teardown; aborting before remaining "+
					"members, triggers, and quota fields are touched",
				"failedObject", objectName,
				"error", err,
				"remainingChannelMembers", datautils.FromMap(sfRes.EventChannelMembers).Keys(),
				"remainingApexTriggers", datautils.FromMap(sfRes.ApexTriggers).Keys(),
				"remainingQuotaFields", datautils.FromMap(sfRes.QuotaOptimizationObjectFields).Keys(),
			)

			return fmt.Errorf("failed to delete event channel member '%s': %w", objectName, err)
		}

		delete(sfRes.EventChannelMembers, objectName)
	}

	// In manual mode the caller manages trigger lifecycle: leave sfRes.ApexTriggers
	// populated (they still exist in Salesforce) and do not destructively deploy.
	if !sfRes.ManualApexTriggerCreation {
		for objName, trigger := range sfRes.ApexTriggers {
			if trigger == nil {
				slog.Warn("apex trigger entry is nil, skipping delete",
					"object", objName,
				)

				continue
			}

			// TODO: check existence before delete
			if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
				slog.Warn(
					"apex trigger delete failed mid-teardown; aborting before "+
						"remaining triggers and quota fields are touched",
					"failedObject", objName,
					"error", err,
					"remainingApexTriggers", datautils.FromMap(sfRes.ApexTriggers).Keys(),
					"remainingQuotaFields", datautils.FromMap(sfRes.QuotaOptimizationObjectFields).Keys(),
				)

				return fmt.Errorf("failed to delete apex trigger for object '%s': %w", objName, err)
			}

			delete(sfRes.ApexTriggers, objName)
		}
	}

	// In manual mode the caller manages checkbox field lifecycle: leave the
	// QuotaOptimizationObjectFields map populated and do not call DeleteMetadata.
	if !sfRes.ManualCheckboxCreation && len(sfRes.QuotaOptimizationObjectFields) > 0 {
		deleteFields := make(map[common.ObjectName][]string)

		for objectName, fieldName := range sfRes.QuotaOptimizationObjectFields {
			deleteFields[objectName] = append(
				deleteFields[objectName], customFieldAPIName(fieldName),
			)
		}

		// Best-effort: a quota-field delete can fail when the field is still referenced by
		// other metadata (e.g. PlatformEventChannelMember filter expressions on channels we
		// don't manage). Log and continue so the rest of the subscription teardown succeeds.
		// TODO: check existence before delete
		if _, err := c.DeleteMetadata(ctx, &common.DeleteMetadataParams{
			Fields: deleteFields,
		}); err != nil {
			slog.Warn("failed to delete quota optimization fields, continuing",
				"error", err,
				"fields", deleteFields,
			)
		}
		// Clear regardless of error: residual checkbox fields left in Salesforce
		// are inert once the trigger and channel member referencing them are gone,
		// and sfRes must not advertise fields this teardown attempted to remove.
		clear(sfRes.QuotaOptimizationObjectFields)
	}

	return nil
}

// updateSubscriptionProgress tracks which reversible operations completed during
// executeUpdateSubscription, so that rollbackUpdateSubscription knows what to undo.
type updateSubscriptionProgress struct {
	quotaFieldsUpserted bool
	newQuotaFields      map[common.ObjectName]string
}

// UpdateSubscription reconciles an existing subscription against the new request,
// performing four operations in order:
//
//  1. Upsert any quota optimization custom fields that the new request configures.
//     Idempotent for fields that already exist.
//  2. Update existing objects (those subscribed in both prevState and the new request):
//     redeploy apex triggers via Metadata API upsert, and PATCH PlatformEventChannelMember
//     filter expressions / enriched fields. Both operations are atomic from
//     Salesforce's side, so per-object failure leaves the prior state intact.
//  3. Delete objects that were in prevState but are not in the new request, via
//     DeleteSubscription. Quota fields whose existing-object references were just
//     cleared in step 2 are also cleaned up here.
//  4. Subscribe newly-added objects (in the new request but not in prevState),
//     via an inner Subscribe call with a narrowed SubscribeParams.
//
// Step 2 runs before step 3 so that filter / enriched references to quota fields
// are cleared before DeleteSubscription tries to remove those fields, satisfying
// the dependency-removal rule (dependents removed before dependency).
//
// On success, returns SubscriptionResult with Status = Success and Result reflecting
// the merged state (kept + new − removed). On failure, returns a partial result
// representing what is currently in Salesforce, with Status = Failed (rollback
// succeeded) or Status = FailedToRollback (rollback also errored). The rollback
// only undoes truly-new quota fields; existing-object updates and removals are not
// undone, but their per-object atomicity means each artifact is in a consistent
// state — a retry of the same UpdateSubscription is idempotent and converges.
//
//nolint:cyclop,nestif,funlen
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Validate params up-front (mirrors Subscribe) so a malformed input is
	// rejected before any Salesforce-side mutation happens. Without this, a
	// missing or invalid params would only be caught later by the inner
	// Subscribe call — after upsertQuotaOptimizationFields, updateExistingSubscriptions,
	// and DeleteSubscription have already run.
	if params.RegistrationResult == nil {
		return nil, fmt.Errorf("%w: missing RegistrationResult", errMissingParams)
	}

	if params.RegistrationResult.Result == nil {
		return nil, fmt.Errorf("%w: missing RegistrationResult.Result", errMissingParams)
	}

	if err := validator.New().Struct(params); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// validate the previous result
	if previousResult.Result == nil {
		return nil, fmt.Errorf("%w: missing previousResult.Result", errMissingParams)
	}

	prevState, ok := previousResult.Result.(*SubscribeResult)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected previousResult.Result to be type '%T', but got '%T'",
			errInvalidRequestType,
			prevState,
			previousResult.Result,
		)
	}

	// Migrate old CheckboxField to IndicatorField for backwards compatibility.
	migrateApexTriggers(prevState.ApexTriggers)

	var req *SubscriptionRequest

	if params.Request != nil {
		var ok bool

		req, ok = params.Request.(*SubscriptionRequest)
		if !ok {
			return nil, fmt.Errorf(
				"%w: expected SubscribeParams.Request to be type '%T', but got '%T'", errInvalidRequestType,
				req, params.Request,
			)
		}
	}

	result, progress, execErr := c.executeUpdateSubscription(ctx, params, previousResult, prevState, req)
	if execErr != nil {
		rollbackErr := c.rollbackUpdateSubscription(ctx, progress)

		// Reflect post-rollback state in the partial result:
		//   - Prune the truly-new quota fields that rollback attempted to remove.
		//     Per project policy, residuals on the Salesforce side are tolerable
		//     and the result must not advertise fields we tried to remove.
		//   - Escalate the status to FailedToRollback when rollback errored, so
		//     the caller can distinguish a clean teardown from a partial one.
		if result != nil {
			if state, ok := result.Result.(*SubscribeResult); ok && progress.quotaFieldsUpserted {
				for objName := range progress.newQuotaFields {
					delete(state.QuotaOptimizationObjectFields, objName)
				}
			}

			if rollbackErr != nil {
				result.Status = common.SubscriptionStatusFailedToRollback
			}
		}

		return result, errors.Join(execErr, rollbackErr)
	}

	return result, nil
}

// executeUpdateSubscription performs the forward-path logic of UpdateSubscription.
// It returns a SubscriptionResult that reflects current Salesforce state at the
// point of return (success or failure) along with partial progress for the
// rollback path. The caller (UpdateSubscription) is responsible for invoking
// rollbackUpdateSubscription on error and adjusting the returned result's
// status accordingly.
//
//nolint:cyclop,funlen
func (c *Connector) executeUpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
	prevState *SubscribeResult,
	req *SubscriptionRequest,
) (*common.SubscriptionResult, *updateSubscriptionProgress, error) {
	progress := &updateSubscriptionProgress{}

	if err := c.upsertQuotaOptimizationFields(ctx, params, req); err != nil {
		// No mutation has occurred yet; prevState still describes Salesforce.
		return buildPartialUpdateResult(prevState, nil, nil, req, common.SubscriptionStatusFailed),
			progress, err
	}

	// Skip the upsert flag in manual mode: no Metadata API write occurred, so
	// rollback must not attempt to delete the (caller-managed) fields.
	if req == nil || !req.ManualCheckboxCreation {
		progress.quotaFieldsUpserted = true
	}

	// Identify truly new quota fields and filter prevState so DeleteSubscription
	// only removes fields for objects being removed.
	progress.newQuotaFields = prepareQuotaOptimizationObjectFieldsForUpdate(req, prevState)

	diff := computeSubscriptionDiff(params, prevState)

	// Try updating existing subscriptions with filter expressions and enriched fields.
	// If this fails, we simply return so minimal state change is done.
	if err := c.updateExistingSubscriptions(ctx, req, diff); err != nil {
		return buildPartialUpdateResult(prevState, &diff, nil, req, common.SubscriptionStatusFailed),
			progress, err
	}

	deleteParams := *previousResult
	deleteParams.Result = prevState
	deleteParams.Objects = diff.objectsToDelete

	// Delete objects selected for removal. Their ECMs and triggers come from
	// prevState; prevState.QuotaOptimizationObjectFields includes (a) fields
	// owned by objects-to-delete and (b) quota fields whose existing-object
	// references were just cleared above by updateExistingSubscriptions, so the
	// field deletion can now succeed for both classes.
	if err := c.DeleteSubscription(ctx, deleteParams); err != nil {
		return buildPartialUpdateResult(prevState, &diff, nil, req, common.SubscriptionStatusFailed),
			progress, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	// Build a narrowed SubscribeParams for the inner Subscribe so it operates
	// only on objects-to-add. We use a fresh SubscriptionEvents map (built from
	// diff.objectsToAdd against the upfront snapshot in diff.allObjectEvents)
	// instead of relying on a side-effecting mutation of the caller's
	// params.SubscriptionEvents.
	//
	// The Request is also a fresh SubscriptionRequest carrying only the new
	// objects' quota field configuration, so upsertQuotaOptimizationFields
	// inside Subscribe can prune without disturbing the outer req's existing-object
	// entries. Salesforce's UpsertMetadata is idempotent, so re-upserting the
	// new-object fields (already done at the outer level) is a wasted call but
	// not a correctness issue.
	innerParams := params

	innerEvents := make(map[common.ObjectName]common.ObjectEvents, len(diff.objectsToAdd))
	for _, objName := range diff.objectsToAdd {
		if events, ok := diff.allObjectEvents[objName]; ok {
			innerEvents[objName] = events
		}
	}

	innerParams.SubscriptionEvents = innerEvents

	if req != nil {
		innerFields := make(map[common.ObjectName]string)

		for _, objName := range diff.objectsToAdd {
			if fieldName, ok := lookupQuotaField(req.QuotaOptimizationObjectFields, objName); ok {
				innerFields[objName] = fieldName
			}
		}

		innerParams.Request = &SubscriptionRequest{
			QuotaOptimizationObjectFields: innerFields,
			ManualCheckboxCreation:        req.ManualCheckboxCreation,
			ManualApexTriggerCreation:     req.ManualApexTriggerCreation,
		}
	}

	createRes, subscribeErr := c.Subscribe(ctx, innerParams)
	if subscribeErr != nil {
		// Subscribe always returns a non-nil result even on failure, with status
		// indicating whether its rollback succeeded; propagate that escalation.
		status := common.SubscriptionStatusFailed
		if createRes != nil && createRes.Status == common.SubscriptionStatusFailedToRollback {
			status = common.SubscriptionStatusFailedToRollback
		}

		return buildPartialUpdateResult(prevState, &diff, createRes, req, status),
			progress, fmt.Errorf("failed to subscribe to new objects: %w", subscribeErr)
	}

	newState := buildUpdatedSubscribeResult(prevState, createRes, diff, req)

	return &common.SubscriptionResult{
		Status: common.SubscriptionStatusSuccess,
		Result: newState,
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
		Objects: datautils.FromMap(newState.EventChannelMembers).Keys(),
	}, progress, nil
}

// buildPartialUpdateResult assembles a SubscriptionResult that describes the
// current Salesforce state at the point an error was hit during UpdateSubscription.
// diff is nil when computeSubscriptionDiff has not yet run; createRes is nil when
// the inner Subscribe was not reached.
func buildPartialUpdateResult(
	prevState *SubscribeResult,
	diff *subscriptionDiff,
	createRes *common.SubscriptionResult,
	req *SubscriptionRequest,
	status common.SubscriptionStatus,
) *common.SubscriptionResult {
	var state *SubscribeResult

	if diff == nil {
		// No work has begun; prevState already reflects Salesforce.
		state = prevState
	} else {
		state = buildUpdatedSubscribeResult(prevState, createRes, *diff, req)
	}

	return &common.SubscriptionResult{
		Status: status,
		Result: state,
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
		Objects: datautils.FromMap(state.EventChannelMembers).Keys(),
	}
}

// subscriptionDiff holds the result of diffing current subscription events against previous state.
type subscriptionDiff struct {
	channelMembersExisting map[common.ObjectName]*EventChannelMember
	apexTriggersExisting   map[common.ObjectName]*ApexTrigger
	allObjectEvents        map[common.ObjectName]common.ObjectEvents
	objectsToDelete        []common.ObjectName
	// objectsToAdd lists objects present in the new params but not in prevState.
	// Tracked explicitly so callers can iterate the new-object set without
	// relying on the side-effecting mutation of params.SubscriptionEvents that
	// computeSubscriptionDiff also performs.
	objectsToAdd []common.ObjectName
}

// computeSubscriptionDiff determines which objects to add, keep, and delete.
// It mutates prevState.EventChannelMembers (removes objects being kept) as a
// side effect; this is load-bearing for the subsequent DeleteSubscription call,
// which scopes its work via prevState.EventChannelMembers.
//
// params.SubscriptionEvents is NOT mutated. Callers needing the new-object
// subset should iterate diff.objectsToAdd.
func computeSubscriptionDiff(
	params common.SubscribeParams,
	prevState *SubscribeResult,
) subscriptionDiff {
	// Snapshot all subscription events for downstream consumers (e.g. building
	// a new-objects-only map for the inner Subscribe call).
	allObjectEvents := make(map[common.ObjectName]common.ObjectEvents, len(params.SubscriptionEvents))
	maps.Copy(allObjectEvents, params.SubscriptionEvents)

	objectsExisting := []common.ObjectName{}

	for objName := range prevState.EventChannelMembers {
		if _, ok := params.SubscriptionEvents[objName]; ok {
			objectsExisting = append(objectsExisting, objName)
		}
	}

	objectsToAdd := make([]common.ObjectName, 0, len(params.SubscriptionEvents))
	for objName := range params.SubscriptionEvents {
		if _, ok := prevState.EventChannelMembers[objName]; !ok {
			objectsToAdd = append(objectsToAdd, objName)
		}
	}

	channelMembersExisting := make(map[common.ObjectName]*EventChannelMember)
	apexTriggersExisting := make(map[common.ObjectName]*ApexTrigger)

	for _, objName := range objectsExisting {
		channelMembersExisting[objName] = prevState.EventChannelMembers[objName]
		delete(prevState.EventChannelMembers, objName)

		if trigger, ok := prevState.ApexTriggers[objName]; ok {
			apexTriggersExisting[objName] = trigger
			delete(prevState.ApexTriggers, objName)
		}
	}

	objectsToDelete := make([]common.ObjectName, 0, len(prevState.EventChannelMembers))
	for objName := range prevState.EventChannelMembers {
		objectsToDelete = append(objectsToDelete, objName)
	}

	return subscriptionDiff{
		channelMembersExisting: channelMembersExisting,
		apexTriggersExisting:   apexTriggersExisting,
		allObjectEvents:        allObjectEvents,
		objectsToDelete:        objectsToDelete,
		objectsToAdd:           objectsToAdd,
	}
}

// buildUpdatedSubscribeResult merges three sources into a single SubscribeResult
// reflecting current Salesforce state:
//
//   - prevState's residual ECMs/Triggers — entries that DeleteSubscription either
//     hasn't deleted yet (early-failure paths) or failed to delete (partial-failure).
//     The corruption-A fix in DeleteSubscription removes successfully-deleted
//     entries from prevState as it goes, so what remains here is exactly what is
//     still in Salesforce among the objects-to-delete set.
//   - diff.channelMembersExisting / diff.apexTriggersExisting — existing objects in
//     their post-update state (or pre-update if updateExistingSubscriptions failed
//     before reaching them).
//   - createRes from the inner Subscribe (when non-nil) — newly added objects,
//     reflecting post-rollback state on Subscribe failure.
//
// createRes may be nil when called on a partial-failure path before the inner
// Subscribe ran (or when the inner Subscribe returned a result without a typed
// *SubscribeResult).
func buildUpdatedSubscribeResult(
	prevState *SubscribeResult,
	createRes *common.SubscriptionResult,
	diff subscriptionDiff,
	req *SubscriptionRequest,
) *SubscribeResult {
	var newCreateRes *SubscribeResult

	if createRes != nil {
		if val, ok := createRes.Result.(*SubscribeResult); ok {
			newCreateRes = val
		}
	}

	// Capture prevState's current ECMs/Triggers before the field reassignments
	// below replace prevState's pointers. After corruption-A handling in
	// DeleteSubscription, these contain only objects-to-delete entries that are
	// still in Salesforce (either DeleteSubscription hasn't run yet or failed
	// for those entries). Including them ensures the partial result accurately
	// reflects "what's still in Salesforce" rather than over-claiming deletion.
	residualECMs := prevState.EventChannelMembers
	residualTriggers := prevState.ApexTriggers

	newState := prevState

	newState.EventChannelMembers = make(map[common.ObjectName]*EventChannelMember)
	maps.Copy(newState.EventChannelMembers, residualECMs)
	maps.Copy(newState.EventChannelMembers, diff.channelMembersExisting)

	if newCreateRes != nil {
		maps.Copy(newState.EventChannelMembers, newCreateRes.EventChannelMembers)
	}

	newState.ApexTriggers = make(map[common.ObjectName]*ApexTrigger)
	maps.Copy(newState.ApexTriggers, residualTriggers)
	maps.Copy(newState.ApexTriggers, diff.apexTriggersExisting)

	if newCreateRes != nil && len(newCreateRes.ApexTriggers) > 0 {
		maps.Copy(newState.ApexTriggers, newCreateRes.ApexTriggers)
	}

	// Clone so post-processing (e.g. pruning rolled-back fields in the
	// UpdateSubscription error path) doesn't mutate the caller's req map.
	if req != nil && req.QuotaOptimizationObjectFields != nil {
		newState.QuotaOptimizationObjectFields = maps.Clone(req.QuotaOptimizationObjectFields)
	}

	if req != nil {
		newState.ManualCheckboxCreation = req.ManualCheckboxCreation
		newState.ManualApexTriggerCreation = req.ManualApexTriggerCreation
	}

	return newState
}

// lookupQuotaField finds the checkbox field name for an object using case and plurality
// insensitive matching. Returns the field name and true if found.
func lookupQuotaField(
	quotaFields map[common.ObjectName]string, objName common.ObjectName,
) (string, bool) {
	for key, value := range quotaFields {
		if naming.PluralityAndCaseIgnoreEqual(string(key), string(objName)) {
			return value, true
		}
	}

	return "", false
}

// buildQuotaFilterExpression builds a CDC filter expression for an object based on its
// quota optimization checkbox field. The expression allows all non-UPDATE events through
// and only passes UPDATE events where the checkbox field is true (set by the Apex trigger
// when watched fields change).
// Returns empty string if the object has no quota optimization field configured.
func buildQuotaFilterExpression(
	quotaFields map[common.ObjectName]string, objName common.ObjectName,
) string {
	checkboxField, ok := lookupQuotaField(quotaFields, objName)
	if !ok {
		return ""
	}

	return fmt.Sprintf("ChangeEventHeader.changeType != 'UPDATE' OR %s = true",
		customFieldAPIName(checkboxField))
}

// buildQuotaEnrichedFields returns the enriched fields needed for the quota optimization
// filter expression. The checkbox field must be included as an enriched field so that
// CDC events contain its value for filter evaluation.
// Returns nil if the object has no quota optimization field configured.
func buildQuotaEnrichedFields(
	quotaFields map[common.ObjectName]string, objName common.ObjectName,
) []*EnrichedField {
	checkboxField, ok := lookupQuotaField(quotaFields, objName)
	if !ok {
		return nil
	}

	return []*EnrichedField{
		{Name: customFieldAPIName(checkboxField)},
	}
}

func customFieldAPIName(fieldName string) string {
	if strings.HasSuffix(fieldName, "__c") {
		return fieldName
	}

	return fieldName + "__c"
}

func customFieldDisplayName(fieldName string) string {
	return strings.TrimSuffix(fieldName, "__c")
}

// prepareQuotaOptimizationObjectFieldsForUpdate identifies truly new quota fields (in req but not
// in prevState) and filters prevState.QuotaOptimizationObjectFields so DeleteSubscription only
// removes fields for objects being removed. Returns the new-only quota fields for rollback use.
func prepareQuotaOptimizationObjectFieldsForUpdate(
	req *SubscriptionRequest, prevState *SubscribeResult,
) map[common.ObjectName]string {
	var newQuotaFields map[common.ObjectName]string

	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return newQuotaFields
	}

	newQuotaFields = make(map[common.ObjectName]string)

	for objectName, fieldName := range req.QuotaOptimizationObjectFields {
		if _, existed := prevState.QuotaOptimizationObjectFields[objectName]; !existed {
			newQuotaFields[objectName] = fieldName
		}
	}

	for objectName := range req.QuotaOptimizationObjectFields {
		delete(prevState.QuotaOptimizationObjectFields, objectName)
	}

	return newQuotaFields
}

// upsertQuotaOptimizationFields creates the quota-optimization checkbox custom field
// for each object that the caller has both (a) subscribed to via params.SubscriptionEvents
// and (b) configured a quota field for via req.QuotaOptimizationObjectFields.
//
// As an intentional side effect, drops entries from req.QuotaOptimizationObjectFields
// in place when the corresponding object is not in params.SubscriptionEvents. Those
// entries are phantoms — their fields would never be created in Salesforce — and the
// in-place filtering ensures they don't propagate to:
//   - sfRes.QuotaOptimizationObjectFields (cloned post-mutation in executeSubscribe,
//     so it captures only the actually-upserted set);
//   - progress.req (still the same pointer, so rollback iterates the post-mutation
//     view and only attempts to delete what was actually created);
//   - prepareQuotaOptimizationObjectFieldsForUpdate (sees the post-mutation view
//     when computing newQuotaFields for UpdateSubscription).
//
// The mutation is expected and produces no inconsistency: the returned result
// advertises exactly the fields that exist in Salesforce, rollback targets exactly
// those fields, and DeleteSubscription later cleans up exactly those fields. The
// only observable effect on callers is that their req.QuotaOptimizationObjectFields
// map is silently shrunk; callers that reuse the same req pointer across multiple
// calls should construct a fresh req per call to avoid observing the mutation.
//
// When req.ManualCheckboxCreation is true the function takes a verify-only path:
// it skips the UpsertMetadata call and instead checks each (now post-pruning)
// field exists in Salesforce, returning errManualCheckboxFieldMissing for any
// that don't. The phantom-pruning still runs so downstream consumers see the
// same post-mutation view.
func (c *Connector) upsertQuotaOptimizationFields(
	ctx context.Context, params common.SubscribeParams, req *SubscriptionRequest,
) error {
	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return nil
	}

	fields := make(map[string][]common.FieldDefinition)

	for objectName, fieldName := range req.QuotaOptimizationObjectFields {
		if !isObjectSubscribed(params.SubscriptionEvents, objectName) {
			delete(req.QuotaOptimizationObjectFields, objectName)

			continue
		}

		fields[string(objectName)] = []common.FieldDefinition{
			{
				FieldName:   customFieldAPIName(fieldName),
				DisplayName: customFieldDisplayName(fieldName),
				ValueType:   common.FieldTypeBoolean,
				Description: "THIS IS AUTOMATED FIELD. DO NOT EDIT THIS FIELD. " + //nolint:lll
					"This field is used to track if the quota optimization is used for the object",
				StringOptions: &common.StringFieldOptions{
					DefaultValue: new("false"),
				},
			},
		}
	}

	if len(fields) == 0 {
		return nil
	}

	if req.ManualCheckboxCreation {
		return c.verifyQuotaOptimizationFieldsExist(ctx, req)
	}

	if _, err := c.UpsertMetadata(ctx, &common.UpsertMetadataParams{
		Fields: fields,
	}); err != nil {
		return fmt.Errorf("failed to upsert quota optimization fields: %w", err)
	}

	return nil
}

// verifyQuotaOptimizationFieldsExist checks that every quota-optimization field
// declared in req.QuotaOptimizationObjectFields already exists in Salesforce.
// Used by the ManualCheckboxCreation path of upsertQuotaOptimizationFields where
// the caller is responsible for creating the fields outside this connector.
//
// Returns errManualCheckboxFieldMissing wrapping the list of missing fields if
// any are absent so the caller learns about every missing field in one round
// rather than fixing them one at a time.
func (c *Connector) verifyQuotaOptimizationFieldsExist(
	ctx context.Context, req *SubscriptionRequest,
) error {
	missing := make([]string, 0)

	for objectName, fieldName := range req.QuotaOptimizationObjectFields {
		apiName := customFieldAPIName(fieldName)

		exists, err := c.CustomCheckboxFieldExists(ctx, string(objectName), apiName)
		if err != nil {
			return fmt.Errorf("failed to check checkbox field existence for %s.%s: %w",
				objectName, apiName, err)
		}

		if !exists {
			missing = append(missing, fmt.Sprintf("%s.%s", objectName, apiName))
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", errManualCheckboxFieldMissing, strings.Join(missing, ", "))
	}

	return nil
}

// isObjectSubscribed reports whether objName matches any key in events using the
// same case-and-plurality-insensitive comparison as lookupQuotaField.
func isObjectSubscribed(
	events map[common.ObjectName]common.ObjectEvents, objName common.ObjectName,
) bool {
	for key := range events {
		if naming.PluralityAndCaseIgnoreEqual(string(key), string(objName)) {
			return true
		}
	}

	return false
}

// updateExistingSubscriptions reconciles the per-object Salesforce configuration
// for objects that remain subscribed across an UpdateSubscription call (i.e. the
// overlap of prevState and the new request — neither newly added nor being
// removed). Even though the subscription itself isn't being added or removed,
// the new request can change how that subscription is configured, and those
// changes need to be pushed to Salesforce:
//
//   - WatchFields may differ from the previous request, which changes which
//     record-field updates fire the apex trigger. The trigger code is generated
//     from WatchFields, so a change here requires redeploying the trigger.
//   - QuotaOptimizationObjectFields[obj] may be added, removed, or renamed.
//     Adding requires deploying a trigger and setting the channel member's
//     filter expression / enriched fields. Removing requires destructively
//     deleting the trigger (handled by the orphan branch in
//     redeployExistingApexTriggers) and clearing the filter / enriched fields.
//     Renaming requires both: point trigger and filter at the new field name.
//
// We don't diff old vs new configuration before acting — both Metadata API
// deploy (used for triggers) and Tooling API PATCH (used for channel members)
// are idempotent on Salesforce, so unconditional reconciliation converges to
// the desired state at the cost of one round trip per existing object when
// nothing has changed. Diffing per-object config would be more complex than
// that round trip is expensive.
//
// Order: triggers first, then channel members. Both reference the quota
// custom field; the order between them is free, but doing triggers first
// keeps the indicator field maintained while the new channel-member filter
// goes live.
func (c *Connector) updateExistingSubscriptions(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	if req == nil {
		return nil
	}

	if err := c.redeployExistingApexTriggers(ctx, req, diff); err != nil {
		return err
	}

	return c.updateExistingChannelMembers(ctx, req, diff)
}

// updateExistingChannelMembers updates filter expressions and enriched fields on
// existing PlatformEventChannelMember records via Tooling API PATCH. selectedEntity
// is unchanged for existing objects (still <ObjectName>ChangeEvent), so PATCH suffices —
// no delete-and-recreate window where CDC traffic is interrupted, and the member's
// Id is preserved. The PATCH is also atomic from Salesforce's side, so a failure
// leaves the prior FilterExpression / EnrichedFields intact.
//
// When the new request has no quota config for an object, buildQuotaFilterExpression
// and buildQuotaEnrichedFields return empty values; the PATCH then explicitly
// clears the existing filter / enriched fields on Salesforce. This frees the
// quota custom field's reference so the subsequent DeleteSubscription can
// successfully remove it (the dependency-removal ordering rule: filter and
// enriched fields are removed before the field they reference).
func (c *Connector) updateExistingChannelMembers(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	for objName, member := range diff.channelMembersExisting {
		if member == nil || member.Metadata == nil {
			slog.Warn("kept channel member entry is malformed, skipping update",
				"object", objName,
				"memberNil", member == nil,
				"metadataNil", member != nil && member.Metadata == nil,
			)

			continue
		}

		// Build the PATCH body in a fresh struct so the existing diff entry is
		// not mutated until Salesforce has acknowledged the update. On PATCH
		// failure the prior state is preserved both in Salesforce (PATCH is
		// atomic) and in diff.channelMembersExisting — the two stay in sync.
		updated := &EventChannelMember{
			Id:       member.Id,
			FullName: member.FullName,
			Metadata: &EventChannelMemberMetadata{
				EventChannel:     member.Metadata.EventChannel,
				SelectedEntity:   member.Metadata.SelectedEntity,
				FilterExpression: buildQuotaFilterExpression(req.QuotaOptimizationObjectFields, objName),
				EnrichedFields:   buildQuotaEnrichedFields(req.QuotaOptimizationObjectFields, objName),
			},
		}

		if _, err := c.UpdateEventChannelMember(ctx, updated); err != nil {
			return fmt.Errorf("failed to update event channel member for object %s: %w", objName, err)
		}

		diff.channelMembersExisting[objName] = updated
	}

	return nil
}
