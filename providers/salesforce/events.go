package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
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
	Id       string `json:"id"`
	Success  bool   `json:"success"`
	Errors   []any  `json:"errors"`
	Infos    []any  `json:"infos"`
	Warnings []any  `json:"warnings"`
}

// EventChannel
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

// EventRelayConfig
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

// EventChannelMember
// nolint:tagliatelle
type EventChannelMember struct {
	Id       string                      `json:"Id,omitempty"`
	FullName string                      `json:"FullName"`
	Metadata *EventChannelMemberMetadata `json:"Metadata"`
}

type EventChannelMemberMetadata struct {
	EventChannel   string `json:"eventChannel"`
	SelectedEntity string `json:"selectedEntity,omitempty"`
	// EnrichedFields are the fields are list of fields that are used in filter expression
	EnrichedFields []*EnrichedField `json:"enrichedFields"`
	// Filter expression is used to filter COC events to reduce the number of events
	FilterExpression string `json:"filterExpression"`
}

type EnrichedField struct {
	Name string `json:"name"`
}

type NamedCredentialParameterType string

// ToolingApiBaseParams
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

// NamedCredentialMetadata
// nolint: lll
type NamedCredentialMetadata struct {
	AllowMergeFieldsInBody      bool                        `json:"allowMergeFieldsInBody,omitempty"`
	AllowMergeFieldsInHeader    bool                        `json:"allowMergeFieldsInHeader,omitempty"`
	GenerateAuthorizationHeader bool                        `json:"generateAuthorizationHeader,omitempty"`
	FullName                    string                      `json:"fullName,omitempty"`
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

// NamedCredential
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

// CreateEventChannel .
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_platformeventchannel.htm
func (c *Connector) CreateEventChannel(ctx context.Context, channel *EventChannel) (*EventChannel, error) {
	res, err := c.postToSFAPI(ctx, channel, "tooling/sobjects/PlatformEventChannel", "PlatformEventChannel")
	if err != nil {
		if isDuplicateDeveloperName(err) {
			// PlatformEventChannel.DeveloperName is FullName without the "__chn" suffix.
			developerName := strings.TrimSuffix(channel.FullName, "__chn")

			return recoverDuplicateByDeveloperName[EventChannel](ctx, c, "PlatformEventChannel", developerName)
		}

		return nil, err
	}

	channel.Id = res.Id

	return channel, nil
}

func (c *Connector) DeleteEventChannel(ctx context.Context, channelId string) (*common.JSONHTTPResponse, error) {
	return c.deleteToSFAPI(ctx, "tooling/sobjects/PlatformEventChannel/"+channelId, "PlatformEventChannel")
}

// CreateEventChannelMember
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_platformeventchannelmember.htm
func (c *Connector) CreateEventChannelMember(
	ctx context.Context,
	member *EventChannelMember,
) (*EventChannelMember, error) {
	res, err := c.postToSFAPI(ctx, member, "tooling/sobjects/PlatformEventChannelMember", "EventChannelMember")
	if err != nil {
		if isDuplicateDeveloperName(err) {
			return recoverDuplicateByDeveloperName[EventChannelMember](ctx, c, "PlatformEventChannelMember", member.FullName)
		}

		return nil, err
	}

	member.Id = res.Id

	return member, nil
}

// UpdateEventChannelMember updates an existing PlatformEventChannelMember via PATCH.
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_platformeventchannelmember.htm
//
// Salesforce's Tooling API treats the PATCH body as a full Metadata replacement:
// EventChannel and SelectedEntity are REQUIRED (a body without them returns
// "Required field is missing: selectedEntity") but also IMMUTABLE (a body
// whose value differs from what Salesforce has stored returns "Update is not
// supported for the selectedEntity field on platform event channel members.").
//
// To avoid the case-mismatch hazard — Salesforce normalizes the stored
// SelectedEntity / EventChannel (e.g. "account" → "Account__ChangeEvent")
// while our locally-cached prevState may still hold the caller's original
// casing — we first GET the live record to pick up its canonical immutable
// fields, then PATCH using those values plus the caller-supplied mutable
// fields (FilterExpression and EnrichedFields). The extra round trip is
// worth it: passing a stale SelectedEntity tanks the whole UpdateSubscription
// with a 400 even though the value didn't semantically change.
func (c *Connector) UpdateEventChannelMember(
	ctx context.Context,
	member *EventChannelMember,
) (*EventChannelMember, error) {
	if member == nil || member.Metadata == nil {
		return nil, errEventChannelMemberNilInput
	}

	current, err := getToolingEntityByID[EventChannelMember](
		ctx, c, "PlatformEventChannelMember", member.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current EventChannelMember %s before PATCH: %w", member.Id, err)
	}

	if current == nil || current.Metadata == nil {
		return nil, fmt.Errorf("%w: id=%s", errEventChannelMemberMissingMD, member.Id)
	}

	// Build the PATCH body using Salesforce's canonical immutable values for
	// EventChannel and SelectedEntity, plus the caller's intended new values
	// for the mutable fields.
	body := &EventChannelMember{
		FullName: current.FullName,
		Metadata: &EventChannelMemberMetadata{
			EventChannel:     current.Metadata.EventChannel,
			SelectedEntity:   current.Metadata.SelectedEntity,
			FilterExpression: member.Metadata.FilterExpression,
			EnrichedFields:   member.Metadata.EnrichedFields,
		},
	}

	_, err = c.patchToSFAPI(ctx, body,
		"tooling/sobjects/PlatformEventChannelMember/"+member.Id, "EventChannelMember")
	if err != nil {
		return nil, err
	}

	// Reflect the canonical immutable values back into the caller's member so
	// downstream consumers (e.g. diff.channelMembersExisting) see the
	// authoritative casing.
	member.FullName = current.FullName
	member.Metadata.EventChannel = current.Metadata.EventChannel
	member.Metadata.SelectedEntity = current.Metadata.SelectedEntity

	return member, nil
}

func (c *Connector) DeleteEventChannelMember(ctx context.Context, memberId string) (*common.JSONHTTPResponse, error) {
	return c.deleteToSFAPI(ctx, "tooling/sobjects/PlatformEventChannelMember/"+memberId, "EventChannelMember")
}

// CreateEventRelayConfig .
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) CreateEventRelayConfig(
	ctx context.Context,
	cfg *EventRelayConfig,
) (*EventRelayConfig, error) {
	res, err := c.postToSFAPI(ctx, cfg, "/tooling/sobjects/EventRelayConfig", "EventRelayConfig")
	if err != nil {
		if isDuplicateDeveloperName(err) {
			return recoverDuplicateByDeveloperName[EventRelayConfig](ctx, c, "EventRelayConfig", cfg.FullName)
		}

		return nil, err
	}

	cfg.Id = res.Id

	return cfg, nil
}

