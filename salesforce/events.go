package salesforce

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// nolint: revive
const (
	AllowedManagedPackageNamespaces NamedCredentialParameterType = "AllowedManagedPackageNamespaces"
	ClientCertificate               NamedCredentialParameterType = "ClientCertificate"
	HttpHeader                      NamedCredentialParameterType = "HttpHeader"
	OutboundNetworkConnection       NamedCredentialParameterType = "OutboundNetworkConnection"
	Url                             NamedCredentialParameterType = "Url"
)

type SFAPIResponseBody struct {
	Id       string        `json:"id"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Infos    []interface{} `json:"infos"`
	Warnings []interface{} `json:"warnings"`
}

// nolint:tagliatelle
type EventChannel struct {
	Id       string                `json:"Id,omitempty"`
	FullName string                `json:"FullName"`
	Metadata *EventChannelMetadata `json:"Metadata"`
}

type EventChannelMetadata struct {
	ChannelType string `json:"channelType"`
	Label       string `json:"label"`
}

// nolint:tagliatelle
type EventRelayConfig struct {
	Id                      string                    `json:"Id,omitempty"`
	FullName                string                    `json:"FullName,omitempty"`
	Metadata                *EventRelayConfigMetadata `json:"Metadata,omitempty"`
	DeveloperName           string                    `json:"DeveloperName,omitempty"`
	DestinationResourceName string                    `json:"DestinationResourceName,omitempty"`
	EventChannel            string                    `json:"EventChannel,omitempty"`
}

type EventRelayConfigMetadata struct {
	DestinationResourceName string `json:"destinationResourceName,omitempty"`
	EventChannel            string `json:"eventChannel,omitempty"`
	State                   string `json:"state,omitempty"`
}

// nolint:tagliatelle
type EventChannelMember struct {
	Id       string                      `json:"Id,omitempty"`
	FullName string                      `json:"FullName"`
	Metadata *EventChannelMemberMetadata `json:"Metadata"`
}

type EventChannelMemberMetadata struct {
	EventChannel   string `json:"eventChannel"`
	SelectedEntity string `json:"selectedEntity,omitempty"`
}

type NamedCredentialParameterType string

// nolint: tagliatelle
type ToolingApiBaseParams struct {
	DeveloperName   string `json:"DeveloperName,omitempty"`
	Language        string `json:"Language,omitempty"`
	ManageableState string `json:"ManageableState,omitempty"`
	MasterLabel     string `json:"MasterLabel,omitempty"`
	NamespacePrefix string `json:"NamespacePrefix,omitempty"`
}

type CalloutOptions struct {
	AllowMergeFieldsInBody      bool `json:"allowMergeFieldsInBody"`
	AllowMergeFieldsInHeader    bool `json:"allowMergeFieldsInHeader"`
	GenerateAuthorizationHeader bool `json:"generateAuthorizationHeader"`
}

type NamedCredentialParameter struct {
	Certificate               string `json:"certificate"`
	Description               string `json:"description"`
	ExternalCredential        string `json:"externalCredential"`
	OutboundNetworkConnection string `json:"outboundNetworkConnection"`
	ParameterName             string `json:"parameterName"`
	ParameterType             string `json:"parameterType"`
	ParameterValue            string `json:"parameterValue"`
	SequenceNumber            int    `json:"sequenceNumber"`
}

// nolint: lll
type NamedCredentialMetadata struct {
	AllowMergeFieldsInBody      bool                        `json:"allowMergeFieldsInBody,omitempty"`
	AllowMergeFieldsInHeader    bool                        `json:"allowMergeFieldsInHeader,omitempty"`
	GenerateAuthorizationHeader bool                        `json:"generateAuthorizationHeader,omitempty"`
	FullName                    string                      `json:"fullName,omitempty"                    validate:"required"`
	Label                       string                      `json:"label,omitempty"`
	NamedCredentialParameters   []*NamedCredentialParameter `json:"namedCredentialParameters,omitempty"`
	NamedCredentialType         string                      `json:"namedCredentialType,omitempty"`

	// Below are deprecated fields, but still in use in SF
	AuthProvider             string `json:"authProvider,omitempty"`
	AuthTokenEndpointUrl     string `json:"authTokenEndpointUrl,omitempty"` // nolint: revive
	AwsAccessKey             string `json:"awsAccessKey,omitempty"`
	AwsAccessSecret          string `json:"awsAccessSecret,omitempty"`
	AwsRegion                string `json:"awsRegion,omitempty"`
	AwsService               string `json:"awsService,omitempty"`
	Certificate              string `json:"certificate,omitempty"`
	Endpoint                 string `json:"endpoint,omitempty"`
	JwtAudience              string `json:"jwtAudience,omitempty"`
	JwtFormulaSubject        string `json:"jwtFormulaSubject,omitempty"`
	JwtIssuer                string `json:"jwtIssuer,omitempty"`
	JwtSigningCertificate    string `json:"jwtSigningCertificate,omitempty"`
	JwtTextSubject           string `json:"jwtTextSubject,omitempty"`
	JwtValidityPeriodSeconds int    `json:"jwtValidityPeriodSeconds,omitempty"`
	OauthRefreshToken        string `json:"oauthRefreshToken,omitempty"`
	OauthScope               string `json:"oauthScope,omitempty"`
	OauthToken               string `json:"oauthToken,omitempty"`
	Password                 string `json:"password,omitempty"`
	PrincipalType            string `json:"principalType,omitempty"`
	Protocol                 string `json:"protocol,omitempty"`
	Username                 string `json:"username,omitempty"`
}

// nolint:tagliatelle
type NamedCredential struct {
	FullName string                   `json:"FullName"`
	Metadata *NamedCredentialMetadata `json:"Metadata"`

	// below exist in response, but not in request
	Id string `json:"Id,omitempty"`
}

func (n *NamedCredential) DestinationResourceName() string {
	return fmt.Sprint("callout:", n.FullName)
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_platformeventchannel.htm
func (c *Connector) CreateEventChannel(ctx context.Context, channel *EventChannel) (*EventChannel, error) {
	res, err := c.postToSFAPI(ctx, channel, "tooling/sobjects/PlatformEventChannel", "PlatformEventChannel")
	if err != nil {
		return nil, err
	}

	channel.Id = res.Id

	return channel, nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_platformeventchannelmember.htm
func (c *Connector) CreateEventChannelMember(
	ctx context.Context,
	member *EventChannelMember,
) (*EventChannelMember, error) {
	res, err := c.postToSFAPI(ctx, member, "tooling/sobjects/PlatformEventChannelMember", "EventChannelMember")
	if err != nil {
		return nil, err
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
	res, err := c.postToSFAPI(ctx, cfg, "/tooling/sobjects/EventRelayConfig", "EventRelayConfig")
	if err != nil {
		return nil, err
	}

	cfg.Id = res.Id

	return cfg, nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) RunEventRelay(ctx context.Context, cfg *EventRelayConfig) error {
	location, err := joinURLPath("tooling/sobjects/EventRelayConfig", cfg.Id)
	if err != nil {
		return err
	}

	config := &EventRelayConfig{
		FullName: cfg.FullName,
		Metadata: &EventRelayConfigMetadata{
			State: "RUN",
		},
	}

	_, err = c.patch(ctx, location, config) // patch returns no content with 204. If it fails, it will return an error.
	if err != nil {
		slog.Error("Run EventRelayConfig", "error", err)

		return err
	}

	return nil
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.chatterapi.meta/chatterapi/connect_responses_organization.htm?q=organization
func (c *Connector) getOrganization(ctx context.Context) (map[string]*ajson.Node, error) {
	resp, err := c.get(ctx, "connect/organization")
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

func GetRemoteResource(orgId, channelId string) string {
	return fmt.Sprintf("aws.partner/salesforce.com/%s/%s", orgId, channelId)
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_namedcredential.htm
func (c *Connector) CreateNamedCredential(ctx context.Context, creds *NamedCredential) (*NamedCredential, error) {
	res, err := c.postToSFAPI(ctx, creds, "/tooling/sobjects/NamedCredential", "NamedCredential")
	if err != nil {
		return nil, err
	}

	creds.Id = res.Id

	return creds, nil
}

type Credential interface {
	DestinationResourceName() string
}

func (c *Connector) postToSFAPI(ctx context.Context, body any, path string, entity string) (*SFAPIResponseBody, error) {
	location, err := joinURLPath(c.BaseURL, path)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, location, body)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if res.Warnings != nil && len(res.Warnings) > 0 {
		slog.Warn(entity, "warnings", res.Warnings)
	}

	return res, err
}
