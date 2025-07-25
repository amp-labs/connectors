package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	// ErrAccessToken is a token which isn't valid.
	ErrAccessToken = errors.New("access token invalid")

	// ErrApiDisabled means a customer didn't enable this API on their SaaS instance.
	ErrApiDisabled = errors.New("API disabled")

	// ErrForbidden means the user doesn't have access to this resource.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidSessionId means the session ID is invalid.
	ErrInvalidSessionId = errors.New("invalid session id")

	// ErrUnableToLockRow means the resource couldn't be locked.
	ErrUnableToLockRow = errors.New("unable to lock row")

	// ErrInvalidGrant means the OAuth grant is invalid.
	ErrInvalidGrant = errors.New("invalid grant")

	// ErrLimitExceeded means a quota limit was exceeded.
	ErrLimitExceeded = errors.New("request limit exceeded")

	// ErrRetryable represents a temporary error. Can retry.
	ErrRetryable = errors.New("retryable error")

	// ErrCaller represents non-retryable errors caused by bad input from the caller.
	ErrCaller = errors.New("caller error")

	// ErrServer represents non-retryable errors caused by something on the server.
	ErrServer = errors.New("server error")

	// ErrUnknown represents an unknown status code response.
	ErrUnknown = errors.New("unknown error")

	// ErrNotJSON is returned when a response is not JSON.
	ErrNotJSON = errors.New("response is not JSON")

	// ErrNotXML is returned when a response is not XML.
	ErrNotXML = errors.New("response is not XML")

	// ErrMissingOauthConfig is returned when the OAuth config is missing.
	ErrMissingOauthConfig = errors.New("missing OAuth config")

	// ErrMissingRefreshToken is returned when the refresh token is missing.
	ErrMissingRefreshToken = errors.New("missing refresh token")

	// ErrEmptyBaseURL is returned when the URL is relative, and the base URL is empty.
	ErrEmptyBaseURL = errors.New("empty base URL")

	// ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrRequestFailed is returned when the request failed.
	ErrRequestFailed = errors.New("request failed")

	// ErrParseError is returned data parsing failed.
	ErrParseError = errors.New("parse error")

	// ErrBadRequest is returned when we get a 400 response from the provider.
	ErrBadRequest = errors.New("bad request")

	// ErrNotFound is returned when we get a 404 response from the provider.
	ErrNotFound = errors.New("not found")

	// ErrMissingExpectedValues is returned when response data doesn't have values expected for processing.
	ErrMissingExpectedValues = errors.New("response data is missing expected values")

	// ErrPreprocessingWritePayload is returned when request payload supplied to connector "Write" method
	// couldn't be processed. Likely, the issue is within provided WriteParams.RecordData or implementation.
	ErrPreprocessingWritePayload = errors.New("failed preprocessing write payload")

	// ErrEmptyJSONHTTPResponse is returned when the JSONHTTPResponse is nil.
	ErrEmptyJSONHTTPResponse = errors.New("empty json http response")

	// ErrEmptyRecordIdResponse is returned when the response body doesn't have record id.
	ErrEmptyRecordIdResponse = errors.New("empty record id in response body")

	// ErrRecordDataNotJSON is returned when the record data in WriteParams is not JSON.
	ErrRecordDataNotJSON = errors.New("record data is not JSON")

	// ErrOperationNotSupportedForObject is returned when operation is not supported for this object.
	ErrOperationNotSupportedForObject = errors.New("operation is not supported for this object in this module")

	// ErrObjectNotSupported is returned when operation is not supported for this object.
	ErrObjectNotSupported = errors.New("operation is not supported for this object")

	// ErrResolvingURLPathForObject is returned when URL cannot be implied for object name.
	ErrResolvingURLPathForObject = errors.New("cannot resolve URL path for given object name")

	// ErrFailedToUnmarshalBody is returned when response body cannot be marshalled into some type.
	ErrFailedToUnmarshalBody = errors.New("failed to unmarshal response body")

	// ErrNextPageInvalid is returned when next page token provided in Read operation cannot be understood.
	ErrNextPageInvalid = errors.New("next page token is invalid")

	// ErrInvalidImplementation is returned when implementation assumption is broken.
	// This is not a client issue.
	ErrInvalidImplementation = errors.New("invalid implementation")

	// ErrPayloadNotURLForm is returned when payload is not string key-value pair
	// which could be encoded for POST with content type of application/x-www-form-urlencoded.
	ErrPayloadNotURLForm = errors.New("payload cannot be url-form encoded")

	// ErrResolvingCustomFields is returned when custom fields cannot be retrieved for Read or ListObjectMetadata.
	ErrResolvingCustomFields = errors.New("cannot resolve custom fields")

	ErrGetRecordNotSupportedForObject = errors.New("getRecord is not supported for the object")

	// ErrImplementation is returned when the code takes an unexpected or logically invalid execution path.
	// It should be used to explicitly catch cases that would otherwise lead to panics (e.g., nil pointer dereference).
	// This typically indicates a broken assumption or inconsistency in the implementation logic.
	ErrImplementation = errors.New("code took invalid execution path")
)

