package common

import (
	"errors"
	"fmt"
	"time"
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

	// ErrMissingOauthConfig is returned when the OAuth config is missing.
	ErrMissingOauthConfig = errors.New("missing OAuth config")

	// ErrMissingRefreshToken is returned when the refresh token is missing.
	ErrMissingRefreshToken = errors.New("missing refresh token")

	// ErrEmptyBaseURL is returned when the URL is relative, and the base URL is empty.
	ErrEmptyBaseURL = errors.New("empty base URL")

	// ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrMissingObjects is returned when no objects are provided in the request.
	ErrMissingObjects = errors.New("no objects provided")

	// ErrMissingRecordID is returned when resource id is missing in the request.
	ErrMissingRecordID = errors.New("no object ID provided")

	// ErrInvalidPathJoin is returned when the path join is invalid.
	ErrInvalidPathJoin = errors.New("invalid path join")

	// ErrReadFile is returned when the path is invalid.
	ErrReadFile = errors.New("failed to read file")

	// ErrRequestFailed is returned when the request failed.
	ErrRequestFailed = errors.New("request failed")

	// ErrParseError is returned data parsing failed.
	ErrParseError = errors.New("parse error")

	// ErrBadRequest is returned when we get a 400 response from the provider.
	ErrBadRequest = errors.New("bad request")
)

// ReadParams defines how we are reading data from a SaaS API.
type ReadParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string // required
	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields []string // required, at least one field needed
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage NextPageToken // optional, only set this if you want to read the next page of results
	// Since is a timestamp that can be used to get only records that have changed since that time.
	Since time.Time // optional, omit this to fetch all records
	// Deleted is true if we want to read deleted records instead of active records.
	Deleted bool // optional, defaults to false
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
	Fields map[string]interface{} `json:"fields"`
	// Raw is the raw JSON response from the provider.
	Raw map[string]interface{} `json:"raw"`
}

// WriteResult is what's returned from writing data via the Write call.
type WriteResult struct {
	// Success is true if write succeeded.
	Success bool `json:"success"`
	// RecordId is the ID of the written record.
	RecordId string `json:"recordId,omitempty"` // optional
	// Errors is list of error record returned by the API.
	Errors []interface{} `json:"errors,omitempty"` // optional
	// Data is a JSON node containing data about the properties that were updated.
	Data map[string]interface{} `json:"data,omitempty"` // optional
}

// DeleteResult is what's returned from deleting data via the Delete call.
type DeleteResult struct {
	// Success is true if deletion succeeded.
	Success bool `json:"success"`
}

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
