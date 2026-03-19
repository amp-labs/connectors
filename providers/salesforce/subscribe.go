package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/goutils"
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
}

type ApexTrigger struct {
	ObjectName    common.ObjectName
	TriggerName   string
	CheckboxField string
	WatchFields   []string
	Success       bool
	Warnings      []string
	Errors        []string
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscribeResult{},
	}
}

type Filter struct {
	EnrichedFields   []*EnrichedField
	FilterExpression string
}

type SubscriptionRequest struct {
	Filters map[common.ObjectName]*Filter
	// QuotaOptimizationObjectFields maps object names to custom checkbox field names to create
	// on the object in Salesforce. See SubscribeResult.QuotaOptimizationObjectFields for details.
	QuotaOptimizationObjectFields map[common.ObjectName]string
}

// subscribeProgress tracks which reversible operations completed during executeSubscribe,
// so that rollbackSubscribe knows what to undo.
type subscribeProgress struct {
	quotaFieldsUpserted bool
	req                 *SubscriptionRequest
	createdMembers      map[common.ObjectName]*EventChannelMember
}

// Subscribe subscribes to the events for the given objects.
// It creates event channel members for each object in the subscription.
// If any of the event channel members fail to be created, it will rollback the operation.
// If the rollback fails, it will return the partial result along with the error.
// If the rollback is successful, it will return the original error on object.
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
		Objects: objectNames(sfRes.EventChannelMembers),
	}, nil
}

// executeSubscribe performs the forward-path logic of Subscribe.
// It returns partial results and progress on error, without performing any rollback.
func (c *Connector) executeSubscribe(
	ctx context.Context,
	params common.SubscribeParams,
	registrationParams *ResultData,
	req *SubscriptionRequest,
) (*SubscribeResult, *subscribeProgress, error) {
	sfRes := &SubscribeResult{
		EventChannelMembers: make(map[common.ObjectName]*EventChannelMember),
	}

	progress := &subscribeProgress{
		req:            req,
		createdMembers: sfRes.EventChannelMembers,
	}

	if err := c.upsertQuotaOptimizationFields(ctx, req); err != nil {
		return sfRes, progress, fmt.Errorf("failed to upsert quota optimization fields: %w", err)
	}

	progress.quotaFieldsUpserted = true

	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(registrationParams.EventChannel.FullName)

		channelMetadata := &EventChannelMemberMetadata{
			EventChannel:   GetChannelName(rawChannelName),
			SelectedEntity: eventName,
		}

		if req != nil && req.Filters != nil {
			for objKey, filter := range req.Filters {
				if naming.PluralityAndCaseIgnoreEqual(string(objKey), string(objName)) {
					channelMetadata.EnrichedFields = filter.EnrichedFields
					channelMetadata.FilterExpression = filter.FilterExpression

					break
				}
			}
		}

		channelMember := &EventChannelMember{
			FullName: GetChangeDataCaptureChannelMembershipName(rawChannelName, eventName),
			Metadata: channelMetadata,
		}

		newChannelMember, err := c.CreateEventChannelMember(ctx, channelMember)
		if err != nil {
			return sfRes, progress, fmt.Errorf("failed to create event channel member for object %s, %w", objName, err)
		}

		sfRes.EventChannelMembers[objName] = newChannelMember
	}

	if req != nil && req.QuotaOptimizationObjectFields != nil {
		sfRes.QuotaOptimizationObjectFields = req.QuotaOptimizationObjectFields
	}

	return sfRes, progress, nil
}

// rollbackSubscribe reverses completed operations in reverse order based on progress.
// It removes successfully rolled-back members from the shared createdMembers map
// and returns a SubscriptionResult reflecting the post-rollback state.
func (c *Connector) rollbackSubscribe(
	ctx context.Context,
	sfRes *SubscribeResult,
	progress *subscribeProgress,
) (*common.SubscriptionResult, error) {
	var rollbackErr error

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

// DeleteSubscription deletes the subscription by deleting all the event channel members.
// If any of the event channel members fail to be deleted, it will return an error.
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

	for objectName, member := range sfRes.EventChannelMembers {
		if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
			return fmt.Errorf("failed to delete event channel member '%s': %w", objectName, err)
		}
	}

	if len(sfRes.QuotaOptimizationObjectFields) > 0 {
		deleteFields := make(map[common.ObjectName][]string)

		for objectName, fieldName := range sfRes.QuotaOptimizationObjectFields {
			deleteFields[objectName] = append(
				deleteFields[objectName], customFieldAPIName(fieldName),
			)
		}

		if _, err := c.DeleteMetadata(ctx, &common.DeleteMetadataParams{
			Fields: deleteFields,
		}); err != nil {
			return fmt.Errorf("failed to delete quota optimization fields: %w", err)
		}
	}

	return nil
}