// ReadParams defines how we are reading data from a SaaS API.
type ReadParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string // required

	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields datautils.StringSet // required, at least one field needed

	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage NextPageToken // optional, only set this if you want to read the next page of results

	// Since is an optional timestamp to fetch only records updated **after** this time.
	// Used for incremental reads.
	Since time.Time

	// Until is an optional timestamp to fetch only records updated **up to and including** this time.
	// Pagination stops when records exceed this timestamp.
	Until time.Time

	// Deleted is true if we want to read deleted records instead of active records.
	Deleted bool // optional, defaults to false

	// Filter defines the filtering criteria for supported connectors.
	// It is optional and behaves differently depending on the connector:
	//	* Salesforce: It is a SOQL string that comes after the WHERE clause which will be used to filter the records.
	//		Reference: https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql.htm
	//	* Klaviyo: Comma separated methods following JSON:API filtering syntax.
	//		Note: timing is already handled by Since argument.
	//		Reference: https://developers.klaviyo.com/en/docs/filtering_
	//	* Marketo: Comma-separated activityTypeIds for filtering lead activities.
	//		Note: Only supported when reading Lead Activities (not other endpoints).
	//		Example: "1,6,12" (for visitWebpage, fillOutForm, emailClicked)
	//		Reference: https://developer.adobe.com/marketo-apis/api/mapi/#tag/Activities
	Filter string // optional

	// AssociatedObjects specifies a list of related objects to fetch along with the main object.
	// It is optional and supported by the following connectors:
	//	* HubSpot: Supported in Read operation, but not Search.
	//	* Stripe: Only nested objects can be expanded. Specify a dot-separated path
	//		to the property to fetch and expand those objects.
	//		Reference: https://docs.stripe.com/expand#how-it-works
	//	* Capsule: Embeds objects in response.
	//		Reference: https://developer.capsulecrm.com/v2/overview/reading-from-the-api
	AssociatedObjects []string // optional
}

// WriteParams defines how we are writing data to a SaaS API.
type WriteParams struct {
	// The name of the object we are writing, e.g. "Account"
	ObjectName string // required

	// The external ID of the object instance we are updating. Provided in the case of UPDATE, but not CREATE.
	RecordId string // optional

	// RecordData is a JSON node representing the record of data we want to insert in the case of CREATE
	// or fields of data we want to modify in case of an update
	RecordData any // required

	// Associations contains associations between the object and other objects.
	Associations any // optional
}

// RecordDataToMap converts WriteParams.RecordData into a map[string]any.
// If RecordData is already a map, it is returned directly.
// Otherwise, it is serialized to JSON and then deserialized back into a map.
func RecordDataToMap(recordData any) (map[string]any, error) {
	if object, ok := recordData.(map[string]any); ok {
		return object, nil
	}

	bytes, err := json.Marshal(recordData)
	if err != nil {
		return nil, err
	}

	object := make(map[string]any)
	if err = json.Unmarshal(bytes, &object); err != nil {
		return nil, err
	}

	return object, nil
}

