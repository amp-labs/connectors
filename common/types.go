package common

import (
	"errors"
	"fmt"
	"time"
)

var (
	// AccessTokenInvalid is a token which isn't valid.
	AccessTokenInvalid = errors.New("access token invalid")

	// ApiDisabled means a customer didn't enable this API on their SaaS instance.
	ApiDisabled = errors.New("API disabled")

	// RetryableError represents a temporary error. Can retry.
	RetryableError = errors.New("retryable error")

	// CallerError represents non-retryable errors caused by bad input from the caller.
	CallerError = errors.New("caller error")

	// ServerError represents non-retryable errors caused by something on the server.
	ServerError = errors.New("server error")
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
	Rows int
	// Data is a list of JSON nodes, where each node represents a record that we read.
	Data []map[string]interface{}
	// NextPage is an opaque token that can be used to get the next page of results.
	NextPage string
	// Done is true if there are no more pages to read.
	Done bool
}

func NewErrorWithStatus(status int, err error) error {
	if status < 1 || status > 599 {
		return err
	}
	return &ErrorWithStatus{
		HttpStatus: status,
		err:        err,
	}
}

type ErrorWithStatus struct {
	// HttpStatus is the original HTTP status.
	HttpStatus int

	// The underlying error
	err error
}

func (r ErrorWithStatus) Error() string {
	if r.HttpStatus > 0 {
		return fmt.Sprintf("HTTP status %d: %v", r.HttpStatus, r.err)
	}
	return fmt.Sprintf("%v", r.err)
}

func (r ErrorWithStatus) Unwrap() error {
	return r.err
}