func (c *Connector) DeleteEventRelayConfig(ctx context.Context, cfgId string) (*common.JSONHTTPResponse, error) {
	return c.deleteToSFAPI(ctx, "tooling/sobjects/EventRelayConfig/"+cfgId, "EventRelayConfig")
}

// RunEventRelay
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) RunEventRelay(ctx context.Context, cfg *EventRelayConfig) error {
	url, err := c.getURLEventRelayConfig(cfg.Id)
	if err != nil {
		return err
	}

	err = c.patchEventRelayState(ctx, url.String(), cfg.FullName)
	if err != nil {
		// In a namespaced org (scratch/packaging orgs), Salesforce prepends the
		// org namespace to the full name of metadata it creates (e.g. the
		// EventRelayConfig created as "amp_x" becomes "acme__amp_x"). A PATCH
		// that echoes the un-namespaced full name is then rejected. The error
		// reports the expected namespaced full name, so adopt it and retry once.
		expected, ok := fullNameMismatch(err, cfg.FullName)
		if !ok {
			return fmt.Errorf("error running event relay: %w", err)
		}

		logging.Logger(ctx).Info("event relay full name mismatch; retrying with namespaced full name",
			"provided", cfg.FullName, "expected", expected)

		if retryErr := c.patchEventRelayState(ctx, url.String(), expected); retryErr != nil {
			// Join the retry error with the original mismatch error so both the
			// namespace-corrected failure and the initial cause are preserved.
			return fmt.Errorf("error running event relay after namespace retry: %w",
				errors.Join(retryErr, err))
		}

		// Persist the corrected full name so callers store the namespaced value.
		cfg.FullName = expected
	}

	cfg.Metadata.State = "RUN"

	return nil
}

