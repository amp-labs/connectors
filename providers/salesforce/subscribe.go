package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type SubscribeParams struct {
	UniqueRef string `json:"uniqueRef" validate:"required"`
	Label     string `json:"label"     validate:"required"`
}

func (conn *Connector) Subscribe(ctx, params *common.SubscribeParams) (*common.SubscriptionResult, error) {
    if params.RegistrationResult == nil {
        return nil, fmt.Errorf("%w: missing RegistrationResult", errMissingParams)
    }

    if params.

	regstrationParams, ok := params.RegistrationResult.Result.(*ResultData)
	if !ok {
		return nil, fmt.Errorf("%w: expeted SubscribeParams.RegistrationResult.Result to be type '%T', but got '%T'", errInvalidRequestType, regstrationParams, params.RegistrationResult.Result)
	}

	for objName := range params.SubscriptionEvents {
		// eventName := GetChangeDataCaptureEventName(string(objName))
		// rawChannelName := GetRawChannelNameFromChannel()

		// channelMember := &EventChannelMember{
		//     FullName: GetChangeDataCaptureChannelMembershipName()
		// }
	}

	return nil, nil
}
