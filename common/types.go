package common

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrAccessToken is a token which isn't valid.
	ErrAccessToken = errors.New("access token invalid")

	// ErrApiDisabled means a customer didn't enable this API on their SaaS instance.
	ErrApiDisabled = errors.New("API disabled")

	// ErrRetryable represents a temporary error. Can retry.
	ErrRetryable = errors.New("retryable error")

	// ErrCaller represents non-retryable errors caused by bad input from the caller.
	ErrCaller = errors.New("caller error")

	// ErrServer represents non-retryable errors caused by something on the server.
	ErrServer = errors.New("server error")

	// ErrUnknown represents an unknown status code response.
	ErrUnknown = errors.New("unknown error")
)

// ReadParams defines how we are reading data from a SaaS API.
type ReadParams struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string
	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields []string
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage string
	// Since is a timestamp that can be used to get only records that have changed since that time.
	Since time.Time
}

// Result from reading data.
type ReadResult struct {
	// Rows is the number of total rows in the result.
	Rows int64 `json:"rows"`
	// Data is a list of JSON nodes, where each node represents a record that we read.
	Data []map[string]interface{} `json:"data"`
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage string `json:"nextPage,omitempty"`
	// Done is true if there are no more pages to read.
	Done bool `json:"done,omitempty"`
}

// TokenProvider is a function that returns a token. The precise type of the
// token is up to the API.
type TokenProvider[Token any] func(ctx context.Context) (Token, error)

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

	return fmt.Sprintf("%v", r.err)
}

func (r HTTPStatusError) Unwrap() error {
	return r.err
}
