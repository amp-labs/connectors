package salesforce

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/go-playground/validator"
)

var (
	errInvalidRequestType = errors.New("invalid request type")
	errMissingParams      = errors.New("missing required parameters")
)

type RegistrationParams struct {
	// UniqueRef is a unique reference for the registration.
	// It is used to create unique names for the Salesforce objects.
	UniqueRef             string `json:"uniqueRef"             validate:"required"`
	Label                 string `json:"label"                 validate:"required"`
	AwsNamedCredentialArn string `json:"awsNamedCredentialArn" validate:"required"`
}

type ResultData struct {
	EventChannel     *EventChannel     `json:"eventChannel"     validate:"required"`
	NamedCredential  *NamedCredential  `json:"namedCredential"  validate:"required"`
	EventRelayConfig *EventRelayConfig `json:"eventRelayConfig" validate:"required"`
}

func (c *Connector) EmptyRegistrationParams() *common.SubscriptionRegistrationParams {
	return &common.SubscriptionRegistrationParams{
		Request: &RegistrationParams{},
	}
}

func (c *Connector) EmptyRegistrationResult() *common.RegistrationResult {
	return &common.RegistrationResult{
		Result: &ResultData{},
	}
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
			logging.Logger(ctx).Error("failed to delete named credential", "error", err)

			return fmt.Errorf("failed to delete named credential: %w", err)
		}
	}

	if res.EventChannel != nil {
		_, err := c.DeleteEventChannel(ctx, res.EventChannel.Id)
		if err != nil {
			logging.Logger(ctx).Error("failed to delete event channel", "error", err)

			return fmt.Errorf("failed to delete event channel: %w", err)
		}
	}

	return nil
}

func (c *Connector) Register(
	ctx context.Context,
	params common.SubscriptionRegistrationParams,
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
				Status: common.RegistrationStatusFailedToRollback,
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

	evtCfg, err := c.createEventRelayConfig(
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

// DeleteRegistration will delete the Salesforce objects created during registration.
// It will delete the EventRelayConfig, NamedCredential, and EventChannel.
func (c *Connector) DeleteRegistration(ctx context.Context, registration common.RegistrationResult) error {
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
			Endpoint:      params.AwsNamedCredentialArn,
			PrincipalType: "NamedUser",
			Protocol:      "NoAuthentication",
		},
	}

	return c.CreateNamedCredential(ctx, namedCred)
}

func (c *Connector) createEventRelayConfig(
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

	return objName + "ChangeEvent"
}

func GetChannelName(rawChannelName string) string {
	return rawChannelName + "__chn"
}

func RemoveSuffix(objName string, suffixLength int) string {
	if len(objName) < suffixLength {
		return ""
	}

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
