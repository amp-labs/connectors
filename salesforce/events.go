package salesforce

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

type SFAPIResponseBody struct {
	Id       string        `json:"id"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Infos    []interface{} `json:"infos"`
	Warnings []interface{} `json:"warnings"`
}

type EventChannel struct {
	Id       string                `json:"Id,omitempty"` // nolint:tagliatelle
	FullName string                `json:"FullName"`     // nolint:tagliatelle
	Metadata *EventChannelMetadata `json:"Metadata"`     // nolint:tagliatelle
}

type EventChannelMetadata struct {
	ChannelType string `json:"channelType"`
	Label       string `json:"label"`
}

type EventRelayConfig struct {
	Id                      string                    `json:"Id,omitempty"`                      // nolint:tagliatelle
	FullName                string                    `json:"FullName,omitempty"`                // nolint:tagliatelle
	Metadata                *EventRelayConfigMetadata `json:"Metadata,omitempty"`                // nolint:tagliatelle
	DeveloperName           string                    `json:"DeveloperName,omitempty"`           // nolint:tagliatelle
	DestinationResourceName string                    `json:"DestinationResourceName,omitempty"` // nolint:tagliatelle
	EventChannel            string                    `json:"EventChannel,omitempty"`            // nolint:tagliatelle
}

type EventRelayConfigMetadata struct {
	DestinationResourceName string `json:"destinationResourceName,omitempty"`
	EventChannel            string `json:"eventChannel,omitempty"`
	State                   string `json:"state,omitempty"`
}

type EventChannelMember struct {
	Id       string                      `json:"Id,omitempty"` // nolint:tagliatelle
	FullName string                      `json:"FullName"`     // nolint:tagliatelle
	Metadata *EventChannelMemberMetadata `json:"Metadata"`     // nolint:tagliatelle
}

type EventChannelMemberMetadata struct {
	EventChannel   string `json:"eventChannel"`
	SelectedEntity string `json:"selectedEntity,omitempty"`
}

type ExternalCredentialParameters struct {
	ParameterName  string `json:"parameterName"`
	ParameterType  string `json:"parameterType"`
	ParameterValue string `json:"parameterValue"`
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.chatterapi.meta/chatterapi/connect_responses_external_credential.htm
type ExternalCredential struct {
	DeveloperName          string                          `json:"developerName"                    validate:"required"`
	MasterLabel            string                          `json:"masterLabel,omitempty"            validate:"required"`
	AuthenticationProtocol string                          `json:"authenticationProtocol,omitempty" validate:"required"`
	Parameters             []*ExternalCredentialParameters `json:"parameters,omitempty"             validate:"required"`

	Id                            string             `json:"id,omitempty"`
	AuthenticationProtocolVariant string             `json:"authenticationProtocolVariant,omitempty"`
	AuthenticationStatus          string             `json:"authenticationStatus,omitempty"`
	CreatedByNamespace            string             `json:"createdByNamespacePrefix,omitempty"`
	CustomHeaders                 []interface{}      `json:"customHeaders,omitempty"`
	Principals                    []interface{}      `json:"principals,omitempty"`
	RelatedNamedCredentials       []*NamedCredential `json:"relatedNamedCredentials,omitempty"`
	URL                           string             `json:"url,omitempty"`
}

type CalloutOptions struct {
	AllowMergeFieldsInBody      bool `json:"allowMergeFieldsInBody"`
	AllowMergeFieldsInHeader    bool `json:"allowMergeFieldsInHeader"`
	GenerateAuthorizationHeader bool `json:"generateAuthorizationHeader"`
}

type NamedCredential struct {
	DeveloperName       string                `json:"developerName"           validate:"required"`
	MasterLabel         string                `json:"masterLabel"             validate:"required"`
	ExternalCredentials []*ExternalCredential `json:"externalCredentials"     validate:"required"`
	CalloutOptions      *CalloutOptions       `json:"calloutOptions"          validate:"required"`
	CalloutStatus       string                `json:"calloutStatus,omitempty" validate:"required"`
	CalloutURL          string                `json:"calloutUrl"              validate:"required"`
	Type                string                `json:"type"                    validate:"required"`

	Id                 string        `json:"id,omitempty"`
	CreatedByNamespace string        `json:"createdByNamespacePrefix,omitempty"`
	CustomHeaders      []interface{} `json:"customHeaders,omitempty"`
	NetworkConnection  interface{}   `json:"networkConnection,omitempty"`
	Parameters         []interface{} `json:"parameters,omitempty"`
	URL                string        `json:"url,omitempty"`
}

type NamedCredentialMetadata struct {
	Endpoint                    string `json:"endpoint"`
	GenerateAuthorizationHeader bool   `json:"generateAuthorizationHeader"`
	Label                       string `json:"label"`
	PrincipalType               string `json:"principalType"`
	Protocol                    string `json:"protocol"`
}

type NamedCredentialLegacy struct {
	Id       string                   `json:"Id,omitempty"` // nolint:tagliatelle
	FullName string                   `json:"FullName"`     // nolint:tagliatelle
	Metadata *NamedCredentialMetadata `json:"Metadata"`     // nolint:tagliatelle
}

func (n *NamedCredentialLegacy) GetRemoteResource() string {
	return fmt.Sprint("callout:", n.FullName)
}

func (c *Connector) CreateEventChannel(ctx context.Context, channel *EventChannel) (*EventChannel, error) {
	location, err := joinURLPath(c.BaseURL, "tooling/sobjects/PlatformEventChannel")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, channel)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	if res.Warnings != nil && len(res.Warnings) > 0 {
		slog.Warn("CreateEventChannelMember", "warnings", res.Warnings)
	}

	channel.Id = res.Id

	return channel, nil
}

func (c *Connector) CreateEventChannelMember(
	ctx context.Context,
	member *EventChannelMember,
) (*EventChannelMember, error) {
	location, err := joinURLPath(c.BaseURL, "tooling/sobjects/PlatformEventChannelMember")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, member)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	if res.Warnings != nil && len(res.Warnings) > 0 {
		slog.Warn("CreateEventChannelMember", "warnings", res.Warnings)
	}

	member.Id = res.Id

	return member, nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) CreateEventRelayConfig(
	ctx context.Context,
	cfg *EventRelayConfig,
) (*EventRelayConfig, error) {
	location, err := joinURLPath(c.BaseURL, "tooling/sobjects/EventRelayConfig")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, cfg)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	cfg.Id = res.Id

	return cfg, nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) RunEventRelay(ctx context.Context, configId string) (*EventRelayConfig, error) {
	location, err := joinURLPath(c.BaseURL, "tooling/sobjects/EventRelayConfig", configId)
	if err != nil {
		return nil, err
	}

	config := &EventRelayConfig{
		Id: configId,
		Metadata: &EventRelayConfigMetadata{
			State: "RUN",
		},
	}

	resp, err := c.patch(ctx, location, config)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	config.Id = res.Id

	return config, nil
}

func (c *Connector) getOrganization(ctx context.Context) (map[string]*ajson.Node, error) {
	location, err := joinURLPath(c.BaseURL, "connect/organization")
	if err != nil {
		return nil, err
	}

	resp, err := c.get(ctx, location)
	if err != nil {
		return nil, err
	}

	return resp.Body.GetObject()
}

func (c *Connector) GetOrganizationId(ctx context.Context) (string, error) {
	org, err := c.getOrganization(ctx)
	if err != nil {
		return "", err
	}

	return org["orgId"].MustString(), nil
}

func (c *Connector) GetRemoteResouece(ctx context.Context, channelId string) (string, error) {
	orgId, err := c.GetOrganizationId(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("aws.partner/salesforce.com/%s/%s", orgId, channelId), nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.chatterapi.meta/chatterapi/connect_resources_named_credentials_external_credentials.htm
func (c *Connector) CreateExternalCredential(
	ctx context.Context,
	creds *ExternalCredential,
) (*ExternalCredential, error) {
	location, err := joinURLPath(c.BaseURL, "named-credentials/external-credentials")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, creds)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	creds.Id = res.Id

	return creds, nil
}

func (n *NamedCredential) GetRemoteResource() string {
	return fmt.Sprint("callout:", n.DeveloperName)
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.chatterapi.meta/chatterapi/connect_resources_setup_named_credentials.htm
func (c *Connector) CreateNamedCredential(ctx context.Context, creds *NamedCredential) (*NamedCredential, error) {
	location, err := joinURLPath(c.BaseURL, "named-credentials/named-credential-setup")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, creds)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	creds.Id = res.Id

	return creds, nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_namedcredential.htm
func (c *Connector) CreateNamedCredentialLegacy(ctx context.Context, creds *NamedCredentialLegacy) (*NamedCredentialLegacy, error) {
	location, err := joinURLPath(c.BaseURL, "tooling/sobjects/NamedCredential")
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, creds)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	creds.Id = res.Id

	return creds, nil
}

type Credential interface {
	GetRemoteResource() string
}
