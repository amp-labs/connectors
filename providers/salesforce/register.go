package salesforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

var errInvalidRequestType = errors.New("invalid request type")

type SalesforceRegistration struct {
	UniqueRef string `json:"uniqueRef" validate:"required"`
	Label     string `json:"label"     validate:"required"`
	AwsArn    string `json:"awsArn"    validate:"required"`
}

type ResultData struct {
	EventChannel     *EventChannel
	NamedCredential  *NamedCredential
	EventRelayConfig *EventRelayConfig
}

func (c *Connector) RollbackRegister(ctx context.Context, res *ResultData) error {
	if res.EventRelayConfig != nil {
		_, err := c.DeleteEventRelayConfig(ctx, res.EventRelayConfig.Id)
		if err != nil {
			return fmt.Errorf("failed to delete event relay config: %w", err)
		}
	}

	if res.NamedCredential != nil {
		_, err := c.DeleteNamedCredential(ctx, res.NamedCredential.Id)
		if err != nil {
			slog.Error("failed to delete named credential", "error", err)

			return fmt.Errorf("failed to delete named credential: %w", err)
		}
	}

	if res.EventChannel != nil {
		_, err := c.DeleteEventChannel(ctx, res.EventChannel.Id)
		if err != nil {
			slog.Error("failed to delete event channel", "error", err)

			return fmt.Errorf("failed to delete event channel: %w", err)
		}
	}

	return nil
}

func (c *Connector) Register(
	ctx context.Context,
	params *common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		return nil, fmt.Errorf("invalid registration params: %w", err)
	}

	sfRegistration, ok := params.Request.(*SalesforceRegistration)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected '%T', but received '%T'",
			errInvalidRequestType,
			sfRegistration,
			params.Request,
		)
	}

	result, err := c.register(ctx, sfRegistration)
	if err != nil {
		if rollbackErr := c.RollbackRegister(ctx, result); rollbackErr != nil {
			return &common.RegistrationResult{
				Status: common.RegistrationStatusError,
			}, errors.Join(rollbackErr, err)
		}

		return &common.RegistrationResult{
			Status: common.RegistrationStatusFailed,
		}, err
	}

	return &common.RegistrationResult{
		RegistrationRef: result.EventRelayConfig.Id,
		Result:          result,
		Status:          common.RegistrationStatusSuccess,
	}, err
}

func (c *Connector) register(
	ctx context.Context,
	sfRegistration *SalesforceRegistration,
) (*ResultData, error) {
	result := &ResultData{}

	eventChannel, err := c.createEventChannel(ctx, sfRegistration)
	if err != nil {
		return result, fmt.Errorf("failed to create event channel: %w", err)
	}

	result.EventChannel = eventChannel

	namedCred, err := c.createNamedCredential(ctx, sfRegistration)
	if err != nil {
		return result, fmt.Errorf("failed to create named credential: %w", err)
	}

	result.NamedCredential = namedCred

	evtCfg, err := c.createEventRelayConfing(
		ctx,
		sfRegistration,
		namedCred.DestinationResourceName(),
		eventChannel.FullName,
	)
	if err != nil {
		return result, fmt.Errorf("failed to create event relay config: %w", err)
	}

	result.EventRelayConfig = evtCfg

	if err := c.RunEventRelay(ctx, evtCfg); err != nil {
		return result, fmt.Errorf("failed to run event relay: %w", err)
	}

	return result, nil
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
