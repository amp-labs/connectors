package common

import (
	"context"
	"errors"
	"fmt"
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

	// ErrMissingExpectedValues is returned when response data doesn't have values expected for processing.
	ErrMissingExpectedValues = errors.New("response data is missing expected values")

	// ErrEmptyJSONHTTPResponse is returned when the JSONHTTPResponse is nil.
	ErrEmptyJSONHTTPResponse = errors.New("empty json http response")

	// ErrEmptyRecordIdResponse is returned when the response body doesn't have record id.
	ErrEmptyRecordIdResponse = errors.New("empty record id in response body")

	// ErrRecordDataNotJSON is returned when the record data in WriteParams is not JSON.
	ErrRecordDataNotJSON = errors.New("record data is not JSON")

	// ErrOperationNotSupportedForObject is returned when operation is not supported for this object.
	ErrOperationNotSupportedForObject = errors.New("operation is not supported for this object in this module")

	// ErrResolvingURLPathForObject is returned when URL cannot be implied for object name.
	ErrResolvingURLPathForObject = errors.New("cannot resolve URL path for given object name")

	// ErrFailedToUnmarshalBody is returned when response body cannot be marshalled into some type.
	ErrFailedToUnmarshalBody = errors.New("failed to unmarshal response body")
)

// ReadParams defines how we are reading data from a SaaS API.
type ReadParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string // required

	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields datautils.StringSet // required, at least one field needed

	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage NextPageToken // optional, only set this if you want to read the next page of results

	// Since is a timestamp that can be used to get only records that have changed since that time.
	Since time.Time // optional, omit this to fetch all records

	// Deleted is true if we want to read deleted records instead of active records.
	Deleted bool // optional, defaults to false

	// Filter is supported for the following connectors:
	//	* Salesforce: it is a SOQL string that comes after the WHERE clause which will be used to filter the records.
	//		Reference: https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql.htm
	//	* Klaviyo: comma separated methods following JSON:API filtering syntax.
	//		Note: timing is already handled by Since argument.
	//		Reference: https://developers.klaviyo.com/en/docs/filtering_
	Filter string // optional
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
	// Raw is the raw JSON response from the provider.
	Raw map[string]any `json:"raw"`
	// RecordId is the ID of the record. Currently only populated for hubspot GetRecord and GetRecordsWithId function
	Id string `json:"id,omitempty"`
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

// NewHTTPStatusError creates a new error with the given HTTP status.
func NewHTTPStatusError(status int, err error) error {
	if status < 1 || status > 599 {
		return err
	}

	return &HTTPStatusError{
		HTTPStatus: status,
		err:        err,
	}
}

// HTTPStatusError is an error that contains an HTTP status code.
type HTTPStatusError struct {
	// HTTPStatus is the original HTTP status.
	HTTPStatus int

	// The underlying error
	err error
}

func (r HTTPStatusError) Error() string {
	if r.HTTPStatus > 0 {
		return fmt.Sprintf("HTTP status %d: %v", r.HTTPStatus, r.err)
	}

	return r.err.Error()
}

func (r HTTPStatusError) Unwrap() error {
	return r.err
}

type ListObjectMetadataResult struct {
	// Result is a map of object names to object metadata
	Result map[string]ObjectMetadata

	// Errors is a map of object names to errors
	Errors map[string]error
}

type ObjectMetadata struct {
	// Provider's display name for the object
	DisplayName string

	// FieldsMap is a map of field names to field display names
	FieldsMap map[string]string
}

type PostAuthInfo struct {
	CatalogVars *map[string]string
	RawResponse *JSONHTTPResponse
}

type WebhookEventType string

const (
	WebhookEventTypeCreate WebhookEventType = "create"
	WebhookEventTypeUpdate WebhookEventType = "update"
	WebhookEventTypeDelete WebhookEventType = "delete"
	WebhookEventTypeOther  WebhookEventType = "other"
)
