package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
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
}

// subscribeProgress tracks which reversible operations completed during executeSubscribe,
// so that rollbackSubscribe knows what to undo.
type subscribeProgress struct {
	quotaFieldsUpserted bool
	req                 *SubscriptionRequest
	createdMembers      map[common.ObjectName]*EventChannelMember
	deployedTriggers    map[common.ObjectName]*ApexTriggerResult
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
		Objects: datautils.FromMap(sfRes.EventChannelMembers).Keys(),
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

	if err := c.createEventChannelMembers(ctx, params, registrationParams, req, sfRes); err != nil {
		return sfRes, progress, err
	}

	if req != nil && req.QuotaOptimizationObjectFields != nil {
		sfRes.QuotaOptimizationObjectFields = req.QuotaOptimizationObjectFields
	}

	deployOut, err := c.deployApexTriggersForCDC(ctx, params, req)

	progress.deployedTriggers = filterSuccessfulTriggers(deployOut)
	sfRes.ApexTriggers = toApexTriggers(deployOut)

	if err != nil {
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

	// Migrate old CheckboxField to IndicatorField for backwards compatibility.
	migrateApexTriggers(sfRes.ApexTriggers)

	// Delete apex triggers first — they reference the quota optimization fields,
	// so they must be removed before the custom fields can be deleted.
	for objName, trigger := range sfRes.ApexTriggers {
		if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
			return fmt.Errorf("failed to delete apex trigger for object '%s': %w", objName, err)
		}
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

	// Update channel members and redeploy apex triggers for kept objects.
	if err := c.updateKeptSubscriptions(ctx, req, diff); err != nil {
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
		Objects: datautils.FromMap(newState.EventChannelMembers).Keys(),
	}, progress, nil
}

// subscriptionDiff holds the result of diffing current subscription events against previous state.
type subscriptionDiff struct {
	channelMembersToKeep map[common.ObjectName]*EventChannelMember
	apexTriggersToKeep   map[common.ObjectName]*ApexTrigger
	keptObjectEvents     map[common.ObjectName]common.ObjectEvents
	objectsToDelete      []common.ObjectName
}

// computeSubscriptionDiff determines which objects to add, keep, and delete.
// It mutates params.SubscriptionEvents (removes already-subscribed objects) and
// prevState.EventChannelMembers (removes objects being kept) as side effects.
func computeSubscriptionDiff(
	params common.SubscribeParams,
	prevState *SubscribeResult,
) subscriptionDiff {
	// Save all subscription events upfront before mutation removes kept objects.
	allObjectEvents := make(map[common.ObjectName]common.ObjectEvents, len(params.SubscriptionEvents))
	maps.Copy(allObjectEvents, params.SubscriptionEvents)

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
	apexTriggersToKeep := make(map[common.ObjectName]*ApexTrigger)

	for _, objName := range objectsExcludeFromDelete {
		channelMembersToKeep[objName] = prevState.EventChannelMembers[objName]
		delete(prevState.EventChannelMembers, objName)

		if trigger, ok := prevState.ApexTriggers[objName]; ok {
			apexTriggersToKeep[objName] = trigger
			delete(prevState.ApexTriggers, objName)
		}
	}

	objectsToDelete := make([]common.ObjectName, 0, len(prevState.EventChannelMembers))
	for objName := range prevState.EventChannelMembers {
		objectsToDelete = append(objectsToDelete, objName)
	}

	return subscriptionDiff{
		channelMembersToKeep: channelMembersToKeep,
		apexTriggersToKeep:   apexTriggersToKeep,
		keptObjectEvents:     allObjectEvents,
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
	//nolint:forcetypeassert
	newCreateRes := createRes.Result.(*SubscribeResult)

	newState := prevState
	newState.EventChannelMembers = diff.channelMembersToKeep
	maps.Copy(newState.EventChannelMembers, newCreateRes.EventChannelMembers)

	newState.ApexTriggers = make(map[common.ObjectName]*ApexTrigger)
	maps.Copy(newState.ApexTriggers, diff.apexTriggersToKeep)

	if len(newCreateRes.ApexTriggers) > 0 {
		maps.Copy(newState.ApexTriggers, newCreateRes.ApexTriggers)
	}

	for _, objName := range diff.objectsToDelete {
		delete(newState.EventChannelMembers, objName)
		delete(newState.ApexTriggers, objName)
	}

	if req != nil && req.QuotaOptimizationObjectFields != nil {
		newState.QuotaOptimizationObjectFields = req.QuotaOptimizationObjectFields
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

// updateKeptSubscriptions updates channel members and redeploys apex triggers for
// objects that are kept across an UpdateSubscription call.
func (c *Connector) updateKeptSubscriptions(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	if req == nil {
		return nil
	}

	if err := c.recreateKeptChannelMembers(ctx, req, diff); err != nil {
		return err
	}

	return c.redeployKeptApexTriggers(ctx, req, diff)
}

// recreateKeptChannelMembers deletes and recreates channel members for kept objects
// with updated filter expressions and enriched fields. Salesforce doesn't support
// PATCH on selectedEntity, so delete+recreate is required.
func (c *Connector) recreateKeptChannelMembers(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	for objName, member := range diff.channelMembersToKeep {
		filterExpr := buildQuotaFilterExpression(req.QuotaOptimizationObjectFields, objName)
		if filterExpr == "" {
			continue
		}

		if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
			return fmt.Errorf("failed to delete event channel member for object %s: %w", objName, err)
		}

		member.Id = ""
		member.Metadata.FilterExpression = filterExpr
		member.Metadata.EnrichedFields = buildQuotaEnrichedFields(req.QuotaOptimizationObjectFields, objName)

		newMember, err := c.CreateEventChannelMember(ctx, member)
		if err != nil {
			return fmt.Errorf("failed to recreate event channel member for object %s: %w", objName, err)
		}

		diff.channelMembersToKeep[objName] = newMember
	}

	return nil
}
