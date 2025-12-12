package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/tools/debug"
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

type Filter struct {
	EnrichedFields   []*EnrichedField
	FilterExpression string
}

type SubscriptionRequest struct {
	Filters map[common.ObjectName]*Filter
}

// Subscribe subscribes to the events for the given objects.
// It creates event channel members for each object in the subscription.
// If any of the event channel members fail to be created, it will rollback the operation.
// If the rollback fails, it will return the partial result along with the error.
// If the rollback is successful, it will return the original error on object.
// Registration is required prior to subscribing.
//
//nolint:funlen,cyclop,varnamelen
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

	registrationParams, ok := params.RegistrationResult.Result.(*ResultData)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected SubscribeParams.RegistrationResult.Result to be type '%T', but got '%T'", errInvalidRequestType,
			registrationParams,
			params.RegistrationResult.Result,
		)
	}

	var req *SubscriptionRequest

	if params.Request != nil {
		//nolint:varnamelen
		req, ok = params.Request.(*SubscriptionRequest)
		if !ok {
			return nil, fmt.Errorf(
				"%w: expected SubscribeParams.Request to be type '%T', but got '%T'", errInvalidRequestType,
				req, params.Request,
			)
		}
	}

	sfRes := &SubscribeResult{
		EventChannelMembers: make(map[common.ObjectName]*EventChannelMember),
	}

	var failError error

	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(registrationParams.EventChannel.FullName)

		channelMetadata := &EventChannelMemberMetadata{
			EventChannel:   GetChannelName(rawChannelName),
			SelectedEntity: eventName,
		}

		fmt.Println("channelMetadata before create=========", debug.PrettyFormatStringJSON(req))

		channelMember := &EventChannelMember{
			FullName: GetChangeDataCaptureChannelMembershipName(rawChannelName, eventName),
			Metadata: channelMetadata,
		}

		if req != nil && req.Filters != nil {
			for objKey, filter := range req.Filters {
				objName := string(objKey)
				if naming.PluralityAndCaseIgnoreEqual(objName, string(objName)) {
					channelMetadata.EnrichedFields = filter.EnrichedFields
					channelMetadata.FilterExpression = filter.FilterExpression

					break
				}
			}
		}

		fmt.Println("channelMember before create=========", debug.PrettyFormatStringJSON(channelMember))

		newChannelMember, err := c.CreateEventChannelMember(ctx, channelMember)
		if err != nil {
			failError = fmt.Errorf("failed to create event channel member for object %s, %w", objName, err)

			break
		}

		fmt.Println("newChannelMember after create=========", debug.PrettyFormatStringJSON(newChannelMember))

		sfRes.EventChannelMembers[objName] = newChannelMember
	}

	res := &common.SubscriptionResult{
		// Salesforce is all or nothing for an object,
		// so if successful, we will subscribe to all events.
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
	}

	var rollbackError error

	if failError != nil {
		for objName, member := range sfRes.EventChannelMembers {
			if _, err := c.DeleteEventChannelMember(ctx, member.Id); err != nil {
				rollbackError = errors.Join(
					rollbackError,
					fmt.Errorf("failed to delete event channel member for object %s: %w",
						objName,
						err,
					),
				)
			} else {
				// remove the object from the map
				delete(sfRes.EventChannelMembers, objName)
			}
		}

		if rollbackError != nil {
			res.Status = common.SubscriptionStatusFailedToRollback

			for objName := range sfRes.EventChannelMembers {
				res.Objects = append(res.Objects, objName)
			}

			res.Result = sfRes

			// we still return the partial result along with the error
			return res, errors.Join(failError, rollbackError)
		}

		res.Events = nil
		res.Status = common.SubscriptionStatusFailed
		res.Result = sfRes

		return res, failError
	}

	res.Status = common.SubscriptionStatusSuccess
	res.Result = sfRes

	for objName := range sfRes.EventChannelMembers {
		res.Objects = append(res.Objects, objName)
	}

	return res, nil
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
//
//nolint:funlen,cyclop
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

	objectsToExcludeFromSubscription := []common.ObjectName{}
	objectsExcludeFromDelete := []common.ObjectName{}

	// collect objects to exclude from subscription
	for objName := range params.SubscriptionEvents {
		if _, ok := prevState.EventChannelMembers[objName]; ok {
			objectsToExcludeFromSubscription = append(objectsToExcludeFromSubscription, objName)
		}
	}

	// collect objects to exclude from delete
	for objName := range prevState.EventChannelMembers {
		if _, ok := params.SubscriptionEvents[objName]; ok {
			objectsExcludeFromDelete = append(objectsExcludeFromDelete, objName)
		}
	}

	// remove objects to exclude from subscription and delete
	for _, objName := range objectsToExcludeFromSubscription {
		delete(params.SubscriptionEvents, objName)
	}

	channelMembersToKeep := make(map[common.ObjectName]*EventChannelMember)

	// remove objects to exclude from delete
	for _, objName := range objectsExcludeFromDelete {
		channelMembersToKeep[objName] = prevState.EventChannelMembers[objName]
		delete(prevState.EventChannelMembers, objName)
	}

	objectsToDelete := []common.ObjectName{}

	// get list of objects to delete to remove from result of update after delete
	for objName := range prevState.EventChannelMembers {
		objectsToDelete = append(objectsToDelete, objName)
	}

	// rename the previous result to deleteParam for clarity
	// we will use this to delete the previous subscription
	deleteParams := *previousResult
	deleteParams.Result = prevState
	deleteParams.Objects = objectsToDelete

	// this is the delete step, but it looks for only object that were selected to delete
	// in objectsToDelete array, so we are still preserving some objects
	// that needs to remain in the subscription
	if err := c.DeleteSubscription(ctx, deleteParams); err != nil {
		return nil, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	// create new subscription
	createRes, err := c.Subscribe(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new objects: %w", err)
	}

	// for clarity, rename the state since we will return the object as the result of update
	newState := prevState
	// reset the ChannelMembers that was not deleted
	newState.EventChannelMembers = channelMembersToKeep

	//nolint:forcetypeassert
	// update the previous result with the new subscription result
	maps.Copy(newState.EventChannelMembers, createRes.Result.(*SubscribeResult).EventChannelMembers)

	// remove delete objects from the previous result to return
	for _, objName := range objectsToDelete {
		delete(newState.EventChannelMembers, objName)
	}

	objectsSubscribed := []common.ObjectName{}
	for objName := range newState.EventChannelMembers {
		objectsSubscribed = append(objectsSubscribed, objName)
	}

	res := &common.SubscriptionResult{
		Status: common.SubscriptionStatusSuccess,
		Result: newState,
		Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		},
		Objects: objectsSubscribed,
	}

	return res, nil
}