// DeleteParams defines how we are deleting data in SaaS API.
type DeleteParams struct {
	// The name of the object we are deleting, e.g. "Account"
	ObjectName string // required

	// The external ID of the object instance we are removing.
	RecordId string // required
}

// NextPageToken is an opaque token that can be used to get the next page of results.
// Callers are encouraged to treat this as an opaque string, and not attempt to parse it.
// And although each provider will be different, callers should expect that this token
// will expire after some period of time. So long-term storage of this token is not recommended.
type NextPageToken string

func (t NextPageToken) String() string {
	return string(t)
}

// ReadResult is what's returned from reading data via the Read call.
type ReadResult struct {
	// Rows is the number of total rows in the result.
	Rows int64 `json:"rows"`
	// Data is an array where each element represents a ReadResultRow.
	Data []ReadResultRow `json:"data"`
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage NextPageToken `json:"nextPage,omitempty"`
	// Done is true if there are no more pages to read.
	Done bool `json:"done,omitempty"`
}

// ReadResultRow is a single row of data returned from a Read call, which contains
// the requested fields, as well as the raw JSON response from the provider.
type ReadResultRow struct {
	// Fields is a map of requested provider field names to values.
	// All field names are in lowercase (eg: accountid, name, billingcityid)
	Fields map[string]any `json:"fields"`
	// Associations is a map of associated objects to the main object.
	// The key is the associated object name, and the value is an array of associated object ids.
	Associations map[string][]Association `json:"associations,omitempty"`
	// Raw is the raw JSON response from the provider.
	Raw map[string]any `json:"raw"`
	// RecordId is the ID of the record. Currently only populated for hubspot GetRecord and GetRecordsWithId function
	Id string `json:"id,omitempty"`
}

// Association is a struct that represents an association between two objects.
// If you think of an association as a directed edge between two nodes, then
// the ObjectID is the target node, and the AssociationType is the type of edge.
// The source node is represented by ReadResultRow.
type Association struct {
	// ObjectID is the ID of the associated object.
	ObjectId string `json:"objectId"`
	// AssociationType is the type of association.
	AssociationType string         `json:"associationType,omitempty"`
	Raw             map[string]any `json:"raw,omitempty"`
}

// WriteResult is what's returned from writing data via the Write call.
type WriteResult struct {
	// Success is true if write succeeded.
	Success bool `json:"success"`
	// RecordId is the ID of the written record.
	RecordId string `json:"recordId,omitempty"` // optional
	// Errors is list of error record returned by the API.
	Errors []any `json:"errors,omitempty"` // optional
	// Data is a JSON node containing data about the properties that were updated.
	Data map[string]any `json:"data,omitempty"` // optional
}

// DeleteResult is what's returned from deleting data via the Delete call.
type DeleteResult struct {
	// Success is true if deletion succeeded.
	Success bool `json:"success"`
}

// WriteMethod is signature for any HTTP method that performs write modifications.
// Ex: Post/Put/Patch.
type WriteMethod func(context.Context, string, any, ...Header) (*JSONHTTPResponse, error)

// NewHTTPError creates a new error with the given HTTP status.
func NewHTTPError(status int, body []byte, headers Headers, err error) error {
	if status < 1 || status > 599 {
		return err
	}

	// Just in case the caller mutates the body after passing it in,
	// we make a copy of the body to ensure that the error contains
	// the original body.
	var bodyCopy []byte

	if body != nil {
		bodyCopy = make([]byte, len(body))
		copy(bodyCopy, body)
	}

	return &HTTPError{
		Status:  status,
		Headers: headers,
		Body:    bodyCopy,
		err:     err,
	}
}

// HTTPError is an error that contains both an error and details
// about the HTTP response that caused the error. It includes
// the HTTP status code, headers, and body of the response.
// Body and Headers are optional and may be nil if not available.
type HTTPError struct {
	// Status is the original HTTP status.
	Status int

	// Headers are the HTTP headers of the response, if available.
	Headers Headers // optional

	// Body is the raw response body, if available.
	Body []byte // optional

	// The underlying error
	err error
}

