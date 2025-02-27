package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

type SubscribeResult struct {
	EventChannelMembers map[common.ObjectName]*EventChannelMember
}

// Subscribe subscribes to the events for the given objects.
// It creates event channel members for each object in the subscription.
// If any of the event channel members fail to be created, it will rollback the operation.
// If the rollback fails, it will return the partial result along with the error.
// If the rollback is successful, it will return the original error on object.
// Registration is required prior to subscribing.
//
//nolint:funlen,cyclop
func (conn *Connector) Subscribe(
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

	regstrationParams, ok := params.RegistrationResult.Result.(*ResultData)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected SubscribeParams.RegistrationResult.Result to be type '%T', but got '%T'", errInvalidRequestType,
			regstrationParams,
			params.RegistrationResult.Result,
		)
	}

	sfRes := &SubscribeResult{
		EventChannelMembers: make(map[common.ObjectName]*EventChannelMember),
	}

	var failError error

	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(regstrationParams.EventChannel.FullName)

		channelMetadata := &EventChannelMemberMetadata{
			EventChannel:   GetChannelName(rawChannelName),
			SelectedEntity: eventName,
		}
		channelMember := &EventChannelMember{
			FullName: GetChangeDataCaptureChannelMembershipName(rawChannelName, eventName),
			Metadata: channelMetadata,
		}

		newChannelMember, err := conn.CreateEventChannelMember(ctx, channelMember)
		if err != nil {
			failError = fmt.Errorf("failed to create event channel member for object %s, %w", objName, err)

			break
		}

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
			if _, err := conn.DeleteEventChannelMember(ctx, member.Id); err != nil {
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
func (conn *Connector) DeleteSubscription(ctx context.Context, params common.SubscriptionResult) error {
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
		if _, err := conn.DeleteEventChannelMember(ctx, member.Id); err != nil {
			return fmt.Errorf("failed to delete event channel member '%s': %w", objectName, err)
		}
	}

	return nil
}

func (conn *Connector) GetSubscriptionResultUnMarshalFunc() common.UnmarshalFunc {
	return subscriptionResultUnMarshalFunc
}

func subscriptionResultUnMarshalFunc(data []byte) (any, error) {
	return common.Unmarshal[SubscribeResult](data)
}
