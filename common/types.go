package common

import (
	"fmt"
)

type ErrorMode string

const (
	// Token expired or invalid. Refresh the token and retry.
	AccessTokenInvalid ErrorMode = "ACCESS_TOKEN_INVALID"
	// Customer didn't enable this API on their SaaS instance.
	ApiDisabled ErrorMode = "API_DISABLED"
	// Temporary error. Retry.
	RetryableError ErrorMode = "RETRYABLE_ERROR"
	// Other API-related errors.
	OtherError ErrorMode = "OTHER_API_ERROR"
	// Non-API errors, the request didn't make it to the API server.
	NonApiError ErrorMode = "NON_API_ERROR"
)

// ReadConfig defines what we are reading and provides the necessary credentials.
type ReadConfig struct {
	ObjectName string
	Fields [] string
}

type GetCallConfig struct {
	// The endpoint to call, e.g. "sobjects/Account/describe"
	Endpoint string
	// Optional. Which fields from the API response should be returned. 
	// If not provided, we will return all fields in the response.
	Fields [] string
}

type GenericResult struct {
	Data map [string] interface {}
}

type ReadResult struct {
	// Rows is the number of total rows in the result.
	Rows int
	// Data is a list of maps, where each map represents a record that we read.
	Data [] map [string] interface {}
}

type ErrorWithStatus struct {
	// We map common errors to modes with known resolution strategies.
	Mode ErrorMode
	// HttpStatus is the original HTTP status.
	HttpStatus int
	// A human-readable error message.
	Message string
}

func (r ErrorWithStatus) Error() string {
	if r.HttpStatus > 0 {
		return fmt.Sprintf("[%v] Status %d: %v", r.Mode, r.HttpStatus, r.Message)
	}
	return fmt.Sprintf("[%v] %v", r.Mode, r.Message)
}