func (r HTTPError) Error() string {
	if r.Status > 0 {
		return fmt.Sprintf("HTTP status %d: %v", r.Status, r.err)
	}

	return r.err.Error()
}

func (r HTTPError) Unwrap() error {
	return r.err
}

type ListObjectMetadataResult struct {
	// Result is a map of object names to object metadata
	Result map[string]ObjectMetadata

	// Errors is a map of object names to errors
	Errors map[string]error
}

func NewListObjectMetadataResult() *ListObjectMetadataResult {
	return &ListObjectMetadataResult{
		Result: make(map[string]ObjectMetadata),
		Errors: make(map[string]error),
	}
}

// AppendError will associate an error with the object.
// It is possible that single object may have multiple errors.
func (r ListObjectMetadataResult) AppendError(objectName string, err error) {
	r.Errors[objectName] = errors.Join(r.Errors[objectName], err)
}

type ObjectMetadata struct {
	// Provider's display name for the object.
	DisplayName string

	// Fields is a map of field names to FieldMetadata.
	Fields FieldsMetadata

	// FieldsMap is a map of field names to field display names.
	// Deprecated: this map includes only display names.
	// Refer to Fields for extended description of field properties.
	FieldsMap map[string]string
}

// AddFieldMetadata updates Fields and FieldsMap fields ensuring data consistency.
func (m *ObjectMetadata) AddFieldMetadata(fieldName string, fieldMetadata FieldMetadata) {
	m.Fields[fieldName] = fieldMetadata
	m.FieldsMap[fieldName] = fieldMetadata.DisplayName
}

// NewObjectMetadata constructs ObjectMetadata.
// This will automatically infer fields map from field metadata map. This construct exists for such convenience.
func NewObjectMetadata(displayName string, fields FieldsMetadata) *ObjectMetadata {
	return &ObjectMetadata{
		DisplayName: displayName,
		Fields:      fields,
		FieldsMap:   inferDeprecatedFieldsMap(fields),
	}
}

type FieldMetadata struct {
	// DisplayName is a human-readable field name.
	DisplayName string

	// ValueType is a set of Ampersand defined field types.
	ValueType ValueType

	// ProviderType is the raw type, a term used by provider API.
	// Each is mapped to an Ampersand ValueType.
	ProviderType string

	// ReadOnly would indicate if field can be modified or only read.
	ReadOnly bool

	// Values is a list of possible values for this field.
	// It is applicable only if the type is either singleSelect or multiSelect, otherwise slice is nil.
	Values []FieldValue
}

type FieldsMetadata map[string]FieldMetadata

func (f FieldsMetadata) AddFieldWithDisplayOnly(fieldName string, displayName string) {
	f[fieldName] = FieldMetadata{
		DisplayName:  displayName,
		ValueType:    "",
		ProviderType: "",
		ReadOnly:     false,
		Values:       nil,
	}
}

type FieldValue struct {
	Value        string
	DisplayValue string
}

type FieldValues []FieldValue

type PostAuthInfo struct {
	CatalogVars          *map[string]string
	RawResponse          *JSONHTTPResponse
	ProviderWorkspaceRef string
}

type SubscriptionEventType string

const (
	SubscriptionEventTypeCreate            SubscriptionEventType = "create"
	SubscriptionEventTypeUpdate            SubscriptionEventType = "update"
	SubscriptionEventTypeDelete            SubscriptionEventType = "delete"
	SubscriptionEventTypeAssociationUpdate SubscriptionEventType = "associationUpdate"
	SubscriptionEventTypeOther             SubscriptionEventType = "other"
)

// SubscriptionEvent is an interface for webhook events coming from the provider.
// This interface defines methods to extract information from the webhook event.
type SubscriptionEvent interface {
	EventType() (SubscriptionEventType, error)
	RawEventName() (string, error)
	ObjectName() (string, error)
	Workspace() (string, error)
	RecordId() (string, error)
	EventTimeStampNano() (int64, error)
	RawMap() (map[string]any, error)
}

