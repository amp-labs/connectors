// nolint:revive,godoclint
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

	// ErrCursorGone is returned when a cursor used for pagination is no longer valid.
	ErrCursorGone = errors.New("pagination cursor gone or expired")

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
	//  * GetResponse: An ampersand-style filter string that maps directly to GetResponse's
	//      bracket-notation query parameters. Supports both `query[...]` and `sort[...]`.
	//      Multiple filters can be separated by '&'.
	//      Examples:
	//          - "query[name]=campaign_name"
	//          - "query[isDefault]=true"
	//          - "sort[name]=ASC"
	//          - "sort[createdOn]=DESC"
	//          - "query[name]=test&sort[createdOn]=DESC"
	//      Reference: https://apireference.getresponse.com/#operation/getCampaignList
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

	// PageSize specifies the # of records to request when making a read request.
	PageSize int // optional
}

type WriteHeader struct {
	Key   string
	Value string
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

	Headers []WriteHeader // optional
}

func (p WriteParams) GetRecord() (Record, error) {
	return RecordDataToMap(p.RecordData)
}

// RecordDataToMap converts WriteParams.RecordData into a map[string]any.
//
// When possible use WriteParams.GetRecord instead.
//
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

// WriteResult represents the outcome of a single record write operation.
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

// DeleteResult represents the outcome of a single record delete operation.
type DeleteResult struct {
	// Success is true if deletion succeeded.
	Success bool `json:"success"`
}

// BatchStatus describes the aggregate outcome of a batch operation.
type BatchStatus string

const (
	BatchStatusSuccess BatchStatus = "success"
	BatchStatusFailure BatchStatus = "failure"
	BatchStatusPartial BatchStatus = "partial"
)

// BatchWriteType specifies the intended operation type within a batch modification.
type BatchWriteType string

const (
	BatchWriteTypeCreate BatchWriteType = "create"
	BatchWriteTypeUpdate BatchWriteType = "update"
)

// BatchWriteParam defines the input required to execute a batch write operation.
// It allows creating, updating, or upserting multiple records in a single request.
type BatchWriteParam struct {
	// ObjectName identifies the target object for the write operation.
	ObjectName ObjectName
	// Type defines how the records should be processed: create, update, or upsert.
	Type BatchWriteType
	// Batch contains the collection of record payloads to be written.
	Batch BatchItems
	// Headers contains additional headers to be added to the request.
	Headers []WriteHeader // optional
}

func TransformWriteHeaders(headers []WriteHeader, mode HeaderMode) []Header {
	transformedHeaders := []Header{}
	for _, header := range headers {
		transformedHeaders = append(transformedHeaders, Header{
			Key:   header.Key,
			Value: header.Value,
			Mode:  mode,
		})
	}

	return transformedHeaders
}

type BatchItem struct {
	Record       map[string]any
	Associations any
}

func (i BatchItem) GetRecord() (Record, error) {
	return RecordDataToMap(i.Record)
}

type BatchItems []BatchItem

func (p BatchWriteParam) IsCreate() bool {
	return p.Type == BatchWriteTypeCreate
}

func (p BatchWriteParam) IsUpdate() bool {
	return p.Type == BatchWriteTypeUpdate
}

type Record map[string]any

func (p BatchWriteParam) GetRecords() ([]Record, error) {
	return datautils.ForEachWithErr(p.Batch, func(batchItem BatchItem) (Record, error) {
		return RecordDataToMap(batchItem.Record)
	})
}