// updateSubscriptionProgress tracks which reversible operations completed during
// executeUpdateSubscription, so that rollbackUpdateSubscription knows what to undo.
type updateSubscriptionProgress struct {
	quotaFieldsUpserted bool
	newQuotaFields      map[common.ObjectName]string
}

// UpdateSubscription will update the subscription by:
// 1. Removing objects from the previous subscription that are not in the new subscription.
// 2. Adding new objects to the subscription that are in the new subscription but not in the previous subscription.
// 3. Returning the updated subscription result.
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
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

		return nil, errors.Join(execErr, rollbackErr)
	}

	return result, nil
}

// executeUpdateSubscription performs the forward-path logic of UpdateSubscription.
// It returns partial progress on error, without performing any rollback.
//
//nolint:cyclop
func (c *Connector) executeUpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
	prevState *SubscribeResult,
	req *SubscriptionRequest,
) (*common.SubscriptionResult, *updateSubscriptionProgress, error) {
	progress := &updateSubscriptionProgress{}

	if err := c.upsertQuotaOptimizationFields(ctx, req); err != nil {
		return nil, progress, err
	}

	progress.quotaFieldsUpserted = true

	// Identify truly new quota fields and filter prevState so DeleteSubscription
	// only removes fields for objects being removed.
	progress.newQuotaFields = prepareQuotaOptimizationObjectFieldsForUpdate(req, prevState)

	diff := computeSubscriptionDiff(params, prevState)

	deleteParams := *previousResult
	deleteParams.Result = prevState
	deleteParams.Objects = diff.objectsToDelete

	// Delete only objects that were selected for removal, preserving objects
	// that need to remain in the subscription.
	if err := c.DeleteSubscription(ctx, deleteParams); err != nil {
		return nil, progress, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	// Update filters on kept channel members if the request includes new filters.
	if err := c.updateChannelMemberFilters(ctx, req, diff.channelMembersToKeep); err != nil {
		return nil, progress, err
	}

	// Temporarily clear QuotaOptimizationObjectFields from the request before calling Subscribe,
	// since we already upserted them above. This avoids a duplicate UpsertMetadata call.
	var savedQuotaOptimizationObjectFields map[common.ObjectName]string
	if req != nil {
		savedQuotaOptimizationObjectFields = req.QuotaOptimizationObjectFields
		req.QuotaOptimizationObjectFields = nil
	}

	// create new subscription
	createRes, err := c.Subscribe(ctx, params)
	if err != nil {
		return nil, progress, fmt.Errorf("failed to subscribe to new objects: %w", err)
	}

	// Restore QuotaOptimizationObjectFields so it can be saved in the new state.
	if req != nil {
		req.QuotaOptimizationObjectFields = savedQuotaOptimizationObjectFields
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
		Objects: objectNames(newState.EventChannelMembers),
	}, progress, nil
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

// subscriptionDiff holds the result of diffing current subscription events against previous state.
type subscriptionDiff struct {
	channelMembersToKeep map[common.ObjectName]*EventChannelMember
	objectsToDelete      []common.ObjectName
}

// computeSubscriptionDiff determines which objects to add, keep, and delete.
// It mutates params.SubscriptionEvents (removes already-subscribed objects) and
// prevState.EventChannelMembers (removes objects being kept) as side effects.
func computeSubscriptionDiff(
	params common.SubscribeParams,
	prevState *SubscribeResult,
) subscriptionDiff {
	objectsToExcludeFromSubscription := []common.ObjectName{}
	objectsExcludeFromDelete := []common.ObjectName{}

	for objName := range params.SubscriptionEvents {
		if _, ok := prevState.EventChannelMembers[objName]; ok {
			objectsToExcludeFromSubscription = append(objectsToExcludeFromSubscription, objName)
		}
	}

	for objName := range prevState.EventChannelMembers {
		if _, ok := params.SubscriptionEvents[objName]; ok {
			objectsExcludeFromDelete = append(objectsExcludeFromDelete, objName)
		}
	}

	for _, objName := range objectsToExcludeFromSubscription {
		delete(params.SubscriptionEvents, objName)
	}

	channelMembersToKeep := make(map[common.ObjectName]*EventChannelMember)

	for _, objName := range objectsExcludeFromDelete {
		channelMembersToKeep[objName] = prevState.EventChannelMembers[objName]
		delete(prevState.EventChannelMembers, objName)
	}

	objectsToDelete := make([]common.ObjectName, 0, len(prevState.EventChannelMembers))
	for objName := range prevState.EventChannelMembers {
		objectsToDelete = append(objectsToDelete, objName)
	}

	return subscriptionDiff{
		channelMembersToKeep: channelMembersToKeep,
		objectsToDelete:      objectsToDelete,
	}
}

// buildUpdatedSubscribeResult merges the kept members with newly created members
// and applies quota optimization fields from the request.
func buildUpdatedSubscribeResult(
	prevState *SubscribeResult,
	createRes *common.SubscriptionResult,
	diff subscriptionDiff,
	req *SubscriptionRequest,
) *SubscribeResult {
	newState := prevState
	newState.EventChannelMembers = diff.channelMembersToKeep

	//nolint:forcetypeassert
	maps.Copy(newState.EventChannelMembers, createRes.Result.(*SubscribeResult).EventChannelMembers)

	for _, objName := range diff.objectsToDelete {
		delete(newState.EventChannelMembers, objName)
	}

	if req != nil && req.QuotaOptimizationObjectFields != nil {
		newState.QuotaOptimizationObjectFields = req.QuotaOptimizationObjectFields
	}

	return newState
}

func objectNames(members map[common.ObjectName]*EventChannelMember) []common.ObjectName {
	names := make([]common.ObjectName, 0, len(members))
	for objName := range members {
		names = append(names, objName)
	}

	return names
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

func (c *Connector) upsertQuotaOptimizationFields(
	ctx context.Context, req *SubscriptionRequest,
) error {
	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return nil
	}

	fields := make(map[string][]common.FieldDefinition)

	for objectName, fieldName := range req.QuotaOptimizationObjectFields {
		fields[string(objectName)] = []common.FieldDefinition{
			{
				FieldName:   customFieldAPIName(fieldName),
				DisplayName: customFieldDisplayName(fieldName),
				ValueType:   common.FieldTypeBoolean,
				Description: "THIS IS AUTOMATED FIELD. DO NOT EDIT THIS FIELD. " + //nolint:lll
					"This field is used to track if the quota optimization is used for the object",
				StringOptions: &common.StringFieldOptions{
					DefaultValue: goutils.Pointer("false"),
				},
			},
		}
	}

	if _, err := c.UpsertMetadata(ctx, &common.UpsertMetadataParams{
		Fields: fields,
	}); err != nil {
		return fmt.Errorf("failed to upsert quota optimization fields: %w", err)
	}

	return nil
}

func (c *Connector) updateChannelMemberFilters(
	ctx context.Context, req *SubscriptionRequest, members map[common.ObjectName]*EventChannelMember,
) error {
	if req == nil || req.Filters == nil {
		return nil
	}

	for objName, member := range members {
		for objKey, filter := range req.Filters {
			if naming.PluralityAndCaseIgnoreEqual(string(objKey), string(objName)) {
				member.Metadata.EnrichedFields = filter.EnrichedFields
				member.Metadata.FilterExpression = filter.FilterExpression

				updatedMember, err := c.UpdateEventChannelMember(ctx, member)
				if err != nil {
					return fmt.Errorf("failed to update event channel member filters for object %s: %w", objName, err)
				}

				members[objName] = updatedMember

				break
			}
		}
	}

	return nil
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
