package common

import (
	"fmt"
)

// We map common errors to modes with known resolution strategies.
type ErrorMode string

const (
	// Token expired or invalid. Refresh the token and retry.
	AccessTokenInvalid ErrorMode = "ACCESS_TOKEN_INVALID"
	// Customer didn't enable this API on their SaaS instance.
	ApiDisabled ErrorMode = "API_DISABLED"
	// Temporary error. Can retry.
	RetryableError ErrorMode = "RETRYABLE_ERROR"
	// Other API-related errors.
	OtherError ErrorMode = "OTHER_API_ERROR"
	// Non-API errors, the request didn't make it to the API server.
	NonApiError ErrorMode = "NON_API_ERROR"
)

// ReadConfig defines how we are reading data from a SaaS API.
type ReadConfig struct {
	// The name of the object we are reading, e.g. "Account"
	ObjectName string
	// The fields we are reading from the object, e.g. ["Id", "Name", "BillingCity"]
	Fields []string
}

// GetCallConfig defines the parameters for a generic GET call to a SaaS API.
type GetCallConfig struct {
	// The endpoint to call, e.g. "sobjects/Account/describe"
	Endpoint string
}

// Result from reading data.
type ReadResult struct {
	// Rows is the number of total rows in the result.
	Rows int
	// Data is a list of maps, where each map represents a record that we read.
	Data []map[string]interface{}
}

// Result from a generic API call.
type GenericResult struct {
	Data map[string]interface{}
}

type ErrorWithStatus struct {
	// The error mode that we've mapped this error to.
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
