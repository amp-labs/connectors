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

var (
	errInvalidRequestType = errors.New("invalid request type")
	errMissingParams      = errors.New("missing required parameters")
)

type RegistrationParams struct {
	// UniqueRef is a unique reference for the registration.
	// It is used to create unique names for the Salesforce objects.
	UniqueRef string `json:"uniqueRef" validate:"required"`
	Label     string `json:"label"     validate:"required"`
	AwsArn    string `json:"awsArn"    validate:"required"`
}

type ResultData struct {
	EventChannel *EventChannel `json:"eventChannel" validate:"required"`
	// structonly
	NamedCredential  *NamedCredential  `json:"namedCredential"  validate:"required"`
	EventRelayConfig *EventRelayConfig `json:"eventRelayConfig" validate:"required"`
}

func (c *Connector) rollbackRegister(ctx context.Context, res *ResultData) error {
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

	sfParams, ok := params.Request.(*RegistrationParams)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected '%T', but received '%T'",
			errInvalidRequestType,
			sfParams,
			params.Request,
		)
	}

	result, err := c.register(ctx, sfParams)
	if err != nil {
		if rollbackErr := c.rollbackRegister(ctx, result); rollbackErr != nil {
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
	params *RegistrationParams,
) (*ResultData, error) {
	result := &ResultData{}

	eventChannel, err := c.createEventChannel(ctx, params)
	if err != nil {
		return result, fmt.Errorf("failed to create event channel: %w", err)
	}

	result.EventChannel = eventChannel

	namedCred, err := c.createNamedCredential(ctx, params)
	if err != nil {
		return result, fmt.Errorf("failed to create named credential: %w", err)
	}

	result.NamedCredential = namedCred

	evtCfg, err := c.createEventRelayConfing(
		ctx,
		params,
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

func (c *Connector) DeleteRegistration(ctx context.Context, registration *common.RegistrationResult) error {
	if registration == nil {
		return fmt.Errorf("%w: registration is null", errMissingParams)
	}

	if registration.Result == nil {
		return fmt.Errorf("%w: registration result is null", errMissingParams)
	}

	validate := validator.New()
	if err := validate.Struct(registration.Result); err != nil {
		return fmt.Errorf("invalid registration result: %w", err)
	}

	result, ok := registration.Result.(*ResultData)
	if !ok {
		return fmt.Errorf(
			"%w: expected result type '%T', but received '%T'",
			errInvalidRequestType,
			result,
			registration.Result,
		)
	}

	return c.rollbackRegister(ctx, result)
}

func (c *Connector) createEventChannel(ctx context.Context, params *RegistrationParams) (*EventChannel, error) {
	channelName := GetChannelName(params.UniqueRef)

	channel := &EventChannel{
		FullName: channelName,
		Metadata: &EventChannelMetadata{
			ChannelType: "data",
			Label:       params.UniqueRef,
		},
	}

	return c.CreateEventChannel(ctx, channel)
}

func (c *Connector) createNamedCredential(ctx context.Context, params *RegistrationParams) (*NamedCredential, error) {
	namedCred := &NamedCredential{
		FullName: params.UniqueRef,
		Metadata: &NamedCredentialMetadata{
			GenerateAuthorizationHeader: true,
			Label:                       params.Label,

			// below are legacy fields
			Endpoint:      params.AwsArn,
			PrincipalType: "NamedUser",
			Protocol:      "NoAuthentication",
		},
	}

	return c.CreateNamedCredential(ctx, namedCred)
}

func (c *Connector) createEventRelayConfing(
	ctx context.Context,
	params *RegistrationParams,
	destinationResource string,
	channelName string,
) (*EventRelayConfig, error) {
	config := &EventRelayConfig{
		FullName: params.UniqueRef,
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
