package salesforce

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

var errInvalidRequestType = errors.New("invalid request type")

type SalesforceRegistration struct {
	UniqueRef string `json:"uniqueIdentifier" validate:"required"`
	Label     string `json:"label"            validate:"required"`
	AwsArn    string `json:"awsArn"           validate:"required"`
}

type ResultData struct {
	eventChannel     *EventChannel
	namedCredential  *NamedCredential
	eventRelayConfig *EventRelayConfig
}

// func (rb *rollbackData) rollback(ctx context.Context, c *Connector) error {

// 	if rb.eventRelayConfig != nil {
// 		if err := c.DeleteEventRelayConfig(ctx, rb.eventRelayConfig); err != nil {
// 			slog.Error("failed to delete event relay config", "error", err)
// 		}
// 	}

// 	if rb.namedCredential != nil {
// 		if err := c.DeleteNamedCredential(ctx, rb.namedCredential); err != nil {
// 			slog.Error("failed to delete named credential", "error", err)
// 		}
// 	}
// 	if rb.eventChannel != nil {
// 		resp, err := c.DeleteEventChannel(ctx, rb.eventChannel.Id)
// 		if err != nil {
// 			slog.Error("failed to delete event channel", "error", err)
// 		}

// 		fmt.Println("DeleteEventChannel", resp)
// 	}

// 	return nil

// }

func (c *Connector) Register(
	ctx context.Context,
	params common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
	sfRegistration, ok := params.Request.(*SalesforceRegistration)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected '%T', but received '%T'",
			errInvalidRequestType,
			sfRegistration,
			params.Request,
		)
	}

	validate := validator.New()

	if err := validate.Struct(sfRegistration); err != nil {
		return nil, fmt.Errorf("invalid registration request: %w", err)
	}

	eventChannel, err := c.createEventChannel(ctx, sfRegistration)
	if err != nil {
		return nil, fmt.Errorf("failed to create event channel: %w", err)
	}

	result := &ResultData{
		eventChannel: eventChannel,
	}

	namedCred, err := c.createNamedCredential(ctx, sfRegistration)
	if err != nil {
		return nil, fmt.Errorf("failed to create named credential: %w", err)
	}

	result.namedCredential = namedCred

	evtCfg, err := c.createEventRelayConfing(
		ctx,
		sfRegistration,
		namedCred.DestinationResourceName(),
		eventChannel.FullName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event relay config: %w", err)
	}

	if err := c.RunEventRelay(ctx, evtCfg); err != nil {
		return nil, fmt.Errorf("failed to run event relay: %w", err)
	}

	result.eventRelayConfig = evtCfg

	return &common.RegistrationResult{
		RegistrationRef: result.eventRelayConfig.Id,
		Result:          result,
	}, nil
}

func (c *Connector) createEventChannel(ctx context.Context, reg *SalesforceRegistration) (*EventChannel, error) {
	channelName := GetChannelName(reg.UniqueRef)

	channel := &EventChannel{
		FullName: channelName,
		Metadata: &EventChannelMetadata{
			ChannelType: "data",
			Label:       reg.UniqueRef,
		},
	}

	return c.CreateEventChannel(ctx, channel)
}

func (c *Connector) createNamedCredential(ctx context.Context, reg *SalesforceRegistration) (*NamedCredential, error) {
	namedCred := &NamedCredential{
		FullName: reg.UniqueRef,
		Metadata: &NamedCredentialMetadata{
			GenerateAuthorizationHeader: true,
			Label:                       reg.Label,

			// below are legacy fields
			Endpoint:      reg.AwsArn,
			PrincipalType: "NamedUser",
			Protocol:      "NoAuthentication",
		},
	}

	return c.CreateNamedCredential(ctx, namedCred)
}

func (c *Connector) createEventRelayConfing(
	ctx context.Context,
	reg *SalesforceRegistration,
	destinationResource string,
	channelName string,
) (*EventRelayConfig, error) {
	config := &EventRelayConfig{
		FullName: reg.UniqueRef,
		Metadata: &EventRelayConfigMetadata{
			DestinationResourceName: destinationResource,
			EventChannel:            channelName,
		},
	}

	return c.CreateEventRelayConfig(ctx, config)
}

func IsCustomObject(objName string) bool {
	return strings.HasSuffix(objName, "__c")
}

func GetRawObjectName(objName string) string {
	return RemoveSuffix(objName, 3) //nolint:mnd
}

func GetChangeDataCaptureEventName(objName string) string {
	if IsCustomObject(objName) {
		return GetRawObjectName(objName) + "__ChangeEvent"
	}

	return objName + "%ChangeEvent"
}

func GetChannelName(rawChannelName string) string {
	return rawChannelName + "__chn"
}

func RemoveSuffix(objName string, suffixLength int) string {
	return objName[:len(objName)-suffixLength]
}

func GetRawChannelNameFromChannel(channelName string) string {
	if strings.HasSuffix(channelName, "__chn") {
		return RemoveSuffix(channelName, 5) //nolint:mnd
	}

	return channelName
}

func GetRawChannelNameFromObject(objectName string) string {
	if strings.HasSuffix(objectName, "__e") {
		return RemoveSuffix(objectName, 3) //nolint:mnd
	}

	return objectName
}

func GetChangeDataCaptureChannelMembershipName(rawChannelName string, eventName string) string {
	return rawChannelName + "_chn_" + eventName
}

func GetRawPEName(peName string) string {
	return RemoveSuffix(peName, 3) //nolint:mnd
}
