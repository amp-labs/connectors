package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

type SubscribeParams struct {
	UniqueRef string `json:"uniqueRef" validate:"required"`
	Label     string `json:"label"     validate:"required"`
}

func (conn *Connector) Subscribe(ctx, params *common.SubscribeParams) (*common.SubscriptionResult, error) {
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
		return nil, fmt.Errorf("%w: expeted SubscribeParams.RegistrationResult.Result to be type '%T', but got '%T'", errInvalidRequestType, regstrationParams, params.RegistrationResult.Result)
	}

	for objName := range params.SubscriptionEvents {
		eventName := GetChangeDataCaptureEventName(string(objName))
		rawChannelName := GetRawChannelNameFromChannel(regstrationParams.EventChannel.FullName)

		channelMember := &EventChannelMember{
			FullName: GetChangeDataCaptureChannelMembershipName(rawChannelName, eventName),
			Metadata: &EventChannelMemberMetadata{
				EventChannel:   GetChannelName(rawChannelName),
				SelectedEntity: eventName,
			},
		}
	}

	return nil, nil
}
