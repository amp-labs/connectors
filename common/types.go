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
)

// ReadParams defines how we are reading data from a SaaS API.
type ReadParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string // required
	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields []string // required, at least one field needed
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage string // optional, only set this if you want to read the next page of results
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
	ObjectId string // optional

	// ObjectData is a JSON node representing the record of data we want to insert in the case of CREATE
	// or fields of data we want to modify in case of an update
	ObjectData map[string]interface{}
}

// NextPageToken is an opaque token that can be used to get the next page of results.
// Callers are encouraged to treat this as an opaque string, and not attempt to parse it.
// And although each provider will be different, callers should expect that this token
// will expire after some period of time. So long-term storage of this token is not recommended.
type NextPageToken string

// ReadResult is what's returned from reading data via the Read call.
type ReadResult struct {
	// Rows is the number of total rows in the result.
	Rows int64 `json:"rows"`
	// Data is a list of JSON nodes, where each node represents a record that we read.
	Data []map[string]interface{} `json:"data"`
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage NextPageToken `json:"nextPage,omitempty"`
	// Done is true if there are no more pages to read.
	Done bool `json:"done,omitempty"`
}

// WriteResult is what's returned from writing data via the Write call.
type WriteResult struct {
	// ObjectId is the ID of the written record.
	ObjectId string `json:"objectid,omitempty"`
	// Errors is list of error objects returned by the API.
	Errors []interface{} `json:"errors"`
	// Success is true if write succeeded.
	Success bool `json:"success,omitempty"`
	// Data is a JSON node containing data about the properties that were updated.
	Data map[string]interface{} `json:"data"` // optional
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