// patchEventRelayState PATCHes the EventRelayConfig at url to the RUN state
// using the given full name. Salesforce returns 204 No Content on success.
func (c *Connector) patchEventRelayState(ctx context.Context, url, fullName string) error {
	config := &EventRelayConfig{
		FullName: fullName,
		Metadata: &EventRelayConfigMetadata{
			State: "RUN",
		},
	}

	_, err := c.Client.Patch(ctx, url, config)

	return err
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.chatterapi.meta/chatterapi/connect_responses_organization.htm?q=organization
func (c *Connector) getOrganization(ctx context.Context) (map[string]*ajson.Node, error) {
	url, err := c.getRestApiURL("connect/organization")
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, fmt.Errorf("cannot get organization %w", common.ErrEmptyJSONHTTPResponse)
	}

	return body.GetObject()
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

// CreateNamedCredential
// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_namedcredential.htm
func (c *Connector) CreateNamedCredential(ctx context.Context, creds *NamedCredential) (*NamedCredential, error) {
	res, err := c.postToSFAPI(ctx, creds, "/tooling/sobjects/NamedCredential", "NamedCredential")
	if err != nil {
		if isDuplicateDeveloperName(err) {
			return recoverDuplicateByDeveloperName[NamedCredential](ctx, c, "NamedCredential", creds.FullName)
		}

		return nil, err
	}

	creds.Id = res.Id

	return creds, nil
}

func (c *Connector) DeleteNamedCredential(ctx context.Context, credId string) (*common.JSONHTTPResponse, error) {
	return c.deleteToSFAPI(ctx, "tooling/sobjects/NamedCredential/"+credId, "NamedCredential")
}

type Credential interface {
	DestinationResourceName() string
}

func (c *Connector) postToSFAPI(ctx context.Context, body any, path string, entity string) (*SFAPIResponseBody, error) {
	location, err := c.getRestApiURL(path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(ctx, location.String(), body)
	if err != nil {
		return nil, err
	}

	res, err := common.UnmarshalJSON[SFAPIResponseBody](resp)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	if len(res.Warnings) > 0 {
		logging.Logger(ctx).Warn(entity, "warnings", res.Warnings)
	}

	return res, nil
}

func (c *Connector) patchToSFAPI(
	ctx context.Context, body any, path string, entity string,
) (*common.JSONHTTPResponse, error) {
	location, err := c.getRestApiURL(path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Patch(ctx, location.String(), body)
	if err != nil {
		return nil, fmt.Errorf("error updating %s: %w", entity, err)
	}

	return resp, nil
}

func (c *Connector) deleteToSFAPI(ctx context.Context, path string, entity string) (*common.JSONHTTPResponse, error) {
	location, err := c.getRestApiURL(path)
	if err != nil {
		return nil, err
	}

	// Check the entity exists before deleting. If it's already gone (404),
	// treat the delete as a no-op so callers can retry safely.
	if _, err := c.Client.Get(ctx, location.String()); err != nil {
		var httpErr *common.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound {
			logging.Logger(ctx).Info("skipping delete, entity not found", "entity", entity)

			return nil, nil //nolint:nilnil
		}

		return nil, fmt.Errorf("error checking %s before delete: %w", entity, err)
	}

	resp, err := c.Client.Delete(ctx, location.String())
	if err != nil {
		return nil, fmt.Errorf("error deleting %s: %w", entity, err)
	}

	return resp, nil
}

const errCodeDuplicateDeveloperName = "DUPLICATE_DEVELOPER_NAME"

var (
	errToolingQueryNoRecordsField  = errors.New("tooling/query response missing 'records' field")
	errToolingEntityNotFound       = errors.New("no tooling entity found by FullName")
	errToolingRecordMissingID      = errors.New("tooling/query record missing 'Id' field")
	errEventChannelMemberNilInput  = errors.New("UpdateEventChannelMember: member or metadata is nil")
	errEventChannelMemberMissingMD = errors.New("UpdateEventChannelMember: GET returned empty metadata")
)

// CustomFieldExists reports whether a custom field with the given API
// name (must include the __c suffix) exists on the named Salesforce object.
//
// Implementation queries the Tooling API CustomField sobject by TableEnumOrId
// and DeveloperName. CustomField.DeveloperName stores the suffix-less form, so
// the trailing __c is stripped before the query.
func (c *Connector) CustomFieldExists(
	ctx context.Context, objectName, fieldAPIName string,
) (bool, error) {
	developerName := strings.TrimSuffix(fieldAPIName, "__c")
	soql := fmt.Sprintf(
		"SELECT Id FROM CustomField WHERE TableEnumOrId = '%s' AND DeveloperName = '%s'",
		escapeSOQLString(objectName), escapeSOQLString(developerName),
	)

	return c.toolingEntityExists(ctx, soql)
}

// toolingEntityExists runs the given Tooling API SOQL query and reports whether
// it returned at least one record. Used by the existence-check helpers to power
// the manual-creation flow.
func (c *Connector) toolingEntityExists(ctx context.Context, soql string) (bool, error) {
	location, err := c.getRestApiURL("tooling/query")
	if err != nil {
		return false, err
	}

	location.WithQueryParam("q", soql)

	resp, err := c.Client.Get(ctx, location.String())
	if err != nil {
		return false, err
	}

	body, ok := resp.Body()
	if !ok {
		return false, common.ErrEmptyJSONHTTPResponse
	}

	obj, err := body.GetObject()
	if err != nil {
		return false, err
	}

	recordsNode, exists := obj["records"]
	if !exists {
		return false, fmt.Errorf("%w: soql=%q", errToolingQueryNoRecordsField, soql)
	}

	records, err := recordsNode.GetArray()
	if err != nil {
		return false, err
	}

	return len(records) > 0, nil
}

// sfAPIError is a single entry in a Salesforce API error response array.
// nolint:tagliatelle
type sfAPIError struct {
	Message   string   `json:"message"`
	ErrorCode string   `json:"errorCode"`
	Fields    []string `json:"fields"`
}

// isDuplicateDeveloperName reports whether err is a 400 response whose body
// contains the DUPLICATE_DEVELOPER_NAME Salesforce error code.
func isDuplicateDeveloperName(err error) bool {
	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusBadRequest {
		return false
	}

	var entries []sfAPIError
	if jsonErr := json.Unmarshal(httpErr.Body, &entries); jsonErr != nil {
		return false
	}

	for _, e := range entries {
		if e.ErrorCode == errCodeDuplicateDeveloperName {
			return true
		}
	}

	return false
}

// fullNameMismatch reports whether err is a 400 response indicating the supplied
// metadata FullName omitted the org's namespace prefix, and if so returns the
// expected (namespaced) full name.
//
// In a namespaced org, Salesforce stores created metadata as
// "<namespace>__<fullName>" and rejects a later reference to the un-namespaced
// name with a message like: "Full name amp_<id> does not match the full name
// speedboatdev__amp_<id> of the entity (id: ...)". Because the supplied full
// name is known, we match "<namespace>__<fullName>" directly rather than
// parsing the free-form message.
//
// A Salesforce namespace prefix is 1-15 alphanumeric characters, begins with a
// letter, and cannot contain two consecutive underscores (single underscores
// are allowed, e.g. "my_np"). The pattern below encodes that: a leading letter
// followed by units of an optional single underscore plus an alphanumeric, so
// underscores can never be consecutive or trailing — which also makes it stop
// cleanly at the "__" that separates the namespace from the full name.
// https://developer.salesforce.com/docs/atlas.en-us.pkg1_dev.meta/pkg1_dev/register_namespace_prefix.htm
func fullNameMismatch(err error, fullName string) (string, bool) {
	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusBadRequest || fullName == "" {
		return "", false
	}

	var entries []sfAPIError
	if jsonErr := json.Unmarshal(httpErr.Body, &entries); jsonErr != nil {
		return "", false
	}

	// Matches "<namespace>__<fullName>", e.g. "speedboatdev__amp_<id>" or
	// "my_np__amp_<id>".
	pattern, compileErr := regexp.Compile(`\b[A-Za-z](?:_?[A-Za-z0-9])*__` + regexp.QuoteMeta(fullName))
	if compileErr != nil {
		return "", false
	}

	for _, e := range entries {
		if match := pattern.FindString(e.Message); match != "" {
			return match, true
		}
	}

	return "", false
}

// recoverDuplicateByDeveloperName looks up a tooling entity by DeveloperName
// and returns the full record. Used to make Create* idempotent when Salesforce
// reports DUPLICATE_DEVELOPER_NAME — the SOQL query yields the existing Id,
// and the follow-up GET returns the same shape a successful Create would have.
//
// Salesforce does not allow FullName in a SOQL WHERE clause for these metadata
// objects, so callers must pass a DeveloperName. For most entities here that
// equals the FullName, but PlatformEventChannel strips its "__chn" suffix.
func recoverDuplicateByDeveloperName[T any](
	ctx context.Context, conn *Connector, objectType, developerName string,
) (*T, error) {
	logging.Logger(ctx).Info("create returned duplicate, fetching existing record",
		"objectType", objectType, "developerName", developerName)

	id, err := conn.findToolingEntityIDByDeveloperName(ctx, objectType, developerName)
	if err != nil {
		return nil, fmt.Errorf("%s duplicate detected, but SOQL lookup failed: %w",
			objectType, err)
	}

	existing, err := getToolingEntityByID[T](ctx, conn, objectType, id)
	if err != nil {
		return nil, fmt.Errorf("%s duplicate detected, found id=%s but GET failed: %w",
			objectType, id, err)
	}

	return existing, nil
}

// findToolingEntityIDByDeveloperName runs a Tooling API SOQL query to find
// the Id of the given object type by DeveloperName.
func (c *Connector) findToolingEntityIDByDeveloperName(
	ctx context.Context, objectType, developerName string,
) (string, error) {
	location, err := c.getRestApiURL("tooling/query")
	if err != nil {
		return "", err
	}

	soql := fmt.Sprintf("SELECT Id FROM %s WHERE DeveloperName = '%s'",
		objectType, escapeSOQLString(developerName))
	location.WithQueryParam("q", soql)

	resp, err := c.Client.Get(ctx, location.String())
	if err != nil {
		return "", err
	}

	body, ok := resp.Body()
	if !ok {
		return "", common.ErrEmptyJSONHTTPResponse
	}

	obj, err := body.GetObject()
	if err != nil {
		return "", err
	}

	recordsNode, exists := obj["records"]
	if !exists {
		return "", fmt.Errorf("%w: objectType=%s", errToolingQueryNoRecordsField, objectType)
	}

	records, err := recordsNode.GetArray()
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "", fmt.Errorf("%w: objectType=%s, developerName=%s", errToolingEntityNotFound, objectType, developerName)
	}

	rec, err := records[0].GetObject()
	if err != nil {
		return "", err
	}

	idNode, exists := rec["Id"]
	if !exists {
		return "", fmt.Errorf("%w: objectType=%s", errToolingRecordMissingID, objectType)
	}

	return idNode.MustString(), nil
}

// getToolingEntityByID fetches an entity by Id from the Tooling API and
// unmarshals into T.
func getToolingEntityByID[T any](
	ctx context.Context, conn *Connector, objectType, id string,
) (*T, error) {
	location, err := conn.getRestApiURL(fmt.Sprintf("tooling/sobjects/%s/%s", objectType, id))
	if err != nil {
		return nil, err
	}

	resp, err := conn.Client.Get(ctx, location.String())
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[T](resp)
}

// escapeSOQLString escapes a value for safe inclusion in a SOQL string
// literal. Backslashes must be escaped before single quotes.
func escapeSOQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)

	return s
}
