package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/go-playground/validator"
)

type SubscribeResult struct {
	EventChannelMembers map[common.ObjectName]*EventChannelMember
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscribeResult{},
	}
}

// subscribeProgress tracks which reversible operations completed during executeSubscribe,
// so that rollbackSubscribe knows what to undo.
type subscribeProgress struct {
	createdMembers map[common.ObjectName]*EventChannelMember
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

	sfRes, progress, execErr := c.executeSubscribe(ctx, params, registrationParams)
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
) (*SubscribeResult, *subscribeProgress, error) {
	sfRes := &SubscribeResult{
		EventChannelMembers: make(map[common.ObjectName]*EventChannelMember),
	}

	progress := &subscribeProgress{
		createdMembers: sfRes.EventChannelMembers,
	}

	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(registrationParams.EventChannel.FullName)

		channelMetadata := &EventChannelMemberMetadata{
			EventChannel:   GetChannelName(rawChannelName),
			SelectedEntity: eventName,
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

	return nil
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

	return c.executeUpdateSubscription(ctx, params, previousResult, prevState)
}

// executeUpdateSubscription performs the forward-path logic of UpdateSubscription.
func (c *Connector) executeUpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
	prevState *SubscribeResult,
) (*common.SubscriptionResult, error) {
	diff := computeSubscriptionDiff(params, prevState)

	deleteParams := *previousResult
	deleteParams.Result = prevState
	deleteParams.Objects = diff.objectsToDelete

	// Delete only objects that were selected for removal, preserving objects
	// that need to remain in the subscription.
	if err := c.DeleteSubscription(ctx, deleteParams); err != nil {
		return nil, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	// Update filters on kept channel members if the request includes new filters.
	if err := c.updateChannelMemberFilters(ctx, params, diff.channelMembersToKeep); err != nil {
		return nil, err
	}

	// create new subscription
	createRes, err := c.Subscribe(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new objects: %w", err)
	}

	newState := buildUpdatedSubscribeResult(prevState, createRes, diff)

	return &common.SubscriptionResult{
		Status: common.SubscriptionStatusSuccess,
		Result: newState,
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
		Objects: objectNames(newState.EventChannelMembers),
	}, nil
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
) *SubscribeResult {
	newState := prevState
	newState.EventChannelMembers = diff.channelMembersToKeep

	//nolint:forcetypeassert
	maps.Copy(newState.EventChannelMembers, createRes.Result.(*SubscribeResult).EventChannelMembers)

	for _, objName := range diff.objectsToDelete {
		delete(newState.EventChannelMembers, objName)
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

func (c *Connector) updateChannelMemberFilters(
	ctx context.Context,
	params common.SubscribeParams,
	members map[common.ObjectName]*EventChannelMember,
) error {
	for objName, member := range members {
		for eventObjName := range params.SubscriptionEvents {
			if naming.PluralityAndCaseIgnoreEqual(string(eventObjName), string(objName)) {
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