// BatchWriteResult represents the outcome of a provider batch write operation.
//
// It contains both a high-level summary of the batch and detailed per-record results.
//
// Providers may return more errors than there are payload items, or omit identifiers
// that would allow matching errors to specific records. In such cases, unmatched or
// batch-level issues are collected in the top-level Errors slice.
//
// Each identifiable record—matched by reference ID or record ID—produces a WriteResult
// entry in Results. If multiple identifiable errors occurred for the same record, they
// are grouped under WriteResult.Errors.
//
// Top-level Errors represent issues that apply to the batch as a whole or to records
// that could not be reliably matched to payload items.
type BatchWriteResult struct {
	// Status summarizes the batch outcome (success, failure, or partial).
	Status BatchStatus
	// Errors lists top-level errors that are not tied to specific records.
	// While errors that are specific to certain records are found in Results[i].Errors
	Errors []any
	// Results contains the detailed outcomes for each record in the batch.
	Results []WriteResult
	// SuccessCount is the number of successfully written records.
	SuccessCount int `json:"successCount"`
	// FailureCount is the number of failed records.
	FailureCount int `json:"failureCount"`
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
	// Some legacy connectors do not populate this, but only populates FieldsMap.
	Fields FieldsMetadata

	// Deprecated: for new connectors, please only populate and read `ObjectMetadata.Fields`.
	// FieldsMap is a map of field names to field display names.
	// TODO: Remove this field once all connectors populate Fields.
	FieldsMap map[string]string
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

// AddFieldMetadata updates Fields and FieldsMap fields ensuring data consistency.
func (m *ObjectMetadata) AddFieldMetadata(fieldName string, fieldMetadata FieldMetadata) {
	m.Fields[fieldName] = fieldMetadata
	m.FieldsMap[fieldName] = fieldMetadata.DisplayName
}

func (m *ObjectMetadata) RemoveFieldMetadata(fieldName string) {
	delete(m.Fields, fieldName)
	delete(m.FieldsMap, fieldName)
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
	ReadOnly *bool

	// IsCustom indicates whether the field is user-defined or custom.
	// True means the field was added by the user, false means it is native to the provider.
	IsCustom *bool

	// IsRequired indicates whether a value for the field is mandatory
	// when creating or updating the object.
	// True means the field must have a value, false means it is optional.
	IsRequired *bool

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

// CollapsedSubscriptionEvent some providers send multiple events in a single webhook payload.
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
	// RegistrationStatusPending registration is pending and not yet complete.
	RegistrationStatusPending RegistrationStatus = "pending"
	// RegistrationStatusFailed registration returned error, and all intermittent steps are rolled back.
	RegistrationStatusFailed RegistrationStatus = "failed"
	// RegistrationStatusSuccess successful registration.
	RegistrationStatusSuccess RegistrationStatus = "success"
	// RegistrationStatusFailedToRollback registration returned error, and failed to rollback some intermittent steps.
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

func (n ObjectName) String() string {
	return string(n)
}

type SubscribeParams struct {
	// Request contains provider-specific parameters that are unique to each subscription.
	// These parameters can be either optional or required depending on the provider's API requirements.
	// Each provider defines its own request structure (e.g., webhook URL, secret, unique reference, etc.)
	// that must be provided when creating subscriptions. The structure is provider-specific and should
	// match the provider's expected request format for subscription operations.
	Request any
	// RegistrationResult is the result of the Connector.Register call.
	// Connector.Subscribe requires information from the registration.
	// Not all providers require registration, so this is optional.
	// eg) Salesforce and HubSpot require registration because
	RegistrationResult *RegistrationResult
	// SubscriptionEvents is a normalized view representing the exact subscription state
	// that should be maintained in the provider's system. This field specifies which
	// objects and events the caller wants to subscribe to, using a provider-agnostic
	// format. Connector.Subscribe method will translate these normalized events to its own event
	// naming conventions when creating subscriptions. This represents the desired
	// state, which may differ from the actual state returned in SubscriptionResult.ObjectEvents
	// if some subscription operations fail or are rolled back.
	SubscriptionEvents map[ObjectName]ObjectEvents
}

type SubscriptionResult struct { // this corresponds to each API call.
	// Result contains the provider's original response data in its raw form.
	// This field preserves the complete, unmodified response from the provider's API,
	// allowing callers to access all original metadata, IDs, timestamps, and other
	// provider-specific information that may not be represented in the transformed ObjectEvents.
	// The structure of Result is provider-specific and should match the provider's actual API response format.
	Result any
	// ObjectEvents represents the transformed view of the subscription state after the operation.
	// This field contains a normalized, provider-agnostic representation of which objects and events
	// are currently subscribed in the provider's system. It reflects the exact state after
	// creating, updating, or deleting subscriptions, and may differ from the requested subscriptions
	// if some operations failed or were rolled back. This transformed view is useful for tracking
	// the actual subscription state without needing to parse provider-specific response formats.
	ObjectEvents map[ObjectName]ObjectEvents
	Status       SubscriptionStatus

	// Below fields are soon to be DEPRECATED, and will be removed in a future release.
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
	// SubscriptionStatusPending registration is pending and not yet complete.
	SubscriptionStatusPending SubscriptionStatus = "pending"
	// SubscriptionStatusFailed registration returned error, and all intermittent steps are rolled back.
	SubscriptionStatusFailed SubscriptionStatus = "failed"
	// SubscriptionStatusSuccess successful registration.
	SubscriptionStatusSuccess SubscriptionStatus = "success"
	// SubscriptionStatusFailedToRollback registration returned error, and failed to rollback some intermittent steps.
	SubscriptionStatusFailedToRollback SubscriptionStatus = "failed_to_rollback"
)
