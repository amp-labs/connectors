package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrNotArray           = errors.New("records is not an array")
	ErrNotObject          = errors.New("record isn't an object")
	ErrNoFields           = errors.New("no fields specified")
	ErrNotString          = errors.New("nextRecordsUrl isn't a string")
	ErrNotBool            = errors.New("done isn't a boolean")
	ErrNotNumeric         = errors.New("totalSize isn't numeric")
	ErrMissingSubdomain   = errors.New("missing Salesforce workspace name")
	ErrMissingClient      = errors.New("JSON http client not set")
	ErrCannotReadMetadata = errors.New("cannot read object metadata, it is possible you don't have the correct permissions set")
)

type jsonError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	if res.Header.Get("Content-Type") == "application/json" {
		return c.interpretJSONError(res, body)
	}

	return common.InterpretError(res, body)
}

func createError(baseErr error, sfErr jsonError) error {
	if len(sfErr.Message) > 0 {
		return fmt.Errorf("%w: %s", baseErr, sfErr.Message)
	}

	return baseErr
}

func (c *Connector) interpretJSONError(res *http.Response, body []byte) error {
	errs := []jsonError{}
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	for _, sfErr := range errs {
		switch sfErr.ErrorCode {
		case "INVALID_SESSION_ID":
			return createError(common.ErrInvalidSessionId, sfErr)
		case "INSUFFICIENT_ACCESS_OR_READONLY":
			return createError(common.ErrForbidden, sfErr)
		case "API_DISABLED_FOR_ORG":
			return createError(common.ErrApiDisabled, sfErr)
		case "UNABLE_TO_LOCK_ROW":
			return createError(common.ErrUnableToLockRow, sfErr)
		case "INVALID_GRANT":
			return createError(common.ErrInvalidGrant, sfErr)
		case "REQUEST_LIMIT_EXCEEDED":
			return createError(common.ErrLimitExceeded, sfErr)
		default:
			continue
		}
	}

	// No known errors, just do the normal error handling logic
	return common.InterpretError(res, body)
}