type SubscriptionUpdateEvent interface {
	SubscriptionEvent
	// GetUpdatedFields returns the fields that were updated in the event.
	UpdatedFields() ([]string, error)
}

// Some providers send multiple events in a single webhook payload.
// This interface is used to extract individual events to SubscriptionEvent type
// from a collapsed event for webhook parsing and processing.
type CollapsedSubscriptionEvent interface {
	SubscriptionEventList() ([]SubscriptionEvent, error)
	RawMap() (map[string]any, error)
}

// WebhookRequest is a struct that contains the request parameters for a webhook.
type WebhookRequest struct {
	Headers http.Header
	Body    []byte
	URL     string
	Method  string
}

// VerificationParams is a struct that contains the parameters specific to the provider.
type VerificationParams struct {
	Param any
}

func inferDeprecatedFieldsMap(fields FieldsMetadata) map[string]string {
	fieldsMap := make(map[string]string)

	for name, field := range fields {
		fieldsMap[name] = field.DisplayName
	}

	return fieldsMap
}

type RegistrationResult struct {
	RegistrationRef string
	Result          any // struct depends on the provider
	Status          RegistrationStatus
}

type RegistrationStatus string

const (
	// registration is pending and not yet complete.
	RegistrationStatusPending RegistrationStatus = "pending"
	// registration returned error, and all intermittent steps are rolled back.
	RegistrationStatusFailed RegistrationStatus = "failed"
	// successful registration.
	RegistrationStatusSuccess RegistrationStatus = "success"
	// registration returned error, and failed to rollback some intermittent steps.
	RegistrationStatusFailedToRollback RegistrationStatus = "failed_to_rollback"
)

type SubscriptionRegistrationParams struct {
	Request any `json:"request" validate:"required"`
}

type ObjectEvents struct {
	// ["create", "update", "delete"] our regular CRUD operation events
	// we translate to provider-specific names contact.creation
	Events []SubscriptionEventType
	// ["email", "fax"] fields to watch for an update subscription
	WatchFields []string
	// true if all fields should be watched for an update subscription
	// this is provider specific, and not all providers support this.
	WatchFieldsAll bool
	// any non CRUD operations with provider specific event names
	// eg)  ["contact.merged"] for hubspot or ["jira_issue:restored", "jira_issue:archived"] for jira.
	PassThroughEvents []string
}

type ObjectName string

type SubscribeParams struct {
	Request any
	// RegistrationResult is the result of the Connector.Register call.
	// Connector.Subscribe requires information from the registration.
	// Not all providers require registration, so this is optional.
	// eg) Salesforce and HubSpot require registration because
	RegistrationResult *RegistrationResult
	SubscriptionEvents map[ObjectName]ObjectEvents
}

type SubscriptionResult struct { // this corresponds to each API call.
	Result       any
	ObjectEvents map[ObjectName]ObjectEvents
	Status       SubscriptionStatus

	// Below fields are deprecated, and will be removed in a future release.
	// Use ObjectEvents instead.

	Objects []ObjectName
	Events  []SubscriptionEventType
	// ["create", "update", "delete"]
	// our regular CRUD operation events we translate to provider-specific names contact.creation
	UpdateFields []string
	// ["email", "fax"]
	PassThroughEvents []string
	// provider specific events ["contact.merged"] for hubspot or ["jira_issue:restored", "jira_issue:archived"] for jira.
}

type SubscriptionStatus string

const (
	// registration is pending and not yet complete.
	SubscriptionStatusPending SubscriptionStatus = "pending"
	// registration returned error, and all intermittent steps are rolled back.
	SubscriptionStatusFailed SubscriptionStatus = "failed"
	// successful registration.
	SubscriptionStatusSuccess SubscriptionStatus = "success"
	// registration returned error, and failed to rollback some intermittent steps.
	SubscriptionStatusFailedToRollback SubscriptionStatus = "failed_to_rollback"
)
